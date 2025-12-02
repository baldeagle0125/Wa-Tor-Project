# Wa-Tor Simulation Performance Results

## Test Configuration
- **Steps**: 1000 chronons
- **Grid Size**: 100x100
- **Initial Population**: 2000 fish, 400 sharks
- **Test Platform**: macOS (Apple Silicon/Intel)
- **Go Version**: 1.25+

## Performance Results

### Execution Times

| Threads | Execution Time (ms) | Speedup | Efficiency (%) |
|---------|--------------------:|--------:|---------------:|
|       1 |             378.20 |   1.00x |         100.0% |
|       2 |             892.96 |   0.42x |          21.2% |
|       4 |             766.48 |   0.49x |          12.3% |
|       8 |             827.23 |   0.46x |           5.7% |

### Analysis

**Speedup**: Ratio of single-thread time to multi-thread time (higher is better)
- Ideal speedup would be linear (2x for 2 threads, 4x for 4 threads, etc.)

**Efficiency**: Percentage of ideal parallel performance achieved
- 100% efficiency means perfect scaling
- Values below 100% indicate overhead from synchronization and thread management

### Observations

1. **Best Performance**: 1 thread(s) achieved 1.00x speedup
2. **Single Thread Baseline**: 378.20 ms
3. **Limited Scaling**: The parallel implementation shows limited speedup, likely due to:
   - Fine-grained locking causing contention
   - Synchronization overhead exceeding computation benefits
   - Random memory access patterns reducing cache efficiency

### Implementation Notes

The parallel implementation uses:
- Fine-grained mutex locking per entity operation
- Random entity processing order (Fisher-Yates shuffle)
- Work distribution across thread pool

The mutex contention for shared grid access creates a bottleneck that limits
parallel scaling. This is a common challenge in cellular automaton simulations
where entities can interact with any neighboring cell.

## Performance Graphs

See the generated PNG files for visual representation:
- `execution_time.png` - Execution time vs thread count
- `speedup.png` - Speedup and efficiency vs thread count
