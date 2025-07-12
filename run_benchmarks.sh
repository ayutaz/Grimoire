#!/bin/bash

# Run performance benchmarks for Grimoire optimizations

echo "Running Grimoire Performance Benchmarks"
echo "======================================"
echo

# Set benchmark time
BENCH_TIME=${BENCH_TIME:-10s}

# Create results directory
RESULTS_DIR="benchmark_results"
mkdir -p $RESULTS_DIR

# Run detector benchmarks
echo "1. Running Detector Benchmarks..."
echo "---------------------------------"
go test -bench=BenchmarkDetectorComparison -benchtime=$BENCH_TIME -benchmem ./internal/detector/ | tee $RESULTS_DIR/detector_comparison.txt
echo

echo "2. Running Contour Finding Benchmarks..."
echo "---------------------------------------"
go test -bench=BenchmarkFindContoursComparison -benchtime=$BENCH_TIME -benchmem ./internal/detector/ | tee $RESULTS_DIR/contours_comparison.txt
echo

echo "3. Running Symbol Detection Benchmarks..."
echo "----------------------------------------"
go test -bench=BenchmarkSymbolDetectionComparison -benchtime=$BENCH_TIME -benchmem ./internal/detector/ | tee $RESULTS_DIR/symbol_detection.txt
echo

echo "4. Running Cache Performance Benchmarks..."
echo "-----------------------------------------"
go test -bench=BenchmarkDetectorCache -benchtime=$BENCH_TIME -benchmem ./internal/detector/ | tee $RESULTS_DIR/cache_performance.txt
echo

# Run parser benchmarks
echo "5. Running Parser Benchmarks..."
echo "-------------------------------"
go test -bench=BenchmarkParserComparison -benchtime=$BENCH_TIME -benchmem ./internal/parser/ | tee $RESULTS_DIR/parser_comparison.txt
echo

echo "6. Running Symbol Graph Building Benchmarks..."
echo "---------------------------------------------"
go test -bench=BenchmarkSymbolGraphBuilding -benchtime=$BENCH_TIME -benchmem ./internal/parser/ | tee $RESULTS_DIR/symbol_graph.txt
echo

echo "7. Running Connection Inference Benchmarks..."
echo "--------------------------------------------"
go test -bench=BenchmarkConnectionInference -benchtime=$BENCH_TIME -benchmem ./internal/parser/ | tee $RESULTS_DIR/connection_inference.txt
echo

echo "8. Running Spatial Index Benchmarks..."
echo "-------------------------------------"
go test -bench=BenchmarkSpatialIndex -benchtime=$BENCH_TIME -benchmem ./internal/parser/ | tee $RESULTS_DIR/spatial_index.txt
echo

# Run end-to-end benchmarks
echo "9. Running End-to-End Performance Benchmarks..."
echo "----------------------------------------------"
go test -bench=BenchmarkEndToEndPerformance -benchtime=$BENCH_TIME -benchmem ./test/ | tee $RESULTS_DIR/e2e_performance.txt
echo

echo "10. Running Pipeline Stage Benchmarks..."
echo "---------------------------------------"
go test -bench=BenchmarkPipelineStages -benchtime=$BENCH_TIME -benchmem ./test/ | tee $RESULTS_DIR/pipeline_stages.txt
echo

echo "11. Running Memory Usage Benchmarks..."
echo "-------------------------------------"
go test -bench=BenchmarkMemoryUsage -benchtime=$BENCH_TIME -benchmem ./test/ | tee $RESULTS_DIR/memory_usage.txt
echo

# Generate summary report
echo "Generating Performance Summary..."
echo "================================="
cat > $RESULTS_DIR/performance_summary.md << EOF
# Grimoire Performance Optimization Results

## Summary

This document summarizes the performance improvements achieved through optimization of the Grimoire project.

### Key Optimizations Implemented:

1. **Detector Optimizations**
   - Parallel contour processing using goroutines
   - Caching for preprocessed images and intermediate results
   - Optimized image preprocessing pipeline

2. **Parser Optimizations**
   - Spatial indexing for fast symbol lookups
   - Optimized symbol graph building
   - Improved connection inference algorithm
   - Expression parsing memoization

3. **Compiler Optimizations**
   - Already uses strings.Builder efficiently

## Benchmark Results

### Detector Performance

EOF

# Extract key metrics from benchmark results
echo "### Contour Finding Performance" >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md
grep -E "Benchmark.*Comparison" $RESULTS_DIR/contours_comparison.txt | head -6 >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md
echo >> $RESULTS_DIR/performance_summary.md

echo "### Parser Performance" >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md
grep -E "BenchmarkParserComparison" $RESULTS_DIR/parser_comparison.txt | head -8 >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md
echo >> $RESULTS_DIR/performance_summary.md

echo "### End-to-End Performance" >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md
grep -E "BenchmarkEndToEndPerformance" $RESULTS_DIR/e2e_performance.txt | head -6 >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md
echo >> $RESULTS_DIR/performance_summary.md

echo "### Memory Usage" >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md
grep -E "BenchmarkMemoryUsage" $RESULTS_DIR/memory_usage.txt | head -4 >> $RESULTS_DIR/performance_summary.md
echo '```' >> $RESULTS_DIR/performance_summary.md

echo
echo "Performance benchmarks completed!"
echo "Results saved to: $RESULTS_DIR/"
echo "Summary report: $RESULTS_DIR/performance_summary.md"