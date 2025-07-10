package detector

import (
	"image"
	"image/color"
	"math"
)

// gaussianBlur applies a simple Gaussian blur to an image
func gaussianBlur(img *image.Gray, kernelSize int) *image.Gray {
	bounds := img.Bounds()
	blurred := image.NewGray(bounds)
	
	// Simple box blur as approximation of Gaussian
	radius := kernelSize / 2
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var sum, count uint32
			
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						gray := img.GrayAt(nx, ny)
						sum += uint32(gray.Y)
						count++
					}
				}
			}
			
			if count > 0 {
				blurred.SetGray(x, y, color.Gray{uint8(sum / count)})
			}
		}
	}
	
	return blurred
}

// adaptiveThreshold applies adaptive thresholding to create a binary image
func adaptiveThreshold(gray *image.Gray, blockSize int, c int) *image.Gray {
	bounds := gray.Bounds()
	binary := image.NewGray(bounds)
	
	radius := blockSize / 2
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate local mean
			var sum, count uint32
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						gray := gray.GrayAt(nx, ny)
						sum += uint32(gray.Y)
						count++
					}
				}
			}
			
			threshold := uint8(sum/count) - uint8(c)
			pixel := gray.GrayAt(x, y)
			
			if pixel.Y < threshold {
				binary.SetGray(x, y, color.Gray{255}) // White for foreground
			} else {
				binary.SetGray(x, y, color.Gray{0}) // Black for background
			}
		}
	}
	
	return binary
}

// morphologyOpen performs morphological opening (erosion followed by dilation)
func morphologyOpen(binary *image.Gray, kernelSize int) *image.Gray {
	eroded := erode(binary, kernelSize)
	return dilate(eroded, kernelSize)
}

// morphologyClose performs morphological closing (dilation followed by erosion)
func morphologyClose(binary *image.Gray, kernelSize int) *image.Gray {
	dilated := dilate(binary, kernelSize)
	return erode(dilated, kernelSize)
}

// erode performs morphological erosion
func erode(binary *image.Gray, kernelSize int) *image.Gray {
	bounds := binary.Bounds()
	result := image.NewGray(bounds)
	radius := kernelSize / 2
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			allWhite := true
			
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						if binary.GrayAt(nx, ny).Y == 0 {
							allWhite = false
							break
						}
					}
				}
				if !allWhite {
					break
				}
			}
			
			if allWhite {
				result.SetGray(x, y, color.Gray{255})
			} else {
				result.SetGray(x, y, color.Gray{0})
			}
		}
	}
	
	return result
}

// dilate performs morphological dilation
func dilate(binary *image.Gray, kernelSize int) *image.Gray {
	bounds := binary.Bounds()
	result := image.NewGray(bounds)
	radius := kernelSize / 2
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			anyWhite := false
			
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						if binary.GrayAt(nx, ny).Y == 255 {
							anyWhite = true
							break
						}
					}
				}
				if anyWhite {
					break
				}
			}
			
			if anyWhite {
				result.SetGray(x, y, color.Gray{255})
			} else {
				result.SetGray(x, y, color.Gray{0})
			}
		}
	}
	
	return result
}

// distance calculates the Euclidean distance between two points
func distance(p1, p2 image.Point) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// isPointInCircle checks if a point is inside a circle
func isPointInCircle(point image.Point, center image.Point, radius float64) bool {
	return distance(point, center) <= radius
}