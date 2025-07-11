package detector

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
)

// createBenchmarkImage creates a test image with various shapes for benchmarking
func createBenchmarkImage(size int) *image.Gray {
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
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			dist := dx*dx + dy*dy
			if dist > (radius-5)*(radius-5) && dist < (radius+5)*(radius+5) {
				img.Set(x, y, color.Gray{0})
			}
		}
	}

	// Draw some shapes inside
	// Square
	for y := 100; y <= 150; y++ {
		for x := 100; x <= 150; x++ {
			if x == 100 || x == 150 || y == 100 || y == 150 {
				img.Set(x, y, color.Gray{0})
			}
		}
	}

	// Triangle
	for i := 0; i <= 50; i++ {
		img.Set(200+i, 100+i, color.Gray{0})
		img.Set(200+i, 100-i, color.Gray{0})
		img.Set(200+i, 100, color.Gray{0})
	}

	// Star
	for i := -20; i <= 20; i++ {
		img.Set(300+i, 300, color.Gray{0})
		img.Set(300, 300+i, color.Gray{0})
		if i >= -14 && i <= 14 {
			img.Set(300+i, 300+i, color.Gray{0})
			img.Set(300+i, 300-i, color.Gray{0})
		}
	}

	return img
}

func BenchmarkProcessImage(b *testing.B) {
	sizes := []int{400, 800, 1200}

	for _, size := range sizes {
		img := createBenchmarkImage(size)

		b.Run(fmt.Sprintf("size_%dx%d", size, size), func(b *testing.B) {
			detector := NewDetector()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				binary := detector.preprocessImage(img)
				contours := detector.findContours(binary)
				_ = detector.detectSymbolsFromContours(contours, binary)
			}
		})
	}
}

func BenchmarkPreprocessing(b *testing.B) {
	img := createBenchmarkImage(800)
	detector := NewDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.preprocessImage(img)
	}
}

func BenchmarkFindContours(b *testing.B) {
	img := createBenchmarkImage(800)
	detector := NewDetector()
	binary := detector.preprocessImage(img)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.findContours(binary)
	}
}

func BenchmarkClassifyShape(b *testing.B) {
	// Create different contour types
	contours := []Contour{
		// Small circle
		createCircleContour(50, 50, 20),
		// Square
		createSquareContour(100, 100, 40),
		// Complex shape
		createComplexContour(),
	}

	detector := NewDetector()

	for idx, contour := range contours {
		b.Run(fmt.Sprintf("contour_%d", idx), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.classifyShape(contour)
			}
		})
	}
}

func createCircleContour(cx, cy, r int) Contour {
	var points []image.Point
	steps := 36
	for i := 0; i < steps; i++ {
		angle := float64(i) * 2 * 3.14159 / float64(steps)
		x := cx + int(float64(r)*math.Cos(angle))
		y := cy + int(float64(r)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}
	return Contour{Points: points}
}

func createSquareContour(x, y, size int) Contour {
	return Contour{
		Points: []image.Point{
			{X: x, Y: y},
			{X: x + size, Y: y},
			{X: x + size, Y: y + size},
			{X: x, Y: y + size},
		},
	}
}

func createComplexContour() Contour {
	// Star shape
	var points []image.Point
	cx, cy := 200, 200
	outerRadius := 50
	innerRadius := 20

	for i := 0; i < 10; i++ {
		angle := float64(i) * 2 * 3.14159 / 10
		var r int
		if i%2 == 0 {
			r = outerRadius
		} else {
			r = innerRadius
		}
		x := cx + int(float64(r)*math.Cos(angle))
		y := cy + int(float64(r)*math.Sin(angle))
		points = append(points, image.Point{X: x, Y: y})
	}

	return Contour{Points: points}
}
