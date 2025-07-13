package grimoire_test

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
	"time"

	"github.com/ayutaz/grimoire/internal/compiler"
	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/ayutaz/grimoire/internal/parser"
)

// BenchmarkLargeScalePerformance tests performance with various symbol counts
func BenchmarkLargeScalePerformance(b *testing.B) {
	// Skip this benchmark if running in CI to avoid timeouts
	if os.Getenv("CI") != "" {
		b.Skip("Skipping large scale benchmark in CI")
	}

	// Test with different symbol counts
	symbolCounts := []int{100, 300, 500, 1000}

	for _, count := range symbolCounts {
		b.Run(fmt.Sprintf("%d_symbols", count), func(b *testing.B) {
			// Create test image with specified number of symbols
			img, _, _ := createLargeTestImage(b, count)

			// Save image to temporary file
			tmpDir := b.TempDir()
			imgPath := filepath.Join(tmpDir, fmt.Sprintf("test_%d.png", count))
			saveTestImage(b, img, imgPath)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Time the entire pipeline
				start := time.Now()

				// Detection phase
				detectedSymbols, detectedConnections, err := detector.DetectSymbols(imgPath)
				if err != nil {
					b.Fatalf("Detection failed: %v", err)
				}

				// Parsing phase
				p := parser.NewParser()
				program, err := p.Parse(detectedSymbols, detectedConnections)
				if err != nil {
					b.Fatalf("Parsing failed: %v", err)
				}

				// Compilation phase
				c := compiler.NewCompiler()
				_, err = c.Compile(program)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}

				elapsed := time.Since(start)
				b.ReportMetric(float64(elapsed.Milliseconds()), "ms/op")

				// Report individual phase metrics
				if i == 0 {
					b.Logf("Symbol count: %d, Total time: %v", count, elapsed)
				}
			}
		})
	}
}

// BenchmarkOptimizedPipeline tests the optimized versions
func BenchmarkOptimizedPipeline(b *testing.B) {
	// Skip this benchmark if running in CI to avoid timeouts
	if os.Getenv("CI") != "" {
		b.Skip("Skipping optimized pipeline benchmark in CI")
	}

	symbolCounts := []int{100, 300, 500, 1000}

	for _, count := range symbolCounts {
		b.Run(fmt.Sprintf("%d_symbols", count), func(b *testing.B) {
			// Create test image
			img, _, _ := createLargeTestImage(b, count)

			// Save image
			tmpDir := b.TempDir()
			imgPath := filepath.Join(tmpDir, fmt.Sprintf("test_opt_%d.png", count))
			saveTestImage(b, img, imgPath)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				start := time.Now()

				// Use parallel detector
				pd := detector.NewParallelDetector(detector.Config{})
				detectedSymbols, detectedConnections, err := pd.Detect(imgPath)
				if err != nil {
					b.Fatalf("Detection failed: %v", err)
				}

				// Use optimized parser
				op := parser.NewOptimizedParser()
				program, err := op.Parse(detectedSymbols, detectedConnections)
				if err != nil {
					b.Fatalf("Parsing failed: %v", err)
				}

				// Compilation
				c := compiler.NewCompiler()
				_, err = c.Compile(program)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}

				elapsed := time.Since(start)
				b.ReportMetric(float64(elapsed.Milliseconds()), "ms/op")
			}
		})
	}
}

// createLargeTestImage creates a test image with specified number of symbols
func createLargeTestImage(b *testing.B, symbolCount int) (image.Image, []*detector.Symbol, []detector.Connection) {
	b.Helper()

	// Calculate image size based on symbol count
	// Approximate grid size needed
	gridSize := int(math.Sqrt(float64(symbolCount))) + 2
	cellSize := 100
	imageSize := gridSize * cellSize

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle
	center := image.Point{X: imageSize / 2, Y: imageSize / 2}
	radius := float64(imageSize/2 - 50)
	drawCircle(img, center, int(radius), color.Black)

	// Create symbols
	symbols := make([]*detector.Symbol, 0, symbolCount)
	connections := make([]detector.Connection, 0)

	// Add outer circle symbol
	symbols = append(symbols, &detector.Symbol{
		Type:     detector.OuterCircle,
		Position: detector.Position{X: float64(center.X), Y: float64(center.Y)},
		Size:     radius,
	})

	// Distribute symbols in a grid pattern with some randomness
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	symbolTypes := []detector.SymbolType{
		detector.Square, detector.Circle, detector.Triangle,
		detector.Pentagon, detector.Hexagon, detector.Star,
	}

	for i := 1; i < symbolCount; i++ {
		// Calculate position with some randomness
		gridX := (i % gridSize) + 1
		gridY := (i / gridSize) + 1

		x := gridX*cellSize + rng.Intn(cellSize/2) - cellSize/4
		y := gridY*cellSize + rng.Intn(cellSize/2) - cellSize/4

		// Ensure within outer circle
		dx := float64(x - center.X)
		dy := float64(y - center.Y)
		if math.Sqrt(dx*dx+dy*dy) > radius*0.9 {
			continue
		}

		// Draw symbol
		symbolType := symbolTypes[rng.Intn(len(symbolTypes))]
		drawSymbol(img, image.Point{X: x, Y: y}, symbolType, 20)

		// Add to symbols list
		symbols = append(symbols, &detector.Symbol{
			Type:     symbolType,
			Position: detector.Position{X: float64(x), Y: float64(y)},
			Size:     20,
		})

		// Add some connections
		if len(symbols) > 1 && i > 1 && i < len(symbols) && rng.Float32() < 0.3 {
			connections = append(connections, detector.Connection{
				From:           symbols[len(symbols)-2],
				To:             symbols[len(symbols)-1],
				ConnectionType: "solid",
			})
			drawLine(img,
				image.Point{X: int(symbols[len(symbols)-2].Position.X), Y: int(symbols[len(symbols)-2].Position.Y)},
				image.Point{X: x, Y: y},
				color.Black)
		}
	}

	return img, symbols, connections
}

// Helper functions to draw shapes
func drawCircle(img *image.RGBA, center image.Point, radius int, c color.Color) {
	// Draw thicker circle for better detection
	for t := -2; t <= 2; t++ {
		r := radius + t
		for angle := 0.0; angle < 2*math.Pi; angle += 0.01 {
			x := center.X + int(float64(r)*math.Cos(angle))
			y := center.Y + int(float64(r)*math.Sin(angle))
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

func drawSymbol(img *image.RGBA, pos image.Point, symbolType detector.SymbolType, size int) {
	switch symbolType {
	case detector.Square:
		drawSquare(img, pos, size, color.Black)
	case detector.Circle:
		drawCircle(img, pos, size, color.Black)
	case detector.Triangle:
		drawTriangle(img, pos, size, color.Black)
	case detector.Pentagon:
		drawPolygon(img, pos, size, 5, color.Black)
	case detector.Hexagon:
		drawPolygon(img, pos, size, 6, color.Black)
	case detector.Star:
		drawStar(img, pos, size, color.Black)
	}
}

func drawSquare(img *image.RGBA, center image.Point, size int, c color.Color) {
	for x := center.X - size; x <= center.X+size; x++ {
		img.Set(x, center.Y-size, c)
		img.Set(x, center.Y+size, c)
	}
	for y := center.Y - size; y <= center.Y+size; y++ {
		img.Set(center.X-size, y, c)
		img.Set(center.X+size, y, c)
	}
}

func drawTriangle(img *image.RGBA, center image.Point, size int, c color.Color) {
	// Simple equilateral triangle
	points := []image.Point{
		{X: center.X, Y: center.Y - size},
		{X: center.X - size, Y: center.Y + size/2},
		{X: center.X + size, Y: center.Y + size/2},
	}

	// Draw lines between points
	drawLine(img, points[0], points[1], c)
	drawLine(img, points[1], points[2], c)
	drawLine(img, points[2], points[0], c)
}

func drawPolygon(img *image.RGBA, center image.Point, size int, sides int, c color.Color) {
	angleStep := 2 * math.Pi / float64(sides)
	var prevPoint image.Point

	for i := 0; i <= sides; i++ {
		angle := float64(i)*angleStep - math.Pi/2
		x := center.X + int(float64(size)*math.Cos(angle))
		y := center.Y + int(float64(size)*math.Sin(angle))

		point := image.Point{X: x, Y: y}
		if i > 0 {
			drawLine(img, prevPoint, point, c)
		}
		prevPoint = point
	}
}

func drawStar(img *image.RGBA, center image.Point, size int, c color.Color) {
	// 5-pointed star
	outerRadius := float64(size)
	innerRadius := outerRadius * 0.4

	for i := 0; i < 10; i++ {
		angle := float64(i)*math.Pi/5 - math.Pi/2
		radius := outerRadius
		if i%2 == 1 {
			radius = innerRadius
		}

		x := center.X + int(radius*math.Cos(angle))
		y := center.Y + int(radius*math.Sin(angle))

		if i > 0 {
			prevAngle := float64(i-1)*math.Pi/5 - math.Pi/2
			prevRadius := outerRadius
			if (i-1)%2 == 1 {
				prevRadius = innerRadius
			}
			prevX := center.X + int(prevRadius*math.Cos(prevAngle))
			prevY := center.Y + int(prevRadius*math.Sin(prevAngle))

			drawLine(img, image.Point{X: prevX, Y: prevY}, image.Point{X: x, Y: y}, c)
		}
	}

	// Close the star
	angle0 := -math.Pi / 2
	x0 := center.X + int(outerRadius*math.Cos(angle0))
	y0 := center.Y + int(outerRadius*math.Sin(angle0))
	angle9 := 9*math.Pi/5 - math.Pi/2
	x9 := center.X + int(innerRadius*math.Cos(angle9))
	y9 := center.Y + int(innerRadius*math.Sin(angle9))
	drawLine(img, image.Point{X: x9, Y: y9}, image.Point{X: x0, Y: y0}, c)
}

func drawLine(img *image.RGBA, from, to image.Point, c color.Color) {
	// Simple line drawing using Bresenham's algorithm
	dx := abs(to.X - from.X)
	dy := abs(to.Y - from.Y)
	sx := 1
	if from.X > to.X {
		sx = -1
	}
	sy := 1
	if from.Y > to.Y {
		sy = -1
	}
	err := dx - dy

	x, y := from.X, from.Y
	for {
		img.Set(x, y, c)

		if x == to.X && y == to.Y {
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func saveTestImage(b *testing.B, img image.Image, path string) {
	b.Helper()

	file, err := os.Create(path)
	if err != nil {
		b.Fatalf("Failed to create image file: %v", err)
	}
	defer file.Close()

	if err := encodePNG(file, img); err != nil {
		b.Fatalf("Failed to encode PNG: %v", err)
	}
}

// encodePNG encodes an image as PNG
func encodePNG(file *os.File, img image.Image) error {
	return png.Encode(file, img)
}

// BenchmarkCIPerformance is a lightweight benchmark for CI environments
func BenchmarkCIPerformance(b *testing.B) {
	// Create a simple test image
	imgSize := 400
	img := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle
	center := image.Point{X: imgSize / 2, Y: imgSize / 2}
	radius := imgSize/2 - 20
	drawCircle(img, center, radius, color.Black)

	// Draw a few symbols
	drawSquare(img, image.Point{X: 150, Y: 150}, 20, color.Black)
	drawCircle(img, image.Point{X: 250, Y: 150}, 20, color.Black)
	drawTriangle(img, image.Point{X: 200, Y: 250}, 20, color.Black)

	// Save image
	tmpDir := b.TempDir()
	imgPath := filepath.Join(tmpDir, "ci_test_simple.png")
	saveTestImage(b, img, imgPath)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Test basic detection only
		symbols, _, err := detector.DetectSymbols(imgPath)
		if err != nil {
			b.Fatalf("Detection failed: %v", err)
		}

		// Just verify we found some symbols
		if len(symbols) < 2 { // At least outer circle and one symbol
			b.Fatalf("Expected at least 2 symbols, got %d", len(symbols))
		}
	}
}
