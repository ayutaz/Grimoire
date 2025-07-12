package test

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ayutaz/grimoire/internal/compiler"
	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/ayutaz/grimoire/internal/parser"
)

// createPerformanceTestImage creates a test image with specified complexity
func createPerformanceTestImage(filename string, complexity string) error {
	var size int
	var symbols []symbolSpec
	
	switch complexity {
	case "simple":
		size = 400
		symbols = []symbolSpec{
			{detector.OuterCircle, 200, 200, 180},
			{detector.DoubleCircle, 200, 100, 30},
			{detector.Star, 200, 200, 20},
		}
	case "medium":
		size = 800
		symbols = []symbolSpec{
			{detector.OuterCircle, 400, 400, 380},
			{detector.DoubleCircle, 400, 100, 40},
			// Variables
			{detector.Square, 200, 200, 30},
			{detector.Square, 300, 200, 30},
			{detector.Square, 400, 200, 30},
			// Operators
			{detector.Convergence, 250, 250, 25},
			{detector.Amplification, 350, 250, 25},
			// Control flow
			{detector.Triangle, 300, 350, 40},
			{detector.Pentagon, 400, 450, 40},
			// Outputs
			{detector.Star, 250, 500, 20},
			{detector.Star, 350, 500, 20},
			{detector.Star, 450, 500, 20},
		}
	case "complex":
		size = 1200
		symbols = []symbolSpec{
			{detector.OuterCircle, 600, 600, 580},
			{detector.DoubleCircle, 600, 100, 50},
		}
		// Add many symbols in a grid pattern
		for y := 200; y < 1000; y += 100 {
			for x := 200; x < 1000; x += 100 {
				switch ((y-200)/100 + (x-200)/100) % 5 {
				case 0:
					symbols = append(symbols, symbolSpec{detector.Square, x, y, 30})
				case 1:
					symbols = append(symbols, symbolSpec{detector.Circle, x, y, 25})
				case 2:
					symbols = append(symbols, symbolSpec{detector.Triangle, x, y, 30})
				case 3:
					symbols = append(symbols, symbolSpec{detector.Pentagon, x, y, 30})
				case 4:
					symbols = append(symbols, symbolSpec{detector.Star, x, y, 20})
				}
			}
		}
	default:
		return fmt.Errorf("unknown complexity: %s", complexity)
	}
	
	// Create image
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Fill with white
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	
	// Draw symbols
	for _, sym := range symbols {
		drawSymbol(img, sym)
	}
	
	// Save image
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return png.Encode(file, img)
}

type symbolSpec struct {
	symbolType detector.SymbolType
	x, y       int
	size       int
}

func drawSymbol(img *image.RGBA, spec symbolSpec) {
	black := color.RGBA{0, 0, 0, 255}
	
	switch spec.symbolType {
	case detector.OuterCircle:
		drawRGBACircle(img, spec.x, spec.y, spec.size, 5, black)
	case detector.DoubleCircle:
		drawRGBACircle(img, spec.x, spec.y, spec.size, 3, black)
		drawRGBACircle(img, spec.x, spec.y, spec.size-10, 3, black)
	case detector.Circle:
		drawRGBACircle(img, spec.x, spec.y, spec.size, 3, black)
	case detector.Square:
		drawRGBASquare(img, spec.x, spec.y, spec.size, black)
	case detector.Triangle:
		drawRGBATriangle(img, spec.x, spec.y, spec.size, black)
	case detector.Pentagon:
		drawRGBAPentagon(img, spec.x, spec.y, spec.size, black)
	case detector.Star:
		drawRGBAStar(img, spec.x, spec.y, spec.size, black)
	case detector.Convergence, detector.Divergence, detector.Amplification:
		// Draw as special symbols
		drawRGBACircle(img, spec.x, spec.y, spec.size, 2, black)
		// Add internal pattern
		for i := -spec.size/2; i <= spec.size/2; i += 5 {
			img.Set(spec.x+i, spec.y, black)
			img.Set(spec.x, spec.y+i, black)
		}
	}
}

func drawRGBACircle(img *image.RGBA, cx, cy, r, thickness int, c color.RGBA) {
	for y := cy - r - thickness; y <= cy + r + thickness; y++ {
		for x := cx - r - thickness; x <= cx + r + thickness; x++ {
			dx := x - cx
			dy := y - cy
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist >= float64(r-thickness/2) && dist <= float64(r+thickness/2) {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, c)
				}
			}
		}
	}
}

func drawRGBASquare(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	for i := 0; i <= size; i++ {
		img.Set(cx-size/2+i, cy-size/2, c)
		img.Set(cx-size/2+i, cy+size/2, c)
		img.Set(cx-size/2, cy-size/2+i, c)
		img.Set(cx+size/2, cy-size/2+i, c)
	}
}

func drawRGBATriangle(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	h := int(float64(size) * math.Sqrt(3) / 2)
	for i := 0; i <= size; i++ {
		y := cy + h/2 - i*h/size
		x1 := cx - i/2
		x2 := cx + i/2
		if x1 >= 0 && x1 < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x1, y, c)
		}
		if x2 >= 0 && x2 < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x2, y, c)
		}
	}
	for i := -size/2; i <= size/2; i++ {
		img.Set(cx+i, cy+h/2, c)
	}
}

func drawRGBAPentagon(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	for i := 0; i < 5; i++ {
		angle1 := float64(i) * 2 * math.Pi / 5 - math.Pi/2
		angle2 := float64(i+1) * 2 * math.Pi / 5 - math.Pi/2
		
		x1 := cx + int(float64(size)*math.Cos(angle1))
		y1 := cy + int(float64(size)*math.Sin(angle1))
		x2 := cx + int(float64(size)*math.Cos(angle2))
		y2 := cy + int(float64(size)*math.Sin(angle2))
		
		drawRGBALine(img, x1, y1, x2, y2, c)
	}
}

func drawRGBAStar(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	for i := 0; i < 5; i++ {
		angle1 := float64(i) * 2 * math.Pi / 5 - math.Pi/2
		angle2 := float64(i+2) * 2 * math.Pi / 5 - math.Pi/2
		
		x1 := cx + int(float64(size)*math.Cos(angle1))
		y1 := cy + int(float64(size)*math.Sin(angle1))
		x2 := cx + int(float64(size)*math.Cos(angle2))
		y2 := cy + int(float64(size)*math.Sin(angle2))
		
		drawRGBALine(img, x1, y1, x2, y2, c)
	}
}

func drawRGBALine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	
	for {
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
			img.Set(x1, y1, c)
		}
		
		if x1 == x2 && y1 == y2 {
			break
		}
		
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// Benchmark end-to-end performance
func BenchmarkEndToEndPerformance(b *testing.B) {
	complexities := []string{"simple", "medium", "complex"}
	
	// Create temporary directory for test images
	tempDir := b.TempDir()
	
	for _, complexity := range complexities {
		imagePath := filepath.Join(tempDir, fmt.Sprintf("test_%s.png", complexity))
		
		// Create test image
		if err := createPerformanceTestImage(imagePath, complexity); err != nil {
			b.Fatalf("Failed to create test image: %v", err)
		}
		
		b.Run(fmt.Sprintf("Standard_%s", complexity), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Detect symbols
				symbols, connections, err := detector.DetectSymbols(imagePath)
				if err != nil {
					b.Fatalf("Detection failed: %v", err)
				}
				
				// Parse to AST
				ast, err := parser.Parse(symbols, connections)
				if err != nil {
					b.Fatalf("Parsing failed: %v", err)
				}
				
				// Compile to code
				_, err = compiler.Compile(ast)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}
			}
		})
		
		b.Run(fmt.Sprintf("Optimized_%s", complexity), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use optimized detector
				parallelDetector := detector.NewParallelDetector(detector.Config{Debug: false})
				symbols, connections, err := parallelDetector.Detect(imagePath)
				if err != nil {
					b.Fatalf("Detection failed: %v", err)
				}
				
				// Use optimized parser
				optimizedParser := parser.NewOptimizedParser()
				ast, err := optimizedParser.Parse(symbols, connections)
				if err != nil {
					b.Fatalf("Parsing failed: %v", err)
				}
				
				// Compile to code (already optimized)
				_, err = compiler.Compile(ast)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}
			}
		})
	}
}

// Benchmark individual pipeline stages
func BenchmarkPipelineStages(b *testing.B) {
	// Create test image
	tempDir := b.TempDir()
	imagePath := filepath.Join(tempDir, "test_medium.png")
	
	if err := createPerformanceTestImage(imagePath, "medium"); err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}
	
	// Benchmark detection stage
	b.Run("Detection_Standard", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := detector.DetectSymbols(imagePath)
			if err != nil {
				b.Fatalf("Detection failed: %v", err)
			}
		}
	})
	
	b.Run("Detection_Optimized", func(b *testing.B) {
		parallelDetector := detector.NewParallelDetector(detector.Config{Debug: false})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := parallelDetector.Detect(imagePath)
			if err != nil {
				b.Fatalf("Detection failed: %v", err)
			}
		}
	})
	
	// Get symbols for parsing benchmarks
	symbols, connections, _ := detector.DetectSymbols(imagePath)
	
	// Benchmark parsing stage
	b.Run("Parsing_Standard", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := parser.Parse(symbols, connections)
			if err != nil {
				b.Fatalf("Parsing failed: %v", err)
			}
		}
	})
	
	b.Run("Parsing_Optimized", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			optimizedParser := parser.NewOptimizedParser()
			_, err := optimizedParser.Parse(symbols, connections)
			if err != nil {
				b.Fatalf("Parsing failed: %v", err)
			}
		}
	})
	
	// Get AST for compilation benchmarks
	ast, _ := parser.Parse(symbols, connections)
	
	// Benchmark compilation stage
	b.Run("Compilation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := compiler.Compile(ast)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})
}

// Benchmark memory usage
func BenchmarkMemoryUsage(b *testing.B) {
	// Create test image
	tempDir := b.TempDir()
	imagePath := filepath.Join(tempDir, "test_complex.png")
	
	if err := createPerformanceTestImage(imagePath, "complex"); err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}
	
	b.Run("Standard", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			symbols, connections, _ := detector.DetectSymbols(imagePath)
			ast, _ := parser.Parse(symbols, connections)
			_, _ = compiler.Compile(ast)
		}
	})
	
	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parallelDetector := detector.NewParallelDetector(detector.Config{Debug: false})
			symbols, connections, _ := parallelDetector.Detect(imagePath)
			optimizedParser := parser.NewOptimizedParser()
			ast, _ := optimizedParser.Parse(symbols, connections)
			_, _ = compiler.Compile(ast)
		}
	})
}