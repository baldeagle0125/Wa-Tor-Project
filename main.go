package main

import (
	"fmt"
	"log"
	"time"

	"github.com/baldeagle0125/Wa-Tor-Project/config"
	"github.com/baldeagle0125/Wa-Tor-Project/rendering"
	"github.com/baldeagle0125/Wa-Tor-Project/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Parse configuration from command-line flags
	cfg, err := config.ParseFlags()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Display configuration
	cfg.Print()

	// Create world with configuration parameters
	world := simulation.NewWorld(
		cfg.GridSize, cfg.GridSize,
		cfg.NumFish, cfg.NumShark,
		cfg.FishBreed, cfg.SharkBreed, cfg.Starve,
	)

	// Create game with rendering configuration
	game := rendering.NewGame(
		world,
		cfg.Threads,
		cfg.CellSize,
		cfg.Steps,
		cfg.UpdateFreq,
	)

	// Set up window
	ebiten.SetWindowSize(cfg.GridSize*cfg.CellSize, cfg.GridSize*cfg.CellSize)
	ebiten.SetWindowTitle("Wa-Tor Simulation")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	if err := ebiten.RunGame(game); err != nil {
		if err != ebiten.Termination {
			log.Fatal(err)
		}
	}

	// Print final statistics
	fish, sharks := world.Count()
	step, fishEaten, elapsed := game.GetStats()
	fmt.Printf("\nSimulation completed at step %d\n", step)
	fmt.Printf("Final populations - Fish: %d, Sharks: %d\n", fish, sharks)
	fmt.Printf("Total fish eaten: %d\n", fishEaten)
	fmt.Printf("Total execution time: %v\n", elapsed)
	if step > 0 {
		fmt.Printf("Average time per step: %v\n", elapsed/time.Duration(step))
	}
}
