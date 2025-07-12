package detector

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkDetectorComparison compares standard, parallel, and V2 detectors
func BenchmarkDetectorComparison(b *testing.B) {
	testCases := []struct {
		numSymbols int
		density    float64
	}{
		{50, 0.8},
		{100, 0.7},
		{200, 0.6},
		{500, 0.5},
		{1000, 0.4},
	}

	for _, tc := range testCases {
		// Create test image
		imgPath := createOptimizedBenchmarkImage(b, tc.numSymbols, tc.density)

		b.Run(fmt.Sprintf("Standard_%dsymbols", tc.numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				detector := NewDetector(Config{})
				_, _, _ = detector.Detect(imgPath)
			}
		})

		b.Run(fmt.Sprintf("Parallel_%dsymbols", tc.numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				detector := NewParallelDetector(Config{})
				_, _, _ = detector.Detect(imgPath)
			}
		})

		b.Run(fmt.Sprintf("ParallelV2_%dsymbols", tc.numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				detector := NewParallelDetectorV2(Config{})
				_, _, _ = detector.Detect(imgPath)
				detector.Cleanup()
			}
		})

	}
}

// BenchmarkContourDetection benchmarks only the contour detection phase
func BenchmarkContourDetection(b *testing.B) {
	sizes := []int{500, 1000, 2000, 4000}

	for _, size := range sizes {
		// Create binary image
		binary := createBinaryImage(size, size, 0.05)

		b.Run(fmt.Sprintf("Standard_%dx%d", size, size), func(b *testing.B) {
			detector := NewDetector(Config{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.findContours(binary)
			}
		})

		b.Run(fmt.Sprintf("ParallelV2_%dx%d", size, size), func(b *testing.B) {
			detector := NewParallelDetectorV2(Config{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.findContoursOptimized(binary)
			}
		})
	}
}

// BenchmarkSymbolDetection benchmarks symbol detection from contours
func BenchmarkSymbolDetection(b *testing.B) {
	contourCounts := []int{100, 500, 1000, 2000}

	for _, count := range contourCounts {
		contours := generateTestContours(count)
		binary := createBinaryImage(2000, 2000, 0.01)

		b.Run(fmt.Sprintf("Standard_%dcontours", count), func(b *testing.B) {
			detector := NewDetector(Config{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.detectSymbolsFromContours(contours, binary)
			}
		})

		b.Run(fmt.Sprintf("ParallelV2_%dcontours", count), func(b *testing.B) {
			detector := NewParallelDetectorV2(Config{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.detectSymbolsOptimized(contours, binary)
			}
		})
	}
}

// BenchmarkConnectionDetection benchmarks connection detection
func BenchmarkConnectionDetection(b *testing.B) {
	symbolCounts := []int{50, 100, 200, 500}

	for _, count := range symbolCounts {
		symbols := generateTestSymbols(count)
		binary := createBinaryImageWithConnections(2000, 2000, symbols)

		b.Run(fmt.Sprintf("Standard_%dsymbols", count), func(b *testing.B) {
			detector := NewDetector(Config{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.detectConnections(binary, symbols)
			}
		})

		b.Run(fmt.Sprintf("Improved_%dsymbols", count), func(b *testing.B) {
			detector := NewDetector(Config{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.improvedDetectConnections(binary, symbols)
			}
		})

		b.Run(fmt.Sprintf("ParallelV2_%dsymbols", count), func(b *testing.B) {
			detector := NewParallelDetectorV2(Config{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.detectConnectionsOptimized(binary, symbols)
			}
		})
	}
}

// BenchmarkMemoryAllocation compares memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	contours := generateTestContours(100)
	binary := createBinaryImage(2000, 2000, 0.01)

	b.Run("Standard", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			detector := NewDetector(Config{})
			_ = detector.detectSymbolsFromContours(contours, binary)
		}
	})

	b.Run("ParallelV2_WithPooling", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		detector := NewParallelDetectorV2(Config{})
		for i := 0; i < b.N; i++ {
			_ = detector.detectSymbolsOptimized(contours, binary)
		}
		detector.Cleanup()
	})
}

// Helper functions for benchmark tests

func createOptimizedBenchmarkImage(b *testing.B, numSymbols int, density float64) string {
	b.Helper()

	// Calculate image size
	gridSize := int(math.Sqrt(float64(numSymbols))) + 2
	cellSize := 80
	imageSize := gridSize * cellSize

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle
	center := image.Point{X: imageSize / 2, Y: imageSize / 2}
	radius := imageSize/2 - 50
	drawCircleThick(img, center.X, center.Y, radius, 3)

	// Random seed for reproducibility
	rand.Seed(42)

	// Draw symbols
	for i := 0; i < numSymbols; i++ {
		x := rand.Intn(imageSize-100) + 50
		y := rand.Intn(imageSize-100) + 50

		// Check if within outer circle
		dx := x - center.X
		dy := y - center.Y
		if float64(dx*dx+dy*dy) > float64(radius*radius)*0.8 {
			continue
		}

		// Draw random symbol
		symbolType := rand.Intn(6)
		switch symbolType {
		case 0:
			drawSquareThick(img, x, y, 20, 2)
		case 1:
			drawCircleThick(img, x, y, 20, 2)
		case 2:
			drawTriangleThick(img, x, y, 20, 2)
		case 3:
			drawPentagonThick(img, x, y, 20, 2)
		case 4:
			drawHexagonThick(img, x, y, 20, 2)
		case 5:
			drawStarThick(img, x, y, 20, 2)
		}

		// Add pattern inside some symbols
		if rand.Float64() < 0.3 {
			drawCircleThick(img, x, y, 5, 1)
		}
	}

	// Draw connections based on density
	for i := 0; i < int(float64(numSymbols)*density); i++ {
		x1 := rand.Intn(imageSize-100) + 50
		y1 := rand.Intn(imageSize-100) + 50
		x2 := x1 + rand.Intn(200) - 100
		y2 := y1 + rand.Intn(200) - 100

		drawLineThick(img, x1, y1, x2, y2, 1)
	}

	// Save to temporary file
	tmpDir := b.TempDir()
	imgPath := filepath.Join(tmpDir, fmt.Sprintf("benchmark_%d.png", numSymbols))

	file, err := os.Create(imgPath)
	if err != nil {
		b.Fatalf("Failed to create image file: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		b.Fatalf("Failed to encode PNG: %v", err)
	}

	return imgPath
}

func createBinaryImage(width, height int, fillRatio float64) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Fill with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{255})
		}
	}

	// Add random black pixels
	numPixels := int(float64(width*height) * fillRatio)
	for i := 0; i < numPixels; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		img.Set(x, y, color.Gray{0})
	}

	return img
}

func createBinaryImageWithConnections(width, height int, symbols []*Symbol) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Fill with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{255})
		}
	}

	// Draw connections between nearby symbols
	for i, sym1 := range symbols {
		for j := i + 1; j < len(symbols); j++ {
			sym2 := symbols[j]
			dx := sym2.Position.X - sym1.Position.X
			dy := sym2.Position.Y - sym1.Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < 200 && rand.Float64() < 0.3 {
				// Draw line
				drawBinaryLine(img, 
					int(sym1.Position.X), int(sym1.Position.Y),
					int(sym2.Position.X), int(sym2.Position.Y))
			}
		}
	}

	return img
}

func generateTestContours(count int) []Contour {
	contours := make([]Contour, count)

	for i := 0; i < count; i++ {
		centerX := rand.Intn(1800) + 100
		centerY := rand.Intn(1800) + 100
		radius := rand.Intn(50) + 10

		points := make([]image.Point, 0, 100)
		for angle := 0.0; angle < 2*math.Pi; angle += 0.1 {
			x := centerX + int(float64(radius)*math.Cos(angle))
			y := centerY + int(float64(radius)*math.Sin(angle))
			points = append(points, image.Point{X: x, Y: y})
		}

		contours[i] = Contour{
			Points: points,
			Center: image.Point{X: centerX, Y: centerY},
			Area:   math.Pi * float64(radius) * float64(radius),
			Perimeter: 2 * math.Pi * float64(radius),
			Circularity: 0.8 + rand.Float64()*0.2,
		}
	}

	return contours
}

func generateTestSymbols(count int) []*Symbol {
	symbols := make([]*Symbol, count)
	symbolTypes := []SymbolType{
		Square, Circle, Triangle, Pentagon, Hexagon, Star,
	}

	for i := 0; i < count; i++ {
		symbols[i] = &Symbol{
			Type:     symbolTypes[rand.Intn(len(symbolTypes))],
			Position: Position{
				X: float64(rand.Intn(1800) + 100),
				Y: float64(rand.Intn(1800) + 100),
			},
			Size:       float64(rand.Intn(30) + 10),
			Confidence: 0.8 + rand.Float64()*0.2,
		}
	}

	return symbols
}

// Drawing helpers with thickness
func drawCircleThick(img *image.RGBA, cx, cy, radius, thickness int) {
	for t := 0; t < thickness; t++ {
		r := radius - t
		if r < 0 {
			break
		}
		for angle := 0.0; angle < 2*math.Pi; angle += 0.01 {
			x := cx + int(float64(r)*math.Cos(angle))
			y := cy + int(float64(r)*math.Sin(angle))
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, color.Black)
			}
		}
	}
}

func drawSquareThick(img *image.RGBA, cx, cy, size, thickness int) {
	for t := 0; t < thickness; t++ {
		s := size - t
		if s < 0 {
			break
		}
		// Top and bottom
		for x := cx - s; x <= cx + s; x++ {
			if x >= 0 && x < img.Bounds().Dx() {
				if cy-s >= 0 {
					img.Set(x, cy-s, color.Black)
				}
				if cy+s < img.Bounds().Dy() {
					img.Set(x, cy+s, color.Black)
				}
			}
		}
		// Left and right
		for y := cy - s; y <= cy + s; y++ {
			if y >= 0 && y < img.Bounds().Dy() {
				if cx-s >= 0 {
					img.Set(cx-s, y, color.Black)
				}
				if cx+s < img.Bounds().Dx() {
					img.Set(cx+s, y, color.Black)
				}
			}
		}
	}
}

func drawTriangleThick(img *image.RGBA, cx, cy, size, thickness int) {
	height := int(float64(size) * math.Sqrt(3) / 2)
	
	for t := 0; t < thickness; t++ {
		// Calculate vertices
		x1, y1 := cx, cy-height+t
		x2, y2 := cx-size+t, cy+height/2-t
		x3, y3 := cx+size-t, cy+height/2-t
		
		drawLineThick(img, x1, y1, x2, y2, 1)
		drawLineThick(img, x2, y2, x3, y3, 1)
		drawLineThick(img, x3, y3, x1, y1, 1)
	}
}

func drawPentagonThick(img *image.RGBA, cx, cy, size, thickness int) {
	drawPolygonThick(img, cx, cy, size, 5, thickness)
}

func drawHexagonThick(img *image.RGBA, cx, cy, size, thickness int) {
	drawPolygonThick(img, cx, cy, size, 6, thickness)
}

func drawPolygonThick(img *image.RGBA, cx, cy, size, sides, thickness int) {
	angleStep := 2 * math.Pi / float64(sides)
	
	for t := 0; t < thickness; t++ {
		s := size - t
		if s < 0 {
			break
		}
		
		var prevX, prevY int
		for i := 0; i <= sides; i++ {
			angle := float64(i)*angleStep - math.Pi/2
			x := cx + int(float64(s)*math.Cos(angle))
			y := cy + int(float64(s)*math.Sin(angle))
			
			if i > 0 {
				drawLineThick(img, prevX, prevY, x, y, 1)
			}
			prevX, prevY = x, y
		}
	}
}

func drawStarThick(img *image.RGBA, cx, cy, size, thickness int) {
	outerRadius := float64(size)
	innerRadius := outerRadius * 0.4
	
	for t := 0; t < thickness; t++ {
		outer := outerRadius - float64(t)
		inner := innerRadius - float64(t)*0.4
		if outer < 0 || inner < 0 {
			break
		}
		
		for i := 0; i < 10; i++ {
			angle := float64(i) * math.Pi / 5 - math.Pi/2
			radius := outer
			if i%2 == 1 {
				radius = inner
			}
			
			x := cx + int(radius*math.Cos(angle))
			y := cy + int(radius*math.Sin(angle))
			
			if i > 0 {
				prevAngle := float64(i-1) * math.Pi / 5 - math.Pi/2
				prevRadius := outer
				if (i-1)%2 == 1 {
					prevRadius = inner
				}
				prevX := cx + int(prevRadius*math.Cos(prevAngle))
				prevY := cy + int(prevRadius*math.Sin(prevAngle))
				
				drawLineThick(img, prevX, prevY, x, y, 1)
			}
			
			if i == 9 {
				// Close the star
				firstAngle := -math.Pi/2
				firstX := cx + int(outer*math.Cos(firstAngle))
				firstY := cy + int(outer*math.Sin(firstAngle))
				drawLineThick(img, x, y, firstX, firstY, 1)
			}
		}
	}
}

func drawLineThick(img *image.RGBA, x1, y1, x2, y2, thickness int) {
	// Simple line drawing
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	
	x, y := x1, y1
	for {
		// Draw with thickness
		for tx := -thickness; tx <= thickness; tx++ {
			for ty := -thickness; ty <= thickness; ty++ {
				px, py := x+tx, y+ty
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, color.Black)
				}
			}
		}
		
		if x == x2 && y == y2 {
			break
		}
		
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func drawBinaryLine(img *image.Gray, x1, y1, x2, y2 int) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	
	x, y := x1, y1
	for {
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, color.Gray{0})
		}
		
		if x == x2 && y == y2 {
			break
		}
		
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

