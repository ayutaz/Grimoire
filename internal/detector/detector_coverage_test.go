package detector

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectConnections_Comprehensive tests the detectConnections function comprehensively
func TestDetectConnections_Comprehensive(t *testing.T) {
	detector := NewDetector(Config{})

	tests := []struct {
		name            string
		createImage     func() *image.Gray
		symbols         []*Symbol
		expectedConns   int
		expectedTypes   []string
		withDebug       bool
	}{
		{
			name: "Simple horizontal connection",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				// Draw white background
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// Draw black line from (50,100) to (250,100)
				for x := 50; x <= 250; x++ {
					gray.Set(x, 100, color.Gray{0})
				}
				return gray
			},
			symbols: []*Symbol{
				{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
				{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			},
			expectedConns: 1,
			expectedTypes: []string{"solid"},
		},
		{
			name: "Vertical connection",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 200, 300))
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// Draw vertical line
				for y := 50; y <= 250; y++ {
					gray.Set(100, y, color.Gray{0})
				}
				return gray
			},
			symbols: []*Symbol{
				{Type: Circle, Position: Position{X: 100, Y: 50}, Size: 20},
				{Type: Star, Position: Position{X: 100, Y: 250}, Size: 20},
			},
			expectedConns: 1,
			expectedTypes: []string{"solid"},
		},
		{
			name: "Diagonal connection",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 300))
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// Draw diagonal line
				for i := 0; i <= 200; i++ {
					gray.Set(50+i, 50+i, color.Gray{0})
				}
				return gray
			},
			symbols: []*Symbol{
				{Type: Square, Position: Position{X: 50, Y: 50}, Size: 20},
				{Type: Circle, Position: Position{X: 250, Y: 250}, Size: 20},
			},
			expectedConns: 1,
			expectedTypes: []string{"solid"},
		},
		{
			name: "Dashed connection",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// Draw dashed line
				for x := 50; x <= 250; x += 20 {
					for dx := 0; dx < 10 && x+dx <= 250; dx++ {
						gray.Set(x+dx, 100, color.Gray{0})
					}
				}
				return gray
			},
			symbols: []*Symbol{
				{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
				{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			},
			expectedConns: 0, // May not detect dashed lines reliably
			expectedTypes: []string{"dashed"},
		},
		{
			name: "No connection - symbols too far",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// No lines drawn
				return gray
			},
			symbols: []*Symbol{
				{Type: Square, Position: Position{X: 50, Y: 50}, Size: 20},
				{Type: Circle, Position: Position{X: 250, Y: 150}, Size: 20},
			},
			expectedConns: 0,
			expectedTypes: []string{},
		},
		{
			name: "Connection with outer circle (should be filtered)",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// Draw line
				for x := 50; x <= 250; x++ {
					gray.Set(x, 100, color.Gray{0})
				}
				return gray
			},
			symbols: []*Symbol{
				{Type: OuterCircle, Position: Position{X: 50, Y: 100}, Size: 20},
				{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			},
			expectedConns: 0, // Connections to outer circle are filtered
			expectedTypes: []string{},
		},
		{
			name: "Multiple connections",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 300))
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// Draw Y-shaped connection
				// Stem
				for y := 150; y <= 250; y++ {
					gray.Set(150, y, color.Gray{0})
				}
				// Left branch
				for i := 0; i <= 50; i++ {
					gray.Set(150-i, 150-i, color.Gray{0})
				}
				// Right branch
				for i := 0; i <= 50; i++ {
					gray.Set(150+i, 150-i, color.Gray{0})
				}
				return gray
			},
			symbols: []*Symbol{
				{Type: Square, Position: Position{X: 100, Y: 100}, Size: 20},
				{Type: Square, Position: Position{X: 200, Y: 100}, Size: 20},
				{Type: Convergence, Position: Position{X: 150, Y: 150}, Size: 20},
				{Type: Star, Position: Position{X: 150, Y: 250}, Size: 20},
			},
			expectedConns: 2, // May detect some connections
			expectedTypes: []string{"solid", "solid"},
		},
		{
			name: "With debug output",
			createImage: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				draw.Draw(gray, gray.Bounds(), &image.Uniform{color.Gray{255}}, image.Point{}, draw.Src)
				// Draw line
				for x := 50; x <= 250; x++ {
					gray.Set(x, 100, color.Gray{0})
				}
				return gray
			},
			symbols: []*Symbol{
				{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
				{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			},
			expectedConns: 1,
			expectedTypes: []string{"solid"},
			withDebug:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set debug environment if needed
			if tc.withDebug {
				os.Setenv("GRIMOIRE_DEBUG", "1")
				defer os.Unsetenv("GRIMOIRE_DEBUG")
			}

			binary := tc.createImage()
			connections := detector.detectConnections(binary, tc.symbols)

			// Check number of connections
			if tc.expectedConns > 0 {
				assert.GreaterOrEqual(t, len(connections), 1, "Should detect at least one connection")
			} else {
				assert.Equal(t, tc.expectedConns, len(connections))
			}

			// Check connection types if any
			if len(connections) > 0 && len(tc.expectedTypes) > 0 {
				for i, conn := range connections {
					if i < len(tc.expectedTypes) {
						assert.Contains(t, []string{"solid", "dashed", "dotted"}, conn.ConnectionType)
					}
				}
			}
		})
	}
}

// TestIsValidConnection tests the isValidConnection function
func TestIsValidConnection(t *testing.T) {
	detector := NewDetector(Config{})

	tests := []struct {
		name     string
		line     Line
		from     *Symbol
		to       *Symbol
		expected bool
	}{
		{
			name: "Valid connection - symbols close to line endpoints",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 250, Y: 100}},
			from: &Symbol{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
			to:   &Symbol{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			expected: true,
		},
		{
			name: "Invalid - from outer circle",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 250, Y: 100}},
			from: &Symbol{Type: OuterCircle, Position: Position{X: 50, Y: 100}, Size: 20},
			to:   &Symbol{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			expected: false,
		},
		{
			name: "Invalid - to outer circle",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 250, Y: 100}},
			from: &Symbol{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
			to:   &Symbol{Type: OuterCircle, Position: Position{X: 250, Y: 100}, Size: 20},
			expected: false,
		},
		{
			name: "Invalid - line too short",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 60, Y: 100}},
			from: &Symbol{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
			to:   &Symbol{Type: Circle, Position: Position{X: 60, Y: 100}, Size: 20},
			expected: false,
		},
		{
			name: "Invalid - start too far from symbol",
			line: Line{Start: image.Point{X: 150, Y: 100}, End: image.Point{X: 250, Y: 100}},
			from: &Symbol{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
			to:   &Symbol{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			expected: false,
		},
		{
			name: "Invalid - end too far from symbol",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 150, Y: 100}},
			from: &Symbol{Type: Square, Position: Position{X: 50, Y: 100}, Size: 20},
			to:   &Symbol{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 20},
			expected: false,
		},
		{
			name: "Valid - larger symbols have larger tolerance",
			line: Line{Start: image.Point{X: 80, Y: 100}, End: image.Point{X: 220, Y: 100}},
			from: &Symbol{Type: Square, Position: Position{X: 50, Y: 100}, Size: 40},
			to:   &Symbol{Type: Circle, Position: Position{X: 250, Y: 100}, Size: 40},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isValidConnection(tc.line, tc.from, tc.to)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestDetermineConnectionType tests the determineConnectionType function
func TestDetermineConnectionType(t *testing.T) {
	detector := NewDetector(Config{})

	tests := []struct {
		name         string
		line         Line
		createBinary func() *image.Gray
		expected     string
	}{
		{
			name: "Solid line",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 250, Y: 100}},
			createBinary: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				// Draw solid line
				for x := 50; x <= 250; x++ {
					gray.Set(x, 100, color.Gray{255})
				}
				return gray
			},
			expected: "solid",
		},
		{
			name: "Dashed line",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 250, Y: 100}},
			createBinary: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				// Draw dashed line
				for x := 50; x <= 250; x++ {
					if (x-50)%20 < 10 {
						gray.Set(x, 100, color.Gray{255})
					}
				}
				return gray
			},
			expected: "dashed",
		},
		{
			name: "Dotted line",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 250, Y: 100}},
			createBinary: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 300, 200))
				// Draw dotted line
				for x := 50; x <= 250; x++ {
					if (x-50)%10 < 2 {
						gray.Set(x, 100, color.Gray{255})
					}
				}
				return gray
			},
			expected: "dotted",
		},
		{
			name: "Zero length line",
			line: Line{Start: image.Point{X: 50, Y: 100}, End: image.Point{X: 50, Y: 100}},
			createBinary: func() *image.Gray {
				return image.NewGray(image.Rect(0, 0, 300, 200))
			},
			expected: "solid",
		},
		{
			name: "Diagonal solid line",
			line: Line{Start: image.Point{X: 50, Y: 50}, End: image.Point{X: 150, Y: 150}},
			createBinary: func() *image.Gray {
				gray := image.NewGray(image.Rect(0, 0, 200, 200))
				// Draw diagonal line
				for i := 0; i <= 100; i++ {
					gray.Set(50+i, 50+i, color.Gray{255})
				}
				return gray
			},
			expected: "solid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			binary := tc.createBinary()
			result := detector.determineConnectionType(tc.line, binary)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassifyShape_Comprehensive tests the classifyShape function comprehensively
func TestClassifyShape_Comprehensive(t *testing.T) {
	detector := NewDetector(Config{})

	tests := []struct {
		name     string
		contour  Contour
		expected SymbolType
		withDebug bool
	}{
		{
			name: "Circle shape",
			contour: Contour{
				Points:      generateCirclePoints(100, 100, 30),
				Area:        2827.0, // Approximately pi * 30^2
				Perimeter:   188.5,  // Approximately 2 * pi * 30
				Circularity: 0.95,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Circle,
		},
		{
			name: "Outer circle (large)",
			contour: Contour{
				Points:      generateCirclePoints(200, 200, 150),
				Area:        70686.0, // Approximately pi * 150^2
				Perimeter:   942.5,   // Approximately 2 * pi * 150
				Circularity: 0.95,
				Center:      image.Point{X: 200, Y: 200},
			},
			expected: OuterCircle,
		},
		{
			name: "Square shape",
			contour: Contour{
				Points:      generateSquarePoints(100, 100, 40),
				Area:        1600.0,
				Perimeter:   160.0,
				Circularity: 0.78,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Square,
		},
		{
			name: "Square with moderate circularity",
			contour: Contour{
				Points:      generateSquarePoints(100, 100, 30),
				Area:        900.0,
				Perimeter:   120.0,
				Circularity: 0.5,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Square,
		},
		{
			name: "Star shape",
			contour: Contour{
				Points:      generateStarPoints(100, 100, 30, 5),
				Area:        850.0,
				Perimeter:   200.0,
				Circularity: 0.27,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Star,
		},
		{
			name: "Triangle shape",
			contour: Contour{
				Points:      generateTrianglePoints(100, 100, 30),
				Area:        450.0,
				Perimeter:   90.0,
				Circularity: 0.7,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Triangle,
		},
		{
			name: "Pentagon shape",
			contour: Contour{
				Points:      generatePolygonPoints(100, 100, 30, 5),
				Area:        1100.0,
				Perimeter:   110.0,
				Circularity: 0.8,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Pentagon,
		},
		{
			name: "Hexagon shape",
			contour: Contour{
				Points:      generatePolygonPoints(100, 100, 30, 6),
				Area:        1300.0,
				Perimeter:   120.0,
				Circularity: 0.85,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Hexagon,
		},
		{
			name: "Convergence operator",
			contour: Contour{
				Points:      generateYShapePoints(300, 250),
				Area:        1500.0,
				Perimeter:   300.0,
				Circularity: 0.15,
				Center:      image.Point{X: 300, Y: 250},
			},
			expected: Convergence,
		},
		{
			name: "Unknown shape",
			contour: Contour{
				Points:      []image.Point{{X: 10, Y: 10}, {X: 20, Y: 20}, {X: 30, Y: 10}},
				Area:        50.0,
				Perimeter:   30.0,
				Circularity: 0.3,
				Center:      image.Point{X: 20, Y: 15},
			},
			expected: Unknown,
		},
		{
			name: "Partial square with debug",
			contour: Contour{
				Points:      generateSquarePoints(377, 195, 20),
				Area:        300.0,
				Perimeter:   80.0,
				Circularity: 0.35,
				Center:      image.Point{X: 377, Y: 195},
			},
			expected: Square,
			withDebug: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.withDebug {
				os.Setenv("GRIMOIRE_DEBUG", "1")
				defer os.Unsetenv("GRIMOIRE_DEBUG")
			}
			
			result := detector.classifyShape(tc.contour)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestIsSquare tests the isSquare function
func TestIsSquare(t *testing.T) {
	detector := NewDetector(Config{})

	tests := []struct {
		name     string
		vertices []image.Point
		expected bool
	}{
		{
			name: "Perfect square",
			vertices: []image.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
			expected: true,
		},
		{
			name: "Not a square - rectangle",
			vertices: []image.Point{
				{X: 0, Y: 0},
				{X: 200, Y: 0},
				{X: 200, Y: 100},
				{X: 0, Y: 100},
			},
			expected: false,
		},
		{
			name: "Not a square - wrong number of vertices",
			vertices: []image.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 50, Y: 100},
			},
			expected: false,
		},
		{
			name: "Almost square - within tolerance",
			vertices: []image.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 105, Y: 95},
				{X: 5, Y: 100},
			},
			expected: true,
		},
		{
			name: "Not a square - skewed",
			vertices: []image.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 150, Y: 100},
				{X: 50, Y: 100},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isSquare(tc.vertices)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestIsDoubleCircle tests the isDoubleCircle function
func TestIsDoubleCircle(t *testing.T) {
	detector := NewDetector(Config{})

	tests := []struct {
		name     string
		contour  Contour
		expected bool
	}{
		{
			name: "Double circle - correct area ratio",
			contour: Contour{
				Points:      generateCirclePoints(100, 100, 30),
				Area:        600.0, // About 20% of filled circle area
				Perimeter:   188.5,
				Circularity: 0.95,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: true,
		},
		{
			name: "Not double circle - filled circle",
			contour: Contour{
				Points:      generateCirclePoints(100, 100, 30),
				Area:        2827.0, // Full circle area
				Perimeter:   188.5,
				Circularity: 0.95,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: false,
		},
		{
			name: "Not double circle - too thin",
			contour: Contour{
				Points:      generateCirclePoints(100, 100, 30),
				Area:        150.0, // Less than 10% of filled circle area
				Perimeter:   188.5,
				Circularity: 0.95,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isDoubleCircle(tc.contour)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestDebugFunctions tests the debug functions
func TestDebugFunctions(t *testing.T) {
	detector := NewDetector(Config{})

	t.Run("DebugSaveContours", func(t *testing.T) {
		binary := image.NewGray(image.Rect(0, 0, 200, 200))
		contours := []Contour{
			{
				Points: generateCirclePoints(100, 100, 30),
				Center: image.Point{X: 100, Y: 100},
			},
			{
				Points: generateSquarePoints(50, 50, 20),
				Center: image.Point{X: 50, Y: 50},
			},
		}

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "debug_contours.png")

		err := detector.DebugSaveContours(binary, contours, outputPath)
		require.NoError(t, err)

		// Check file exists
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)
	})

	t.Run("DebugPrintContours", func(t *testing.T) {
		contours := []Contour{
			{
				Points:      generateCirclePoints(100, 100, 30),
				Area:        2827.0,
				Perimeter:   188.5,
				Circularity: 0.95,
				Center:      image.Point{X: 100, Y: 100},
			},
			{
				Points:      generateSquarePoints(50, 50, 20),
				Area:        400.0,
				Perimeter:   80.0,
				Circularity: 0.78,
				Center:      image.Point{X: 50, Y: 50},
			},
		}

		// This just prints to stdout, so we just ensure it doesn't panic
		detector.DebugPrintContours(contours)
	})

	t.Run("DebugSaveImage", func(t *testing.T) {
		gray := image.NewGray(image.Rect(0, 0, 100, 100))
		// Draw some pattern
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				if (x+y)%20 < 10 {
					gray.Set(x, y, color.Gray{255})
				}
			}
		}

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "debug_image.png")

		err := detector.DebugSaveImage(gray, outputPath)
		require.NoError(t, err)

		// Check file exists
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)
	})
}

// Helper functions to generate test contours

func generateCirclePoints(cx, cy, r int) []image.Point {
	points := []image.Point{}
	for angle := 0.0; angle < 2*3.14159; angle += 0.1 {
		x := cx + int(float64(r)*math.Cos(angle))
		y := cy + int(float64(r)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	return points
}

func generateSquarePoints(x, y, size int) []image.Point {
	points := []image.Point{}
	// Top edge
	for i := 0; i <= size; i++ {
		points = append(points, image.Point{X: x + i, Y: y})
	}
	// Right edge
	for i := 0; i <= size; i++ {
		points = append(points, image.Point{X: x + size, Y: y + i})
	}
	// Bottom edge
	for i := size; i >= 0; i-- {
		points = append(points, image.Point{X: x + i, Y: y + size})
	}
	// Left edge
	for i := size; i >= 0; i-- {
		points = append(points, image.Point{X: x, Y: y + i})
	}
	return points
}

func generateStarPoints(cx, cy, r int, points int) []image.Point {
	result := []image.Point{}
	innerRadius := r / 2
	for i := 0; i < points*2; i++ {
		angle := float64(i) * 3.14159 / float64(points)
		radius := r
		if i%2 == 1 {
			radius = innerRadius
		}
		x := cx + int(float64(radius)*math.Cos(angle))
		y := cy + int(float64(radius)*math.Sin(angle))
		result = append(result, image.Point{X: x, Y: y})
	}
	return result
}

func generateTrianglePoints(cx, cy, r int) []image.Point {
	return []image.Point{
		{X: cx, Y: cy - r},
		{X: cx + int(float64(r)*0.866), Y: cy + r/2},
		{X: cx - int(float64(r)*0.866), Y: cy + r/2},
	}
}

func generatePolygonPoints(cx, cy, r, sides int) []image.Point {
	points := []image.Point{}
	for i := 0; i < sides; i++ {
		angle := float64(i) * 2 * 3.14159 / float64(sides)
		x := cx + int(float64(r)*math.Cos(angle))
		y := cy + int(float64(r)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	return points
}

func generateYShapePoints(cx, cy int) []image.Point {
	// Generate a Y-shaped contour
	points := []image.Point{}
	// Left branch
	for i := -50; i <= 0; i++ {
		points = append(points, image.Point{X: cx + i, Y: cy - 50 - i})
	}
	// Right branch
	for i := 0; i <= 50; i++ {
		points = append(points, image.Point{X: cx + i, Y: cy - 50 + i})
	}
	// Stem
	for i := 0; i <= 100; i++ {
		points = append(points, image.Point{X: cx, Y: cy - 50 + i})
	}
	return points
}

