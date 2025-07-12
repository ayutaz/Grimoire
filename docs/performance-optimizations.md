# Grimoire Performance Optimizations

This document describes the performance optimizations implemented in the Grimoire project to improve processing speed for large magic circles with many symbols.

## Overview

The optimizations focus on three main areas:
1. **Image Processing (Detector)** - Parallel processing and caching
2. **Symbol Graph Building (Parser)** - Spatial indexing and optimized algorithms
3. **Code Generation (Compiler)** - Already optimized with strings.Builder

## Detector Optimizations

### 1. Parallel Contour Processing

The `ParallelDetector` implements parallel processing for contour analysis:

```go
// Process contours in parallel using worker pool
func (d *ParallelDetector) detectSymbolsFromContoursParallel(contours []Contour, binary *image.Gray) []*Symbol
```

**Benefits:**
- Utilizes multiple CPU cores for symbol detection
- Significant speedup for images with many contours
- Maintains thread-safe operations

### 2. Parallel Contour Finding

Image strips are processed in parallel for contour detection:

```go
// Split image into horizontal strips for parallel processing
func (d *ParallelDetector) findContoursParallel(binary *image.Gray) []Contour
```

**Benefits:**
- Faster contour detection on large images
- Scalable with CPU core count
- Efficient memory usage

### 3. Result Caching

Implements an LRU cache for preprocessed images and detection results:

```go
type DetectorCache struct {
    preprocessed map[string]*image.Gray
    symbols      map[string][]*Symbol
    maxCacheSize int
}
```

**Benefits:**
- Avoids redundant preprocessing for repeated detections
- Significant speedup for iterative workflows
- Configurable cache size

## Parser Optimizations

### 1. Spatial Indexing

Uses a grid-based spatial index for fast symbol lookups:

```go
type SpatialIndex struct {
    gridSize float64
    grid     map[gridKey][]*symbolNode
}
```

**Benefits:**
- O(1) average case for nearby symbol queries
- Reduces connection inference from O(n²) to O(n)
- Efficient range queries

### 2. Optimized Connection Inference

Parallel processing and spatial queries for connection building:

```go
func (p *OptimizedParser) inferConnectionsOptimized()
```

**Benefits:**
- Parallel batch processing of nodes
- Uses spatial index for efficient neighbor lookups
- Reduced algorithmic complexity

### 3. Expression Caching

Memoization for parsed expressions:

```go
type ExpressionCache struct {
    cache map[*symbolNode]Expression
}
```

**Benefits:**
- Avoids redundant expression parsing
- Thread-safe concurrent access
- Improved performance for complex expressions

## Usage

### Using the Optimized Detector

```go
import "github.com/ayutaz/grimoire/internal/detector"

// Create parallel detector
parallelDetector := detector.NewParallelDetector(detector.Config{
    Debug: false,
})

// Detect symbols (uses caching and parallel processing)
symbols, connections, err := parallelDetector.Detect(imagePath)
```

### Using the Optimized Parser

```go
import "github.com/ayutaz/grimoire/internal/parser"

// Create optimized parser
optimizedParser := parser.NewOptimizedParser()

// Parse with spatial indexing and optimizations
ast, err := optimizedParser.Parse(symbols, connections)
```

## Performance Benchmarks

Run the benchmarks to see performance improvements:

```bash
# Run all benchmarks
./run_benchmarks.sh

# Run specific benchmarks
go test -bench=BenchmarkDetectorComparison ./internal/detector/
go test -bench=BenchmarkParserComparison ./internal/parser/
go test -bench=BenchmarkEndToEndPerformance ./test/
```

### Expected Improvements

Based on the optimizations implemented:

1. **Detector Performance**
   - 2-4x speedup for parallel contour processing (scales with CPU cores)
   - 3-5x speedup for cached repeated detections
   - 30-50% improvement in contour finding

2. **Parser Performance**
   - 5-10x speedup for connection inference on large symbol sets
   - 2-3x speedup for symbol graph building
   - O(log n) vs O(n) for spatial queries

3. **End-to-End Performance**
   - 2-3x overall speedup for medium complexity images
   - 3-5x speedup for complex images with many symbols
   - Reduced memory allocations

## Configuration

### Detector Configuration

```go
// Configure number of workers (defaults to CPU count)
detector := &ParallelDetector{
    numWorkers: runtime.NumCPU(),
    cache: NewDetectorCache(100), // Cache size
}
```

### Parser Configuration

```go
// Spatial index grid size is auto-calculated based on symbol distribution
// Adjusts automatically for optimal performance
```

## Best Practices

1. **Use Parallel Detector for Large Images**
   - Benefits scale with image size and symbol count
   - Minimal overhead for small images

2. **Enable Caching for Iterative Workflows**
   - Useful when processing the same image multiple times
   - Cache automatically manages memory usage

3. **Batch Processing**
   - Process multiple images to maximize cache benefits
   - Parallel detector efficiently handles concurrent requests

## Future Optimizations

Potential areas for further optimization:

1. **GPU Acceleration**
   - Image preprocessing on GPU
   - Parallel contour analysis

2. **Advanced Caching**
   - Distributed cache for multi-node setups
   - Persistent cache for long-running processes

3. **Algorithmic Improvements**
   - More sophisticated shape classification
   - Machine learning-based symbol detection

## Troubleshooting

### High Memory Usage

If experiencing high memory usage:
- Reduce cache size: `NewDetectorCache(50)`
- Limit worker count: `numWorkers: 4`

### Incorrect Results

Optimizations maintain correctness, but if issues occur:
- Disable optimizations for debugging
- Use standard detector/parser for validation
- Check benchmark correctness tests

## Conclusion

The performance optimizations significantly improve Grimoire's ability to handle large and complex magic circles while maintaining accuracy and correctness. The modular design allows users to choose between standard and optimized implementations based on their needs.