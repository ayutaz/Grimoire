package detector

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// DebugSaveContours saves an image with detected contours for debugging
func (d *Detector) DebugSaveContours(binary *image.Gray, contours []Contour, outputPath string) error {
	bounds := binary.Bounds()
	output := image.NewRGBA(bounds)

	// Copy binary image to output
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := binary.GrayAt(x, y)
			if gray.Y > 128 {
				output.Set(x, y, color.RGBA{255, 255, 255, 255})
			} else {
				output.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}

	// Draw contours in different colors
	colors := []color.RGBA{
		{255, 0, 0, 255},   // Red
		{0, 255, 0, 255},   // Green
		{0, 0, 255, 255},   // Blue
		{255, 255, 0, 255}, // Yellow
		{255, 0, 255, 255}, // Magenta
		{0, 255, 255, 255}, // Cyan
	}

	for i, contour := range contours {
		c := colors[i%len(colors)]

		// Draw contour points
		for _, pt := range contour.Points {
			if pt.X >= bounds.Min.X && pt.X < bounds.Max.X &&
				pt.Y >= bounds.Min.Y && pt.Y < bounds.Max.Y {
				output.Set(pt.X, pt.Y, c)
			}
		}

		// Draw center
		drawCross(output, contour.Center, c)
	}

	// Save image
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, output)
}

// DebugPrintContours prints contour information for debugging
func (d *Detector) DebugPrintContours(contours []Contour) {
	fmt.Printf("Found %d contours:\n", len(contours))
	for i, contour := range contours {
		symbolType := d.classifyShape(contour)
		fmt.Printf("[%d] Type: %s, Area: %.2f, Circularity: %.2f, Center: (%d,%d), Points: %d\n",
			i, symbolType, contour.Area, contour.Circularity,
			contour.Center.X, contour.Center.Y, len(contour.Points))
	}
}

// drawCross draws a small cross at the given point
func drawCross(img *image.RGBA, center image.Point, c color.RGBA) {
	size := 5
	bounds := img.Bounds()

	// Horizontal line
	for dx := -size; dx <= size; dx++ {
		x := center.X + dx
		if x >= bounds.Min.X && x < bounds.Max.X {
			img.Set(x, center.Y, c)
		}
	}

	// Vertical line
	for dy := -size; dy <= size; dy++ {
		y := center.Y + dy
		if y >= bounds.Min.Y && y < bounds.Max.Y {
			img.Set(center.X, y, c)
		}
	}
}

// DebugSaveImage saves a grayscale image for debugging
func (d *Detector) DebugSaveImage(img *image.Gray, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}
