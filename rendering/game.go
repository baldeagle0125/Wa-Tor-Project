package rendering

import (
	"fmt"
	"image/color"
	"time"

	"github.com/baldeagle0125/Wa-Tor-Project/simulation"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Colors for rendering
var (
	ColorEmpty = color.RGBA{0, 0, 50, 255}
	ColorFish  = color.RGBA{0, 255, 0, 255}
	ColorShark = color.RGBA{255, 0, 0, 255}
)

// Game implements ebiten.Game interface
type Game struct {
	world      *simulation.World
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

// NewGame creates a new Game instance
func NewGame(world *simulation.World, threads, cellSize, maxSteps, updateFreq int) *Game {
	return &Game{
		world:      world,
		threads:    threads,
		cellSize:   cellSize,
		maxSteps:   maxSteps,
		updateFreq: updateFreq,
		startTime:  time.Now(),
	}
}

// Update updates the game state
func (g *Game) Update() error {
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

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.paused = !g.paused
		time.Sleep(200 * time.Millisecond)
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

	for i := 0; i < g.world.Height; i++ {
		for j := 0; j < g.world.Width; j++ {
			cell := g.world.Grid[i][j]
			if cell.Type != simulation.Empty {
				x := float32(j * g.cellSize)
				y := float32(i * g.cellSize)
				w := float32(g.cellSize)
				h := float32(g.cellSize)

				var c color.Color
				if cell.Type == simulation.Fish {
					c = ColorFish
				} else {
					c = ColorShark
				}

				vector.FillRect(screen, x, y, w, h, c, false)
			}
		}
	}

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

// GetStats returns the final statistics of the simulation
func (g *Game) GetStats() (step int, fishEaten int, elapsed time.Duration) {
	return g.step, g.fishEaten, time.Since(g.startTime)
}
