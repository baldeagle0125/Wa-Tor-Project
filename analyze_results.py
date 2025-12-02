#!/usr/bin/env python3
"""
Analyze Wa-Tor benchmark results and generate performance report with graphs.
"""

import re
import sys

def parse_results(filename):
    """Parse benchmark results file."""
    with open(filename, 'r') as f:
        content = f.read()
    
    # Extract thread counts and times
    threads = []
    times_ms = []
    
    pattern = r'Threads: (\d+).*?Execution Time: ([\d.]+)(ms|µs|s)'
    matches = re.findall(pattern, content, re.DOTALL)
    
    for match in matches:
        thread_count = int(match[0])
        time_value = float(match[1])
        unit = match[2]
        
        # Convert to milliseconds
        if unit == 's':
            time_ms = time_value * 1000
        elif unit == 'µs':
            time_ms = time_value / 1000
        else:  # ms
            time_ms = time_value
        
        threads.append(thread_count)
        times_ms.append(time_ms)
    
    return threads, times_ms

def calculate_speedup(threads, times):
    """Calculate speedup relative to single thread."""
    if not times or times[0] == 0:
        return []
    
    base_time = times[0]
    speedups = [base_time / t for t in times]
    return speedups

def calculate_efficiency(threads, speedups):
    """Calculate parallel efficiency."""
    return [s / t * 100 for s, t in zip(speedups, threads)]

def generate_report(threads, times, speedups, efficiencies):
    """Generate markdown report."""
    report = """# Wa-Tor Simulation Performance Results

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
"""
    
    for t, time, speedup, eff in zip(threads, times, speedups, efficiencies):
        report += f"| {t:7d} | {time:18.2f} | {speedup:6.2f}x | {eff:13.1f}% |\n"
    
    report += """
### Analysis

**Speedup**: Ratio of single-thread time to multi-thread time (higher is better)
- Ideal speedup would be linear (2x for 2 threads, 4x for 4 threads, etc.)

**Efficiency**: Percentage of ideal parallel performance achieved
- 100% efficiency means perfect scaling
- Values below 100% indicate overhead from synchronization and thread management

### Observations

"""
    
    # Add observations
    max_speedup_idx = speedups.index(max(speedups))
    max_speedup_threads = threads[max_speedup_idx]
    max_speedup = speedups[max_speedup_idx]
    
    report += f"1. **Best Performance**: {max_speedup_threads} thread(s) achieved {max_speedup:.2f}x speedup\n"
    report += f"2. **Single Thread Baseline**: {times[0]:.2f} ms\n"
    
    if max_speedup < 1.5:
        report += "3. **Limited Scaling**: The parallel implementation shows limited speedup, likely due to:\n"
        report += "   - Fine-grained locking causing contention\n"
        report += "   - Synchronization overhead exceeding computation benefits\n"
        report += "   - Random memory access patterns reducing cache efficiency\n"
    elif max_speedup > 2.0:
        report += "3. **Good Scaling**: The parallel implementation shows effective scaling\n"
    else:
        report += "3. **Moderate Scaling**: Some benefit from parallelization, but with overhead\n"
    
    report += """
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
"""
    
    return report

def generate_graphs(threads, times, speedups, efficiencies):
    """Generate performance graphs."""
    
    # Import matplotlib only when needed
    try:
        import matplotlib
        matplotlib.use('Agg')  # Use non-interactive backend
        import matplotlib.pyplot as plt
    except ImportError:
        print("Warning: matplotlib not installed. Skipping graph generation.")
        print("To generate graphs, install matplotlib:")
        print("  pip3 install matplotlib")
        return
    
    # Graph 1: Execution Time
    plt.figure(figsize=(10, 6))
    plt.plot(threads, times, 'b-o', linewidth=2, markersize=8)
    plt.xlabel('Number of Threads', fontsize=12)
    plt.ylabel('Execution Time (ms)', fontsize=12)
    plt.title('Wa-Tor Simulation Execution Time vs Thread Count', fontsize=14, fontweight='bold')
    plt.grid(True, alpha=0.3)
    plt.xticks(threads)
    for t, time in zip(threads, times):
        plt.annotate(f'{time:.1f}ms', xy=(t, time), xytext=(5, 5), 
                    textcoords='offset points', fontsize=9)
    plt.tight_layout()
    plt.savefig('execution_time.png', dpi=150)
    print("Generated: execution_time.png")
    plt.close()
    
    # Graph 2: Speedup and Efficiency
    fig, ax1 = plt.subplots(figsize=(10, 6))
    
    color1 = 'tab:blue'
    ax1.set_xlabel('Number of Threads', fontsize=12)
    ax1.set_ylabel('Speedup', color=color1, fontsize=12)
    line1 = ax1.plot(threads, speedups, 'b-o', linewidth=2, markersize=8, label='Speedup')
    
    # Ideal speedup line
    ideal_speedup = threads
    ax1.plot(threads, ideal_speedup, 'k--', linewidth=1, alpha=0.5, label='Ideal (Linear)')
    
    ax1.tick_params(axis='y', labelcolor=color1)
    ax1.set_xticks(threads)
    ax1.grid(True, alpha=0.3)
    ax1.legend(loc='upper left')
    
    # Add efficiency on secondary axis
    ax2 = ax1.twinx()
    color2 = 'tab:red'
    ax2.set_ylabel('Efficiency (%)', color=color2, fontsize=12)
    line2 = ax2.plot(threads, efficiencies, 'r-s', linewidth=2, markersize=8, label='Efficiency')
    ax2.tick_params(axis='y', labelcolor=color2)
    ax2.set_ylim(0, 120)
    ax2.legend(loc='upper right')
    
    plt.title('Wa-Tor Simulation Speedup and Efficiency', fontsize=14, fontweight='bold')
    plt.tight_layout()
    plt.savefig('speedup.png', dpi=150)
    print("Generated: speedup.png")
    plt.close()

def main():
    # Parse results
    threads, times = parse_results('benchmark_results.txt')
    
    if not threads:
        print("Error: No results found in benchmark_results.txt")
        sys.exit(1)
    
    # Calculate metrics
    speedups = calculate_speedup(threads, times)
    efficiencies = calculate_efficiency(threads, speedups)
    
    # Generate report
    report = generate_report(threads, times, speedups, efficiencies)
    
    with open('PERFORMANCE.md', 'w') as f:
        f.write(report)
    print("Generated: PERFORMANCE.md")
    
    # Generate graphs
    try:
        generate_graphs(threads, times, speedups, efficiencies)
    except Exception as e:
        print(f"Warning: Could not generate graphs: {e}")
        print("Install matplotlib with: pip3 install matplotlib")
    
    print("\nPerformance analysis complete!")

if __name__ == '__main__':
    main()
