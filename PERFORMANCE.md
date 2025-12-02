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
|       1 |             380.41 |   1.00x |         100.0% |
|       2 |             822.50 |   0.46x |          23.1% |
|       4 |             705.72 |   0.54x |          13.5% |
|       8 |             785.94 |   0.48x |           6.1% |

### Analysis

**Speedup**: Ratio of single-thread time to multi-thread time (higher is better)
- Ideal speedup would be linear (2x for 2 threads, 4x for 4 threads, etc.)

**Efficiency**: Percentage of ideal parallel performance achieved
- 100% efficiency means perfect scaling
- Values below 100% indicate overhead from synchronization and thread management

### Observations

1. **Best Performance**: 1 thread(s) achieved 1.00x speedup
2. **Single Thread Baseline**: 380.41 ms
3. **Limited Scaling**: The parallel implementation shows limited speedup, likely due to:
   - Fine-grained locking causing mutex contention
   - Synchronization overhead exceeding computation benefits
   - Random entity access patterns reducing cache efficiency
   - Toroidal world boundaries requiring cross-partition coordination

### Implementation Notes

The parallel implementation uses:
- Fine-grained mutex locking per entity operation
- Random entity processing order (Fisher-Yates shuffle)
- Work distribution by entity count across thread pool
- Thread-safe grid updates with moved tracking

The mutex contention for shared grid access creates a bottleneck that limits
parallel scaling. This is a common challenge in cellular automaton simulations
where entities can interact with any neighboring cell, especially in a toroidal
world where edges wrap around.

### Future Optimizations

Potential improvements for better parallel performance:
1. **Domain decomposition**: Partition grid into regions with minimal boundary overlap
2. **Double buffering**: Reduce lock contention with alternating read/write grids
3. **Lock-free algorithms**: Use atomic operations where possible
4. **Coarser granularity**: Lock larger grid sections to reduce synchronization overhead

## Benchmark Data

Raw benchmark results are available in `benchmark_results.txt`.
