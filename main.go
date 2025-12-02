package main

import (
	"fmt"
	"log"
	"time"

	"wa-tor/config"
	"wa-tor/rendering"
	"wa-tor/simulation"

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

	// Run in headless mode if steps is specified
	if cfg.Steps > 0 {
		runHeadless(world, cfg)
		return
	}

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

func runHeadless(world *simulation.World, cfg *config.Config) {
	fmt.Println("Running in headless mode...")
	startTime := time.Now()
	totalFishEaten := 0

	for step := 0; step < cfg.Steps; step++ {
		fish, sharks := world.Count()

		// Check termination conditions
		if fish == 0 {
			fmt.Printf("\nAll fish died at step %d\n", step)
			break
		}
		if sharks == 0 {
			fmt.Printf("\nAll sharks died at step %d\n", step)
			break
		}

		// Perform simulation step
		fishEaten := world.Step(cfg.Threads)
		totalFishEaten += fishEaten
	}

	elapsed := time.Since(startTime)
	fish, sharks := world.Count()

	// Print final statistics
	fmt.Printf("\nSimulation completed\n")
	fmt.Printf("Steps completed: %d\n", cfg.Steps)
	fmt.Printf("Final populations - Fish: %d, Sharks: %d\n", fish, sharks)
	fmt.Printf("Total fish eaten: %d\n", totalFishEaten)
	fmt.Printf("Total execution time: %v\n", elapsed)
	if cfg.Steps > 0 {
		fmt.Printf("Average time per step: %v\n", elapsed/time.Duration(cfg.Steps))
	}
}
