package detector

import (
	"image"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClassifyShape_EdgeCases tests edge cases for classifyShape function
func TestClassifyShape_EdgeCases(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		contour  Contour
		expected SymbolType
		withDebug bool
	}{
		{
			name: "Six-pointed star with 12 vertices",
			contour: Contour{
				Points:      generatePolygonPoints(100, 100, 30, 12),
				Area:        1200.0,
				Perimeter:   150.0,
				Circularity: 0.45,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: SixPointedStar,
		},
		{
			name: "Eight-pointed star with 16 vertices",
			contour: Contour{
				Points:      generatePolygonPoints(100, 100, 30, 16),
				Area:        1400.0,
				Perimeter:   180.0,
				Circularity: 0.40,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: EightPointedStar,
		},
		{
			name: "Amplification operator - 4-pointed star",
			contour: Contour{
				Points:      generateStarPoints(100, 100, 30, 4),
				Area:        800.0,
				Perimeter:   160.0,
				Circularity: 0.35,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Amplification,
		},
		{
			name: "Distribution operator - radial pattern",
			contour: Contour{
				Points:      generateRadialPattern(100, 100, 30),
				Area:        1100.0,
				Perimeter:   140.0,
				Circularity: 0.65,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Distribution,
		},
		{
			name: "Transfer operator - arrow shape",
			contour: Contour{
				Points:      generateArrowPoints(100, 100),
				Area:        600.0,
				Perimeter:   120.0,
				Circularity: 0.35,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Transfer,
		},
		{
			name: "Equal operator - parallel lines",
			contour: Contour{
				Points:      generateParallelLines(100, 100),
				Area:        400.0,
				Perimeter:   160.0,
				Circularity: 0.25,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Equal,
		},
		{
			name: "Less than operator",
			contour: Contour{
				Points:      generateLessThanShape(100, 100),
				Area:        250.0,
				Perimeter:   80.0,
				Circularity: 0.45,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: LessThan,
		},
		{
			name: "Greater than operator",
			contour: Contour{
				Points:      generateGreaterThanShape(100, 100),
				Area:        250.0,
				Perimeter:   80.0,
				Circularity: 0.45,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: GreaterThan,
		},
		{
			name: "Rounded square detection",
			contour: Contour{
				Points:      generateRoundedSquarePoints(100, 100, 30),
				Area:        850.0,
				Perimeter:   115.0,
				Circularity: 0.88,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Square,
		},
		{
			name: "Square with fill ratio check",
			contour: Contour{
				Points:      generateSquarePoints(100, 100, 30),
				Area:        850.0,
				Perimeter:   120.0,
				Circularity: 0.75,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Square,
		},
		{
			name: "Star with 8 vertices (not 4-pointed)",
			contour: Contour{
				Points:      generatePolygonPoints(100, 100, 30, 8),
				Area:        1050.0,
				Perimeter:   95.0,
				Circularity: 0.55,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Unknown,
		},
		{
			name: "Star with 10 vertices (not 5-pointed)",
			contour: Contour{
				Points:      generatePolygonPoints(100, 100, 30, 10),
				Area:        1150.0,
				Perimeter:   105.0,
				Circularity: 0.52,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Unknown,
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

// TestClassifyOperator_Coverage tests classifyOperator function branches
func TestClassifyOperator_Coverage(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		contour  Contour
		expected SymbolType
	}{
		{
			name: "Not an operator - high circularity",
			contour: Contour{
				Points:      generateCirclePoints(100, 100, 30),
				Area:        2827.0,
				Perimeter:   188.5,
				Circularity: 0.95,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Unknown,
		},
		{
			name: "Transfer with wrong aspect ratio",
			contour: Contour{
				Points:      generateSquarePoints(100, 100, 30),
				Area:        900.0,
				Perimeter:   120.0,
				Circularity: 0.78,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Unknown,
		},
		{
			name: "Triangle that's not an operator",
			contour: Contour{
				Points:      generateTrianglePoints(100, 100, 30),
				Area:        100.0, // Too small
				Perimeter:   90.0,
				Circularity: 0.7,
				Center:      image.Point{X: 100, Y: 100},
			},
			expected: Unknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.classifyOperator(tc.contour)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestApproximatePolygon_EdgeCases tests edge cases for approximatePolygon
func TestApproximatePolygon_EdgeCases(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name           string
		contour        Contour
		expectedCount  int
	}{
		{
			name: "Contour with less than 3 points",
			contour: Contour{
				Points:    []image.Point{{X: 10, Y: 10}, {X: 20, Y: 20}},
				Perimeter: 14.14,
			},
			expectedCount: 2,
		},
		{
			name: "Contour with small perimeter",
			contour: Contour{
				Points:      generateTrianglePoints(10, 10, 5),
				Perimeter:   15.0,
			},
			expectedCount: 2, // Will be simplified due to small epsilon
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.approximatePolygon(tc.contour)
			assert.LessOrEqual(t, len(result), tc.expectedCount)
		})
	}
}

// Helper functions for generating specific shapes

func generateRadialPattern(cx, cy, r int) []image.Point {
	points := []image.Point{}
	// Generate a pattern with radial variations
	for i := 0; i < 8; i++ {
		angle := float64(i) * 2 * math.Pi / 8
		// Alternate between inner and outer radius
		radius := r
		if i%2 == 0 {
			radius = int(float64(r) * 1.2)
		}
		x := cx + int(float64(radius)*math.Cos(angle))
		y := cy + int(float64(radius)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	return points
}

func generateArrowPoints(cx, cy int) []image.Point {
	// Generate arrow shape pointing right
	return []image.Point{
		{X: cx - 30, Y: cy - 10},
		{X: cx - 30, Y: cy + 10},
		{X: cx + 10, Y: cy + 10},
		{X: cx + 10, Y: cy + 20},
		{X: cx + 30, Y: cy},
		{X: cx + 10, Y: cy - 20},
		{X: cx + 10, Y: cy - 10},
	}
}

func generateParallelLines(cx, cy int) []image.Point {
	points := []image.Point{}
	// Top line
	for x := cx - 40; x <= cx + 40; x++ {
		points = append(points, image.Point{X: x, Y: cy - 10})
	}
	// Bottom line
	for x := cx + 40; x >= cx - 40; x-- {
		points = append(points, image.Point{X: x, Y: cy + 10})
	}
	return points
}

func generateLessThanShape(cx, cy int) []image.Point {
	// Generate < shape
	return []image.Point{
		{X: cx - 20, Y: cy},      // Leftmost point
		{X: cx + 20, Y: cy - 30}, // Top right
		{X: cx + 20, Y: cy + 30}, // Bottom right
	}
}

func generateGreaterThanShape(cx, cy int) []image.Point {
	// Generate > shape
	return []image.Point{
		{X: cx + 20, Y: cy},      // Rightmost point
		{X: cx - 20, Y: cy - 30}, // Top left
		{X: cx - 20, Y: cy + 30}, // Bottom left
	}
}

func generateRoundedSquarePoints(cx, cy, size int) []image.Point {
	points := []image.Point{}
	cornerRadius := size / 5
	
	// Top edge with rounded corners
	for x := cx - size/2 + cornerRadius; x <= cx + size/2 - cornerRadius; x++ {
		points = append(points, image.Point{X: x, Y: cy - size/2})
	}
	
	// Top-right corner
	for angle := -math.Pi/2; angle <= 0.0; angle += 0.1 {
		x := cx + size/2 - cornerRadius + int(float64(cornerRadius)*math.Cos(angle))
		y := cy - size/2 + cornerRadius + int(float64(cornerRadius)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	
	// Right edge
	for y := cy - size/2 + cornerRadius; y <= cy + size/2 - cornerRadius; y++ {
		points = append(points, image.Point{X: cx + size/2, Y: y})
	}
	
	// Bottom-right corner
	for angle := 0.0; angle <= math.Pi/2; angle += 0.1 {
		x := cx + size/2 - cornerRadius + int(float64(cornerRadius)*math.Cos(angle))
		y := cy + size/2 - cornerRadius + int(float64(cornerRadius)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	
	// Bottom edge
	for x := cx + size/2 - cornerRadius; x >= cx - size/2 + cornerRadius; x-- {
		points = append(points, image.Point{X: x, Y: cy + size/2})
	}
	
	// Bottom-left corner
	for angle := math.Pi/2; angle <= math.Pi; angle += 0.1 {
		x := cx - size/2 + cornerRadius + int(float64(cornerRadius)*math.Cos(angle))
		y := cy + size/2 - cornerRadius + int(float64(cornerRadius)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	
	// Left edge
	for y := cy + size/2 - cornerRadius; y >= cy - size/2 + cornerRadius; y-- {
		points = append(points, image.Point{X: cx - size/2, Y: y})
	}
	
	// Top-left corner
	for angle := math.Pi; angle <= 3*math.Pi/2; angle += 0.1 {
		x := cx - size/2 + cornerRadius + int(float64(cornerRadius)*math.Cos(angle))
		y := cy - size/2 + cornerRadius + int(float64(cornerRadius)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	
	return points
}