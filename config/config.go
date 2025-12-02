package config

import (
	"flag"
	"fmt"
)

// Config holds all simulation configuration parameters
type Config struct {
	NumShark   int
	NumFish    int
	FishBreed  int
	SharkBreed int
	Starve     int
	GridSize   int
	Threads    int
	Steps      int
	CellSize   int
	UpdateFreq int
}

// ParseFlags parses command-line flags and returns a Config
func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.IntVar(&cfg.NumShark, "sharks", 100, "Starting population of sharks")
	flag.IntVar(&cfg.NumFish, "fish", 500, "Starting population of fish")
	flag.IntVar(&cfg.FishBreed, "fbreed", 10, "Fish breeding time")
	flag.IntVar(&cfg.SharkBreed, "sbreed", 10, "Shark breeding time")
	flag.IntVar(&cfg.Starve, "starve", 8, "Shark starvation time")
	flag.IntVar(&cfg.GridSize, "size", 80, "Grid dimensions (square)")
	flag.IntVar(&cfg.Threads, "threads", 1, "Number of threads to use")
	flag.IntVar(&cfg.Steps, "steps", 0, "Number of simulation steps (0=infinite)")
	flag.IntVar(&cfg.CellSize, "cellsize", 8, "Size of each cell in pixels")
	flag.IntVar(&cfg.UpdateFreq, "updatefreq", 3, "Update frequency (higher=slower, 1=every frame)")

	flag.Parse()

	// Validate parameters
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if configuration parameters are valid
func (c *Config) Validate() error {
	if c.NumShark < 0 || c.NumFish < 0 || c.FishBreed < 1 || c.SharkBreed < 1 ||
		c.Starve < 1 || c.GridSize < 1 || c.Threads < 1 {
		return fmt.Errorf("all parameters must be positive")
	}

	if c.NumShark+c.NumFish > (c.GridSize * c.GridSize) {
		return fmt.Errorf("too many entities for grid size")
	}

	return nil
}

// Print displays the configuration parameters
func (c *Config) Print() {
	fmt.Printf("Wa-Tor Simulation\n")
	fmt.Printf("Grid: %dx%d, Fish: %d, Sharks: %d\n", c.GridSize, c.GridSize, c.NumFish, c.NumShark)
	fmt.Printf("Fish Breed: %d, Shark Breed: %d, Starve: %d\n", c.FishBreed, c.SharkBreed, c.Starve)
	fmt.Printf("Threads: %d, Max Steps: %d\n\n", c.Threads, c.Steps)
}
