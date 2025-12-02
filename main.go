package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

	fishEaten := 0

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
					if w.Grid[i][j].Type == Shark {
						mu.Lock()
						if !moved[i][j] {
							eaten := w.moveShark(i, j, newGrid, moved)
							if eaten {
								localEaten++
							}
						}
						mu.Unlock()
					}
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
					if w.Grid[i][j].Type == Fish {
						mu.Lock()
						if !moved[i][j] {
							w.moveFish(i, j, newGrid, moved)
						}
						mu.Unlock()
					}
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

// Colors for rendering
var (
	ColorEmpty = color.RGBA{0, 0, 50, 255}  // Dark blue (ocean)
	ColorFish  = color.RGBA{0, 255, 0, 255} // Green
	ColorShark = color.RGBA{255, 0, 0, 255} // Red
)

// Game implements ebiten.Game interface
type Game struct {
	world      *World
	threads    int
	cellSize   int
	step       int
	maxSteps   int
	updateFreq int
	counter    int
	paused     bool
	ended      bool
	endReason  string
	fishEaten  int
	startTime  time.Time
}

// Update updates the game state
func (g *Game) Update() error {
	// If simulation ended, keep window open but don't update
	if g.ended {
		return nil
	}

	fish, sharks := g.world.Count()
	if fish == 0 {
		fmt.Printf("\nAll fish died at step %d\n", g.step)
		g.ended = true
		g.endReason = "All fish died"
		return nil
	}
	if sharks == 0 {
		fmt.Printf("\nAll sharks died at step %d\n", g.step)
		g.ended = true
		g.endReason = "All sharks died"
		return nil
	}

	// Handle input
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.paused = !g.paused
		time.Sleep(200 * time.Millisecond) // Debounce
	}

	if !g.paused {
		g.counter++
		if g.counter >= g.updateFreq {
			g.fishEaten += g.world.Step(g.threads)
			g.step++
			g.counter = 0
		}
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorEmpty)

	// Draw cells
	for i := 0; i < g.world.Height; i++ {
		for j := 0; j < g.world.Width; j++ {
			cell := g.world.Grid[i][j]
			if cell.Type != Empty {
				x := float32(j * g.cellSize)
				y := float32(i * g.cellSize)
				w := float32(g.cellSize)
				h := float32(g.cellSize)

				var c color.Color
				if cell.Type == Fish {
					c = ColorFish
				} else {
					c = ColorShark
				}

				vector.FillRect(screen, x, y, w, h, c, false)
			}
		}
	}

	// Draw statistics
	fish, sharks := g.world.Count()
	elapsed := time.Since(g.startTime)
	status := "Running"
	if g.paused {
		status = "Paused"
	}
	if g.ended {
		status = "ENDED: " + g.endReason
	}

	stepsDisplay := fmt.Sprintf("%d/%d", g.step, g.maxSteps)
	if g.maxSteps == 0 {
		stepsDisplay = fmt.Sprintf("%d (infinite)", g.step)
	}

	message := fmt.Sprintf(
		"Wa-Tor Simulation [%s]\n"+
			"Step: %s\n"+
			"Fish: %d\n"+
			"Sharks: %d\n"+
			"Fish Eaten: %d\n"+
			"Threads: %d\n"+
			"Time: %.1fs\n"+
			"FPS: %.0f\n"+
			"Update: every %d frames\n",
		status, stepsDisplay, fish, sharks, g.fishEaten, g.threads,
		elapsed.Seconds(), ebiten.ActualFPS(), g.updateFreq,
	)

	if g.ended {
		message += "\nClose window to exit"
	} else {
		message += "\nPress SPACE to pause"
	}

	ebitenutil.DebugPrint(screen, message)
}

// Layout sets the game screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.world.Width * g.cellSize, g.world.Height * g.cellSize
}

func main() {
	// Parse command-line arguments
	numShark := flag.Int("sharks", 100, "Starting population of sharks")
	numFish := flag.Int("fish", 500, "Starting population of fish")
	fishBreed := flag.Int("fbreed", 10, "Fish breeding time")
	sharkBreed := flag.Int("sbreed", 10, "Shark breeding time")
	starve := flag.Int("starve", 8, "Shark starvation time")
	gridSize := flag.Int("size", 80, "Grid dimensions (square)")
	threads := flag.Int("threads", 1, "Number of threads to use")
	steps := flag.Int("steps", 0, "Number of simulation steps (0=infinite)")
	cellSize := flag.Int("cellsize", 8, "Size of each cell in pixels")
	updateFreq := flag.Int("updatefreq", 3, "Update frequency (higher=slower, 1=every frame)")

	flag.Parse()

	// Validate parameters
	if *numShark < 0 || *numFish < 0 || *fishBreed < 1 || *sharkBreed < 1 ||
		*starve < 1 || *gridSize < 1 || *threads < 1 {
		fmt.Println("Error: All parameters must be positive")
		return
	}

	if *numShark+*numFish > (*gridSize * *gridSize) {
		fmt.Println("Error: Too many entities for grid size")
		return
	}

	// Create world
	world := NewWorld(*gridSize, *gridSize, *numFish, *numShark,
		*fishBreed, *sharkBreed, *starve)

	// Create game
	game := &Game{
		world:      world,
		threads:    *threads,
		cellSize:   *cellSize,
		maxSteps:   *steps,
		updateFreq: *updateFreq,
		startTime:  time.Now(),
	}

	// Set up window
	ebiten.SetWindowSize((*gridSize)*(*cellSize), (*gridSize)*(*cellSize))
	ebiten.SetWindowTitle("Wa-Tor Simulation")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	fmt.Printf("Wa-Tor Simulation\n")
	fmt.Printf("Grid: %dx%d, Fish: %d, Sharks: %d\n", *gridSize, *gridSize, *numFish, *numShark)
	fmt.Printf("Fish Breed: %d, Shark Breed: %d, Starve: %d\n", *fishBreed, *sharkBreed, *starve)
	fmt.Printf("Threads: %d, Max Steps: %d\n\n", *threads, *steps)

	// Run game
	if err := ebiten.RunGame(game); err != nil {
		if err != ebiten.Termination {
			log.Fatal(err)
		}
	}

	// Print final statistics
	fish, sharks := world.Count()
	elapsed := time.Since(game.startTime)
	fmt.Printf("\nSimulation completed at step %d\n", game.step)
	fmt.Printf("Final populations - Fish: %d, Sharks: %d\n", fish, sharks)
	fmt.Printf("Total fish eaten: %d\n", game.fishEaten)
	fmt.Printf("Total execution time: %v\n", elapsed)
	if game.step > 0 {
		fmt.Printf("Average time per step: %v\n", elapsed/time.Duration(game.step))
	}
}
