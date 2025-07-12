package detector

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"testing"
)

// createComplexBenchmarkImage creates a complex test image with many symbols
func createComplexBenchmarkImage(size int, numSymbols int) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, size, size))
	
	// Fill with white
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.Gray{255})
		}
	}

	// Draw outer circle
	centerX, centerY := size/2, size/2
	radius := size/2 - 20
	drawCircle(img, centerX, centerY, radius, 5)

	// Draw random symbols inside
	rand.Seed(42) // Fixed seed for reproducible benchmarks
	
	for i := 0; i < numSymbols; i++ {
		// Random position within circle
		angle := rand.Float64() * 2 * math.Pi
		r := rand.Float64() * float64(radius-50)
		x := centerX + int(r*math.Cos(angle))
		y := centerY + int(r*math.Sin(angle))
		
		// Random symbol type
		switch rand.Intn(6) {
		case 0: // Square
			drawSquare(img, x, y, 30)
		case 1: // Circle
			drawCircle(img, x, y, 15, 2)
		case 2: // Triangle
			drawTriangle(img, x, y, 25)
		case 3: // Star
			drawStar(img, x, y, 20)
		case 4: // Pentagon
			drawPentagon(img, x, y, 20)
		case 5: // Hexagon
			drawHexagon(img, x, y, 20)
		}
	}

	return img
}

func drawCircle(img *image.Gray, cx, cy, r, thickness int) {
	for y := cy - r - thickness; y <= cy + r + thickness; y++ {
		for x := cx - r - thickness; x <= cx + r + thickness; x++ {
			dx := x - cx
			dy := y - cy
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist >= float64(r-thickness/2) && dist <= float64(r+thickness/2) {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, color.Gray{0})
				}
			}
		}
	}
}

func drawSquare(img *image.Gray, cx, cy, size int) {
	for i := 0; i <= size; i++ {
		// Top and bottom edges
		img.Set(cx-size/2+i, cy-size/2, color.Gray{0})
		img.Set(cx-size/2+i, cy+size/2, color.Gray{0})
		// Left and right edges
		img.Set(cx-size/2, cy-size/2+i, color.Gray{0})
		img.Set(cx+size/2, cy-size/2+i, color.Gray{0})
	}
}

func drawTriangle(img *image.Gray, cx, cy, size int) {
	// Draw equilateral triangle
	h := int(float64(size) * math.Sqrt(3) / 2)
	for i := 0; i <= size; i++ {
		y := cy + h/2 - i*h/size
		x1 := cx - i/2
		x2 := cx + i/2
		if x1 >= 0 && x1 < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x1, y, color.Gray{0})
		}
		if x2 >= 0 && x2 < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x2, y, color.Gray{0})
		}
	}
	// Draw base
	for i := -size/2; i <= size/2; i++ {
		img.Set(cx+i, cy+h/2, color.Gray{0})
	}
}

func drawStar(img *image.Gray, cx, cy, size int) {
	// Draw 5-pointed star
	for i := 0; i < 5; i++ {
		angle1 := float64(i) * 2 * math.Pi / 5 - math.Pi/2
		angle2 := float64(i+2) * 2 * math.Pi / 5 - math.Pi/2
		
		x1 := cx + int(float64(size)*math.Cos(angle1))
		y1 := cy + int(float64(size)*math.Sin(angle1))
		x2 := cx + int(float64(size)*math.Cos(angle2))
		y2 := cy + int(float64(size)*math.Sin(angle2))
		
		drawLine(img, x1, y1, x2, y2)
	}
}

func drawPentagon(img *image.Gray, cx, cy, size int) {
	for i := 0; i < 5; i++ {
		angle1 := float64(i) * 2 * math.Pi / 5 - math.Pi/2
		angle2 := float64(i+1) * 2 * math.Pi / 5 - math.Pi/2
		
		x1 := cx + int(float64(size)*math.Cos(angle1))
		y1 := cy + int(float64(size)*math.Sin(angle1))
		x2 := cx + int(float64(size)*math.Cos(angle2))
		y2 := cy + int(float64(size)*math.Sin(angle2))
		
		drawLine(img, x1, y1, x2, y2)
	}
}

func drawHexagon(img *image.Gray, cx, cy, size int) {
	for i := 0; i < 6; i++ {
		angle1 := float64(i) * 2 * math.Pi / 6
		angle2 := float64(i+1) * 2 * math.Pi / 6
		
		x1 := cx + int(float64(size)*math.Cos(angle1))
		y1 := cy + int(float64(size)*math.Sin(angle1))
		x2 := cx + int(float64(size)*math.Cos(angle2))
		y2 := cy + int(float64(size)*math.Sin(angle2))
		
		drawLine(img, x1, y1, x2, y2)
	}
}

func drawLine(img *image.Gray, x1, y1, x2, y2 int) {
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
			img.Set(x1, y1, color.Gray{0})
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark standard vs parallel detector
func BenchmarkDetectorComparison(b *testing.B) {
	testCases := []struct {
		imageSize  int
		numSymbols int
	}{
		{400, 10},
		{800, 50},
		{1200, 100},
		{1600, 200},
	}

	for _, tc := range testCases {
		img := createComplexBenchmarkImage(tc.imageSize, tc.numSymbols)
		
		b.Run(fmt.Sprintf("Standard_%dx%d_%dsymbols", tc.imageSize, tc.imageSize, tc.numSymbols), func(b *testing.B) {
			detector := NewDetector(Config{Debug: false})
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				binary := detector.preprocessImage(img)
				contours := detector.findContours(binary)
				_ = detector.detectSymbolsFromContours(contours, binary)
			}
		})
		
		b.Run(fmt.Sprintf("Parallel_%dx%d_%dsymbols", tc.imageSize, tc.imageSize, tc.numSymbols), func(b *testing.B) {
			detector := NewParallelDetector(Config{Debug: false})
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				binary := detector.preprocessImage(img)
				contours := detector.findContoursParallel(binary)
				_ = detector.detectSymbolsFromContoursParallel(contours, binary)
			}
		})
	}
}

// Benchmark contour finding
func BenchmarkFindContoursComparison(b *testing.B) {
	sizes := []int{400, 800, 1200}
	
	for _, size := range sizes {
		img := createComplexBenchmarkImage(size, size/10)
		detector := NewDetector(Config{Debug: false})
		parallelDetector := NewParallelDetector(Config{Debug: false})
		binary := detector.preprocessImage(img)
		
		b.Run(fmt.Sprintf("Standard_%dx%d", size, size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.findContours(binary)
			}
		})
		
		b.Run(fmt.Sprintf("Parallel_%dx%d", size, size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = parallelDetector.findContoursParallel(binary)
			}
		})
	}
}

// Benchmark symbol detection from contours
func BenchmarkSymbolDetectionComparison(b *testing.B) {
	// Create test contours
	var contours []Contour
	for i := 0; i < 100; i++ {
		switch i % 6 {
		case 0:
			contours = append(contours, createCircleContour(50+i*10, 50+i*10, 20))
		case 1:
			contours = append(contours, createSquareContour(100+i*10, 100+i*10, 40))
		default:
			contours = append(contours, createComplexContour())
		}
	}
	
	// Calculate properties for all contours
	for i := range contours {
		contours[i].calculateProperties()
	}
	
	img := createBenchmarkImage(800)
	detector := NewDetector(Config{Debug: false})
	parallelDetector := NewParallelDetector(Config{Debug: false})
	binary := detector.preprocessImage(img)
	
	b.Run("Standard", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = detector.detectSymbolsFromContours(contours, binary)
		}
	})
	
	b.Run("Parallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = parallelDetector.detectSymbolsFromContoursParallel(contours, binary)
		}
	})
}

// Benchmark cache performance
func BenchmarkDetectorCache(b *testing.B) {
	cache := NewDetectorCache(100)
	img := createBenchmarkImage(800)
	
	// Populate cache
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("image_%d", i)
		cache.setPreprocessed(key, img)
		cache.setSymbols(key, []*Symbol{
			{Type: OuterCircle, Position: Position{X: 400, Y: 400}},
			{Type: Star, Position: Position{X: 200, Y: 200}},
		})
	}
	
	b.Run("CacheHit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("image_%d", i%50)
			_ = cache.getPreprocessed(key)
			_ = cache.getSymbols(key)
		}
	})
	
	b.Run("CacheMiss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("nonexistent_%d", i)
			_ = cache.getPreprocessed(key)
			_ = cache.getSymbols(key)
		}
	})
}