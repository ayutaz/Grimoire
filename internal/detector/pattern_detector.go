package detector

import (
	"image"
	"image/color"
)

// detectInternalPattern analyzes the pattern inside a symbol
func (d *Detector) detectInternalPattern(contour Contour, binary *image.Gray) string {
	bbox := contour.getBoundingBox()
	
	// Create a mask for the contour region
	mask := d.createContourMask(contour, binary.Bounds())
	
	// Count white pixels inside the contour
	whitePixels := 0
	totalPixels := 0
	
	for y := bbox.Min.Y; y < bbox.Max.Y; y++ {
		for x := bbox.Min.X; x < bbox.Max.X; x++ {
			if mask.GrayAt(x, y).Y > 0 {
				totalPixels++
				if binary.GrayAt(x, y).Y > 128 {
					whitePixels++
				}
			}
		}
	}
	
	if totalPixels == 0 {
		return "empty"
	}
	
	fillRatio := float64(whitePixels) / float64(totalPixels)
	
	// Analyze pattern based on fill ratio and distribution
	if fillRatio < 0.1 {
		return "empty"
	} else if fillRatio < 0.3 {
		return d.analyzeSparseFill(contour, binary, mask)
	} else if fillRatio < 0.7 {
		return d.analyzeMediumFill(contour, binary, mask)
	} else {
		return d.analyzeDenseFill(contour, binary, mask)
	}
}

// createContourMask creates a mask for pixels inside the contour
func (d *Detector) createContourMask(contour Contour, bounds image.Rectangle) *image.Gray {
	mask := image.NewGray(bounds)
	
	// Simple point-in-polygon test for each pixel
	bbox := contour.getBoundingBox()
	
	for y := bbox.Min.Y; y < bbox.Max.Y; y++ {
		for x := bbox.Min.X; x < bbox.Max.X; x++ {
			if d.isPointInContour(image.Point{X: x, Y: y}, contour) {
				mask.SetGray(x, y, color.Gray{255})
			}
		}
	}
	
	return mask
}

// isPointInContour checks if a point is inside a contour using ray casting
func (d *Detector) isPointInContour(point image.Point, contour Contour) bool {
	if len(contour.Points) < 3 {
		return false
	}
	
	// Ray casting algorithm
	inside := false
	p1 := contour.Points[0]
	
	for i := 1; i <= len(contour.Points); i++ {
		p2 := contour.Points[i%len(contour.Points)]
		
		if point.Y > min(p1.Y, p2.Y) && point.Y <= max(p1.Y, p2.Y) {
			if point.X <= max(p1.X, p2.X) {
				xinters := float64(p1.X)
				if p1.Y != p2.Y {
					xinters = float64(point.Y-p1.Y)*float64(p2.X-p1.X)/float64(p2.Y-p1.Y) + float64(p1.X)
				}
				if p1.X == p2.X || float64(point.X) <= xinters {
					inside = !inside
				}
			}
		}
		p1 = p2
	}
	
	return inside
}

// analyzeSparseFill analyzes patterns with sparse fill (dots, points)
func (d *Detector) analyzeSparseFill(contour Contour, binary *image.Gray, mask *image.Gray) string {
	// Count distinct white regions (dots)
	bbox := contour.getBoundingBox()
	visited := make(map[image.Point]bool)
	dotCount := 0
	
	for y := bbox.Min.Y; y < bbox.Max.Y; y++ {
		for x := bbox.Min.X; x < bbox.Max.X; x++ {
			pt := image.Point{X: x, Y: y}
			if mask.GrayAt(x, y).Y > 0 && binary.GrayAt(x, y).Y > 128 && !visited[pt] {
				// Found a white pixel, count the connected component
				d.markConnectedComponent(binary, mask, pt, visited)
				dotCount++
			}
		}
	}
	
	switch dotCount {
	case 1:
		return "dot"
	case 2:
		return "double_dot"
	case 3:
		return "triple_dot"
	default:
		if dotCount > 3 && dotCount < 10 {
			return "multi_dot"
		}
		return "pattern"
	}
}

// analyzeMediumFill analyzes patterns with medium fill (lines, shapes)
func (d *Detector) analyzeMediumFill(contour Contour, binary *image.Gray, mask *image.Gray) string {
	// Check for line patterns by analyzing horizontal and vertical projections
	bbox := contour.getBoundingBox()
	
	// Count horizontal and vertical lines
	horizontalLines := d.countLines(binary, mask, bbox, true)
	verticalLines := d.countLines(binary, mask, bbox, false)
	
	if horizontalLines > verticalLines*2 {
		return "horizontal_lines"
	} else if verticalLines > horizontalLines*2 {
		return "vertical_lines"
	} else if horizontalLines > 0 || verticalLines > 0 {
		return "lines"
	}
	
	// Check for circular pattern
	if d.hasCircularPattern(binary, mask, contour) {
		return "half_circle"
	}
	
	return "pattern"
}

// analyzeDenseFill analyzes patterns with dense fill
func (d *Detector) analyzeDenseFill(contour Contour, binary *image.Gray, mask *image.Gray) string {
	// Check for cross pattern
	if d.hasCrossPattern(binary, mask, contour) {
		return "cross"
	}
	
	return "filled"
}

// markConnectedComponent marks all pixels in a connected component as visited
func (d *Detector) markConnectedComponent(binary, mask *image.Gray, start image.Point, visited map[image.Point]bool) {
	bounds := binary.Bounds()
	queue := []image.Point{start}
	
	for len(queue) > 0 {
		pt := queue[0]
		queue = queue[1:]
		
		if visited[pt] {
			continue
		}
		
		visited[pt] = true
		
		// Check 4-connected neighbors
		neighbors := []image.Point{
			{X: pt.X + 1, Y: pt.Y},
			{X: pt.X - 1, Y: pt.Y},
			{X: pt.X, Y: pt.Y + 1},
			{X: pt.X, Y: pt.Y - 1},
		}
		
		for _, n := range neighbors {
			if n.X >= bounds.Min.X && n.X < bounds.Max.X &&
				n.Y >= bounds.Min.Y && n.Y < bounds.Max.Y &&
				!visited[n] &&
				mask.GrayAt(n.X, n.Y).Y > 0 &&
				binary.GrayAt(n.X, n.Y).Y > 128 {
				queue = append(queue, n)
			}
		}
	}
}

// countLines counts the number of lines in horizontal or vertical direction
func (d *Detector) countLines(binary, mask *image.Gray, bbox image.Rectangle, horizontal bool) int {
	lineCount := 0
	inLine := false
	
	if horizontal {
		for y := bbox.Min.Y; y < bbox.Max.Y; y++ {
			whiteCount := 0
			for x := bbox.Min.X; x < bbox.Max.X; x++ {
				if mask.GrayAt(x, y).Y > 0 && binary.GrayAt(x, y).Y > 128 {
					whiteCount++
				}
			}
			if whiteCount > bbox.Dx()/3 {
				if !inLine {
					lineCount++
					inLine = true
				}
			} else {
				inLine = false
			}
		}
	} else {
		for x := bbox.Min.X; x < bbox.Max.X; x++ {
			whiteCount := 0
			for y := bbox.Min.Y; y < bbox.Max.Y; y++ {
				if mask.GrayAt(x, y).Y > 0 && binary.GrayAt(x, y).Y > 128 {
					whiteCount++
				}
			}
			if whiteCount > bbox.Dy()/3 {
				if !inLine {
					lineCount++
					inLine = true
				}
			} else {
				inLine = false
			}
		}
	}
	
	return lineCount
}

// hasCircularPattern checks if the pattern forms a circular shape
func (d *Detector) hasCircularPattern(binary, mask *image.Gray, contour Contour) bool {
	// Simplified check - look for arc-like patterns
	center := contour.Center
	radius := float64(contour.getBoundingBox().Dx()) / 4
	
	// Sample points along a circle
	whiteCount := 0
	totalCount := 0
	
	for angle := 0.0; angle < 3.14159; angle += 0.1 {
		x := int(float64(center.X) + radius*cos(angle))
		y := int(float64(center.Y) + radius*sin(angle))
		
		if x >= 0 && y >= 0 && x < binary.Bounds().Max.X && y < binary.Bounds().Max.Y {
			totalCount++
			if mask.GrayAt(x, y).Y > 0 && binary.GrayAt(x, y).Y > 128 {
				whiteCount++
			}
		}
	}
	
	return totalCount > 0 && float64(whiteCount)/float64(totalCount) > 0.5
}

// hasCrossPattern checks if the pattern forms a cross
func (d *Detector) hasCrossPattern(binary, mask *image.Gray, contour Contour) bool {
	center := contour.Center
	bbox := contour.getBoundingBox()
	
	// Check horizontal line through center
	horizontalWhite := 0
	for x := bbox.Min.X; x < bbox.Max.X; x++ {
		if mask.GrayAt(x, center.Y).Y > 0 && binary.GrayAt(x, center.Y).Y > 128 {
			horizontalWhite++
		}
	}
	
	// Check vertical line through center
	verticalWhite := 0
	for y := bbox.Min.Y; y < bbox.Max.Y; y++ {
		if mask.GrayAt(center.X, y).Y > 0 && binary.GrayAt(center.X, y).Y > 128 {
			verticalWhite++
		}
	}
	
	// Cross pattern should have significant pixels in both directions
	return horizontalWhite > bbox.Dx()/3 && verticalWhite > bbox.Dy()/3
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func cos(angle float64) float64 {
	// Simple cosine approximation
	// In production, use math.Cos
	return 1.0 - angle*angle/2.0
}

func sin(angle float64) float64 {
	// Simple sine approximation
	// In production, use math.Sin
	return angle - angle*angle*angle/6.0
}