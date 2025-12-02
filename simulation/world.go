package simulation

import (
	"math/rand"
	"sync"
)

// CellType represents the type of entity in a cell
type CellType int

const (
	Empty CellType = iota
	Fish
	Shark
)

// Cell represents a single cell in the grid
type Cell struct {
	Type      CellType
	Energy    int
	BreedTime int
}

// World represents the Wa-Tor world
type World struct {
	Width       int
	Height      int
	Grid        [][]Cell
	FishBreed   int
	SharkBreed  int
	SharkStarve int
}

// NewWorld creates a new Wa-Tor world
func NewWorld(width, height, numFish, numShark, fishBreed, sharkBreed, sharkStarve int) *World {
	w := &World{
		Width:       width,
		Height:      height,
		Grid:        make([][]Cell, height),
		FishBreed:   fishBreed,
		SharkBreed:  sharkBreed,
		SharkStarve: sharkStarve,
	}

	// Initialize empty grid
	for i := range height {
		w.Grid[i] = make([]Cell, width)
	}

	// Place fish randomly
	for range numFish {
		for {
			x := rand.Intn(width)
			y := rand.Intn(height)
			if w.Grid[y][x].Type == Empty {
				w.Grid[y][x] = Cell{
					Type:      Fish,
					BreedTime: rand.Intn(fishBreed),
				}
				break
			}
		}
	}

	// Place sharks randomly
	for range numShark {
		for {
			x := rand.Intn(width)
			y := rand.Intn(height)
			if w.Grid[y][x].Type == Empty {
				w.Grid[y][x] = Cell{
					Type:      Shark,
					Energy:    sharkStarve,
					BreedTime: rand.Intn(sharkBreed),
				}
				break
			}
		}
	}

	return w
}

// Count returns the number of fish and sharks
func (w *World) Count() (int, int) {
	fish, sharks := 0, 0
	for i := 0; i < w.Height; i++ {
		for j := 0; j < w.Width; j++ {
			switch w.Grid[i][j].Type {
			case Fish:
				fish++
			case Shark:
				sharks++
			}
		}
	}
	return fish, sharks
}

// Step performs one simulation step
func (w *World) Step(threads int) int {
	newGrid := make([][]Cell, w.Height)
	for i := 0; i < w.Height; i++ {
		newGrid[i] = make([]Cell, w.Width)
	}

	moved := make([][]bool, w.Height)
	for i := 0; i < w.Height; i++ {
		moved[i] = make([]bool, w.Width)
	}

	var fishEaten int
	if threads == 1 {
		fishEaten = w.stepSingle(newGrid, moved)
	} else {
		fishEaten = w.stepParallel(newGrid, moved, threads)
	}

	w.Grid = newGrid
	return fishEaten
}

func (w *World) stepSingle(newGrid [][]Cell, moved [][]bool) int {
	fishEaten := 0

	// Process sharks first
	for i := 0; i < w.Height; i++ {
		for j := 0; j < w.Width; j++ {
			if w.Grid[i][j].Type == Shark && !moved[i][j] {
				eaten := w.moveShark(i, j, newGrid, moved)
				if eaten {
					fishEaten++
				}
			}
		}
	}

	// Then process fish
	for i := 0; i < w.Height; i++ {
		for j := 0; j < w.Width; j++ {
			if w.Grid[i][j].Type == Fish && !moved[i][j] {
				w.moveFish(i, j, newGrid, moved)
			}
		}
	}

	return fishEaten
}

func (w *World) stepParallel(newGrid [][]Cell, moved [][]bool, threads int) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	fishEaten := 0

	rowsPerThread := w.Height / threads
	if rowsPerThread == 0 {
		rowsPerThread = 1
	}

	// Process sharks in parallel
	for t := range threads {
		startRow := t * rowsPerThread
		endRow := startRow + rowsPerThread
		if t == threads-1 {
			endRow = w.Height
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			localEaten := 0
			for i := start; i < end; i++ {
				for j := 0; j < w.Width; j++ {
					mu.Lock()
					if w.Grid[i][j].Type == Shark && !moved[i][j] {
						eaten := w.moveShark(i, j, newGrid, moved)
						if eaten {
							localEaten++
						}
					}
					mu.Unlock()
				}
			}
			mu.Lock()
			fishEaten += localEaten
			mu.Unlock()
		}(startRow, endRow)
	}
	wg.Wait()

	// Process fish in parallel
	for t := range threads {
		startRow := t * rowsPerThread
		endRow := startRow + rowsPerThread
		if t == threads-1 {
			endRow = w.Height
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				for j := 0; j < w.Width; j++ {
					mu.Lock()
					if w.Grid[i][j].Type == Fish && !moved[i][j] {
						w.moveFish(i, j, newGrid, moved)
					}
					mu.Unlock()
				}
			}
		}(startRow, endRow)
	}
	wg.Wait()

	return fishEaten
}

func (w *World) moveShark(y, x int, newGrid [][]Cell, moved [][]bool) bool {
	shark := w.Grid[y][x]
	shark.Energy--
	shark.BreedTime++

	// Find adjacent cells with fish
	fishCells := w.getAdjacentCells(y, x, Fish, moved)
	var targetY, targetX int
	fishEaten := false

	if len(fishCells) > 0 {
		// Eat a fish
		idx := rand.Intn(len(fishCells))
		targetY, targetX = fishCells[idx][0], fishCells[idx][1]
		shark.Energy = w.SharkStarve
		fishEaten = true
	} else {
		// Move to empty cell
		emptyCells := w.getAdjacentCells(y, x, Empty, moved)
		if len(emptyCells) > 0 {
			idx := rand.Intn(len(emptyCells))
			targetY, targetX = emptyCells[idx][0], emptyCells[idx][1]
		} else {
			// Can't move, stay in place
			targetY, targetX = y, x
		}
	}

	// Check if shark dies
	if shark.Energy <= 0 {
		// Shark dies, leave empty
		if targetY != y || targetX != x {
			newGrid[targetY][targetX] = Cell{Type: Empty}
			moved[targetY][targetX] = true
		}
		return fishEaten
	}

	// Move shark
	if shark.BreedTime >= w.SharkBreed {
		// Breed
		newGrid[y][x] = Cell{
			Type:      Shark,
			Energy:    w.SharkStarve,
			BreedTime: 0,
		}
		moved[y][x] = true
		shark.BreedTime = 0
	}

	newGrid[targetY][targetX] = shark
	moved[targetY][targetX] = true

	return fishEaten
}

func (w *World) moveFish(y, x int, newGrid [][]Cell, moved [][]bool) {
	fish := w.Grid[y][x]
	fish.BreedTime++

	// Find empty adjacent cells
	emptyCells := w.getAdjacentCells(y, x, Empty, moved)
	var targetY, targetX int

	if len(emptyCells) > 0 {
		idx := rand.Intn(len(emptyCells))
		targetY, targetX = emptyCells[idx][0], emptyCells[idx][1]
	} else {
		// Can't move
		targetY, targetX = y, x
	}

	// Move fish
	if fish.BreedTime >= w.FishBreed {
		// Breed
		newGrid[y][x] = Cell{
			Type:      Fish,
			BreedTime: 0,
		}
		moved[y][x] = true
		fish.BreedTime = 0
	}

	newGrid[targetY][targetX] = fish
	moved[targetY][targetX] = true
}

func (w *World) getAdjacentCells(y, x int, cellType CellType, moved [][]bool) [][]int {
	var cells [][]int
	directions := [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, dir := range directions {
		ny := (y + dir[0] + w.Height) % w.Height
		nx := (x + dir[1] + w.Width) % w.Width

		if !moved[ny][nx] && w.Grid[ny][nx].Type == cellType {
			cells = append(cells, []int{ny, nx})
		}
	}

	return cells
}
