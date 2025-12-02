#!/bin/bash

# Wa-Tor Performance Benchmark Script
# Tests simulation with 1, 2, 4, and 8 threads

echo "Building Wa-Tor simulation..."
go build -o wa-tor

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Running performance benchmarks..."
echo "================================"
echo ""

# Test parameters
STEPS=1000
SIZE=100
FISH=2000
SHARKS=400
FBREED=10
SBREED=10
STARVE=8

# Results file
RESULTS="benchmark_results.txt"
> $RESULTS  # Clear file

echo "Benchmark Configuration:" | tee -a $RESULTS
echo "Steps: $STEPS" | tee -a $RESULTS
echo "Grid Size: ${SIZE}x${SIZE}" | tee -a $RESULTS
echo "Fish: $FISH" | tee -a $RESULTS
echo "Sharks: $SHARKS" | tee -a $RESULTS
echo "Fish Breed: $FBREED" | tee -a $RESULTS
echo "Shark Breed: $SBREED" | tee -a $RESULTS
echo "Starve: $STARVE" | tee -a $RESULTS
echo "" | tee -a $RESULTS
echo "================================" | tee -a $RESULTS
echo "" | tee -a $RESULTS

# Run benchmarks for each thread count
for THREADS in 1 2 4 8; do
    echo "Testing with $THREADS thread(s)..." | tee -a $RESULTS
    
    # Run simulation and capture output
    OUTPUT=$(./wa-tor -steps $STEPS -size $SIZE -fish $FISH -sharks $SHARKS \
             -fbreed $FBREED -sbreed $SBREED -starve $STARVE -threads $THREADS 2>&1)
    
    # Extract execution time
    TIME=$(echo "$OUTPUT" | grep "Total execution time:" | awk '{print $4}')
    
    echo "  Threads: $THREADS" | tee -a $RESULTS
    echo "  Execution Time: $TIME" | tee -a $RESULTS
    echo "" | tee -a $RESULTS
done

echo "Benchmark complete! Results saved to $RESULTS"
echo ""
echo "Summary:"
cat $RESULTS
