# Wa-Tor Simulation

A Go implementation of the Wa-Tor predator-prey simulation with parallel processing and visualization capabilities.

## Overview

Wa-Tor is a population dynamics simulation of predators (sharks) and prey (fish) on a toroidal planet. This implementation features:
- Configurable population parameters
- Multi-threaded parallel processing
- Real-time graphical visualization using Ebiten
- Headless mode for performance testing
- Random chronon ordering for unbiased simulation

## Building

```bash
go build -o wa-tor
```

## Running

### Interactive Mode (with visualization)
```bash
./wa-tor
```

### Headless Mode (no GUI)
```bash
./wa-tor -steps 1000
```

## Command-Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-fish` | 500 | Starting population of fish |
| `-sharks` | 100 | Starting population of sharks |
| `-fbreed` | 10 | Fish breeding time (chronons) |
| `-sbreed` | 10 | Shark breeding time (chronons) |
| `-starve` | 8 | Shark starvation time (chronons) |
| `-size` | 80 | Grid dimensions (square) |
| `-threads` | 1 | Number of parallel threads to use |
| `-steps` | 0 | Max simulation steps (0=infinite, runs headless if >0) |
| `-cellsize` | 8 | Size of each cell in pixels (visualization only) |
| `-updatefreq` | 3 | Update frequency - higher=slower (visualization only) |

## Examples

```bash
# Large grid with more sharks
./wa-tor -size 100 -sharks 200 -fish 800

# Fast breeding fish
./wa-tor -fbreed 5 -sbreed 15

# Performance test with 4 threads
./wa-tor -threads 4 -steps 5000

# Smaller cells for detailed view
./wa-tor -cellsize 4 -size 120
```

## Controls (Interactive Mode)

- **SPACE**: Pause/Resume simulation
- Window can be resized

## Implementation Details

- **Toroidal World**: Edges wrap around (top connects to bottom, left to right)
- **Random Processing**: Entities are processed in random order each chronon
- **Parallel Processing**: World is partitioned by rows for multi-threaded execution
- **Breeding**: Animals breed after reaching their breed time
- **Starvation**: Sharks die if they don't eat within their starve time
- **Priority**: Sharks move first, then fish

## Output

The simulation displays real-time statistics including:
- Current step number
- Fish and shark populations
- Total fish eaten
- Execution time and FPS
- Thread count

Final statistics are printed upon completion or termination.

## Requirements

- Go 1.21 or higher
- Dependencies are managed via go.mod

## Performance Results

See [PERFORMANCE.md](PERFORMANCE.md) for detailed benchmark results including:
- Execution times for 1, 2, 4, and 8 threads
- Speedup analysis
- Parallel efficiency metrics
- Performance observations and optimization notes

To run your own benchmarks:
```bash
./benchmark.sh
```
