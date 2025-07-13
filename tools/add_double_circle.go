package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run add_double_circle.go <input_image>")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	
	// Open the input image
	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding image: %v\n", err)
		os.Exit(1)
	}

	// Convert to RGBA for editing
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Add double circle (main entry point) - positioned in upper left area
	centerX := 100
	centerY := 100
	outerRadius := 25
	innerRadius := 20
	
	// Draw outer circle
	drawCircle(rgba, centerX, centerY, outerRadius, color.Black)
	// Draw inner circle
	drawCircle(rgba, centerX, centerY, innerRadius, color.Black)

	// Save the modified image
	outputPath := inputPath[:len(inputPath)-4] + "_with_entry.png"
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	err = png.Encode(outFile, rgba)
	if err != nil {
		fmt.Printf("Error encoding image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully added double circle to %s\n", outputPath)
}

func drawCircle(img *image.RGBA, centerX, centerY, radius int, c color.Color) {
	for angle := 0.0; angle < 2*math.Pi; angle += 0.01 {
		x := centerX + int(float64(radius)*math.Cos(angle))
		y := centerY + int(float64(radius)*math.Sin(angle))
		
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