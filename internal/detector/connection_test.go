package detector

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"testing"
)

// TestDiagonalLineDetection tests detection of 45° and 135° diagonal lines
func TestDiagonalLineDetection(t *testing.T) {
	tests := []struct {
		name        string
		angle       float64 // in radians
		description string
	}{
		{"45_degree", math.Pi / 4, "45° diagonal line (top-left to bottom-right)"},
		{"135_degree", 3 * math.Pi / 4, "135° diagonal line (top-right to bottom-left)"},
		{"neg_45_degree", -math.Pi / 4, "-45° diagonal line (bottom-left to top-right)"},
		{"neg_135_degree", -3 * math.Pi / 4, "-135° diagonal line (bottom-right to top-left)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			img := createDiagonalTestImage(tt.angle)

			// Save test image for debugging
			if tt.name == "135_degree" {
				debugFile, _ := os.Create("test_135_degree_debug.png")
				if debugFile != nil {
					png.Encode(debugFile, img)
					debugFile.Close()
					t.Logf("Saved debug image to test_135_degree_debug.png")
				}
			}

			// Save to temporary file
			tmpfile, err := os.CreateTemp("", "diagonal_test_*.png")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if encodeErr := png.Encode(tmpfile, img); encodeErr != nil {
				t.Fatal(encodeErr)
			}
			tmpfile.Close()

			// Run detection
			d := NewDetector(Config{})
			symbols, connections, err := d.Detect(tmpfile.Name())
			if err != nil {
				t.Fatalf("Detection failed: %v", err)
			}

			// Debug: print detected symbols and connections
			t.Logf("Test: %s", tt.name)
			t.Logf("Detected %d symbols and %d connections", len(symbols), len(connections))
			for i, sym := range symbols {
				t.Logf("Symbol[%d]: %s at (%.0f, %.0f)", i, sym.Type, sym.Position.X, sym.Position.Y)
			}
			for i, conn := range connections {
				dx := conn.To.Position.X - conn.From.Position.X
				dy := conn.To.Position.Y - conn.From.Position.Y
				angle := math.Atan2(dy, dx)
				t.Logf("Connection[%d]: from (%.0f, %.0f) to (%.0f, %.0f), angle: %.2f radians (%.1f degrees)",
					i, conn.From.Position.X, conn.From.Position.Y, conn.To.Position.X, conn.To.Position.Y,
					angle, angle*180/math.Pi)
			}

			// Verify we found the expected symbols
			if len(symbols) < 3 { // outer circle + 2 squares minimum
				t.Errorf("Expected at least 3 symbols, got %d", len(symbols))
			}

			// Count how many squares were detected
			squareCount := 0
			for _, sym := range symbols {
				if sym.Type == Square {
					squareCount++
				}
			}
			if squareCount != 2 {
				t.Logf("Warning: Expected 2 squares, but found %d", squareCount)
			}

			// Verify we found diagonal connection
			foundDiagonal := false
			var foundSquares []*Symbol
			for _, sym := range symbols {
				if sym.Type == Square {
					foundSquares = append(foundSquares, sym)
				}
			}

			// Look for connections between the two squares
			if len(foundSquares) >= 2 {
				// For debugging, log expected square positions
				expectedDistance := 120.0
				expectedX1 := 200 - int(expectedDistance*math.Cos(tt.angle))
				expectedY1 := 200 - int(expectedDistance*math.Sin(tt.angle))
				expectedX2 := 200 + int(expectedDistance*math.Cos(tt.angle))
				expectedY2 := 200 + int(expectedDistance*math.Sin(tt.angle))
				t.Logf("Expected square positions: (%d, %d) and (%d, %d)", expectedX1, expectedY1, expectedX2, expectedY2)

				for _, conn := range connections {
					// Check if connection is between two squares
					fromSquare := false
					toSquare := false
					for _, sq := range foundSquares {
						if conn.From == sq {
							fromSquare = true
						}
						if conn.To == sq {
							toSquare = true
						}
					}

					if fromSquare && toSquare {
						dx := conn.To.Position.X - conn.From.Position.X
						dy := conn.To.Position.Y - conn.From.Position.Y

						// Calculate connection angle
						connAngle := math.Atan2(dy, dx)

						// Check if angle matches expected (with tolerance)
						angleDiff := math.Abs(normalizeAngle(connAngle - tt.angle))
						reverseDiff := math.Abs(normalizeAngle(connAngle - tt.angle + math.Pi))

						// Check both directions (A->B or B->A)
						if angleDiff < math.Pi/8 || reverseDiff < math.Pi/8 { // 22.5° tolerance
							foundDiagonal = true
							t.Logf("Found diagonal connection between squares with angle %.2f radians (expected %.2f)", connAngle, tt.angle)
							break
						}
					}
				}
			}

			if !foundDiagonal {
				t.Errorf("Failed to detect %s", tt.description)
			}
		})
	}
}

// createDiagonalTestImage creates a test image with two squares connected by a diagonal line
func createDiagonalTestImage(angle float64) image.Image {
	size := 400
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle
	center := image.Point{X: size / 2, Y: size / 2}
	drawTestCircle(img, center, size/2-20, color.Black)

	// Draw single circle (function)
	drawTestCircle(img, image.Point{X: 50, Y: 50}, 25, color.Black)

	// Calculate positions for squares based on angle
	// Use larger distance and ensure squares are well separated
	distance := 120.0
	x1 := center.X - int(distance*math.Cos(angle))
	y1 := center.Y - int(distance*math.Sin(angle))
	x2 := center.X + int(distance*math.Cos(angle))
	y2 := center.Y + int(distance*math.Sin(angle))

	// Draw squares first
	drawTestSquare(img, image.Point{X: x1, Y: y1}, 15, color.Black)
	drawTestSquare(img, image.Point{X: x2, Y: y2}, 15, color.Black)

	// Draw diagonal line between squares (thinner line)
	drawThinTestLine(img, image.Point{X: x1, Y: y1}, image.Point{X: x2, Y: y2}, color.Black)

	return img
}

// Helper functions for drawing
func drawTestCircle(img *image.RGBA, center image.Point, radius int, c color.Color) {
	for angle := 0.0; angle < 2*math.Pi; angle += 0.01 {
		x := center.X + int(float64(radius)*math.Cos(angle))
		y := center.Y + int(float64(radius)*math.Sin(angle))

		// Draw with thickness
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
}

func drawTestSquare(img *image.RGBA, center image.Point, halfSize int, c color.Color) {
	for x := center.X - halfSize; x <= center.X+halfSize; x++ {
		for y := center.Y - halfSize; y <= center.Y+halfSize; y++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}


func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// normalizeAngle normalizes angle to [-π, π]
func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// drawThinTestLine draws a line without thickness
func drawThinTestLine(img *image.RGBA, from, to image.Point, c color.Color) {
	// Bresenham's line algorithm
	dx := absInt(to.X - from.X)
	dy := absInt(to.Y - from.Y)
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
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}

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
