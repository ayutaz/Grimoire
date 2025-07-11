package detector

import (
	"image"
	"image/color"
	"math"
)

// improvedDetectDiagonalLines uses multiple methods to detect diagonal connections
func (d *Detector) improvedDetectDiagonalLines(edges *image.Gray) []Line {
	lines := []Line{}
	
	// Method 1: Original diagonal line following
	lines = append(lines, d.detectDiagonalLines(edges)...)
	
	// Method 2: Hough transform for diagonal lines
	lines = append(lines, d.detectDiagonalLinesHough(edges)...)
	
	// Method 3: Direct symbol-to-symbol diagonal check
	// This will be handled in improvedDetectConnections
	
	// Remove duplicates
	lines = d.removeDuplicateLines(lines)
	
	return lines
}

// detectDiagonalLinesHough uses Hough transform specifically for diagonal lines
func (d *Detector) detectDiagonalLinesHough(edges *image.Gray) []Line {
	lines := []Line{}
	bounds := edges.Bounds()
	
	// Accumulator for diagonal lines at 45 and -45 degrees
	// We'll use a simplified approach focusing on these specific angles
	type DiagonalLine struct {
		intercept int
		angle     float64 // either PI/4 or -PI/4
		points    []image.Point
	}
	
	// Maps for 45 and -45 degree lines
	diag45 := make(map[int]*DiagonalLine)  // y = x + b
	diagM45 := make(map[int]*DiagonalLine) // y = -x + b
	
	// Find edge points
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if edges.GrayAt(x, y).Y > 128 {
				// For 45 degree line: b = y - x
				b45 := y - x
				if diag45[b45] == nil {
					diag45[b45] = &DiagonalLine{
						intercept: b45,
						angle:     math.Pi / 4,
						points:    []image.Point{},
					}
				}
				diag45[b45].points = append(diag45[b45].points, image.Point{X: x, Y: y})
				
				// For -45 degree line: b = y + x
				bM45 := y + x
				if diagM45[bM45] == nil {
					diagM45[bM45] = &DiagonalLine{
						intercept: bM45,
						angle:     -math.Pi / 4,
						points:    []image.Point{},
					}
				}
				diagM45[bM45].points = append(diagM45[bM45].points, image.Point{X: x, Y: y})
			}
		}
	}
	
	// Extract lines from accumulators
	minPoints := 15 // Minimum points to form a line
	
	for _, diag := range diag45 {
		if len(diag.points) >= minPoints {
			line := d.extractLineFromPoints(diag.points)
			if line != nil {
				lines = append(lines, *line)
			}
		}
	}
	
	for _, diag := range diagM45 {
		if len(diag.points) >= minPoints {
			line := d.extractLineFromPoints(diag.points)
			if line != nil {
				lines = append(lines, *line)
			}
		}
	}
	
	return lines
}

// extractLineFromPoints finds the best line segment from a set of points
func (d *Detector) extractLineFromPoints(points []image.Point) *Line {
	if len(points) < 2 {
		return nil
	}
	
	// Find the two points that are farthest apart
	maxDist := 0.0
	var start, end image.Point
	
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			dist := distance(points[i], points[j])
			if dist > maxDist {
				maxDist = dist
				start = points[i]
				end = points[j]
			}
		}
	}
	
	// Only return if the line is long enough
	if maxDist >= 20 {
		return &Line{Start: start, End: end}
	}
	
	return nil
}

// removeDuplicateLines removes duplicate or very similar lines
func (d *Detector) removeDuplicateLines(lines []Line) []Line {
	if len(lines) <= 1 {
		return lines
	}
	
	unique := []Line{}
	
	for i, line1 := range lines {
		isDuplicate := false
		
		for j := 0; j < i; j++ {
			line2 := lines[j]
			
			// Check if lines are very similar
			if d.linesAreSimilar(line1, line2) {
				isDuplicate = true
				break
			}
		}
		
		if !isDuplicate {
			unique = append(unique, line1)
		}
	}
	
	return unique
}

// linesAreSimilar checks if two lines are essentially the same
func (d *Detector) linesAreSimilar(l1, l2 Line) bool {
	// Check if endpoints are very close
	threshold := 10.0
	
	// Check both orientations
	if (distance(l1.Start, l2.Start) < threshold && distance(l1.End, l2.End) < threshold) ||
		(distance(l1.Start, l2.End) < threshold && distance(l1.End, l2.Start) < threshold) {
		return true
	}
	
	return false
}

// improvedDetectConnections enhances connection detection with better diagonal support
func (d *Detector) improvedDetectConnections(binary *image.Gray, symbols []*Symbol) []Connection {
	connections := []Connection{}
	
	// First, use the standard connection detection
	connections = append(connections, d.detectConnections(binary, symbols)...)
	
	// Additionally, check for direct symbol-to-symbol diagonal connections
	// This helps when the line detection misses some diagonal connections
	for i, sym1 := range symbols {
		for j := i + 1; j < len(symbols); j++ {
			sym2 := symbols[j]
			
			// Skip if already connected or invalid pair
			if d.alreadyConnected(connections, sym1, sym2) {
				continue
			}
			
			// Check if symbols might be diagonally connected
			if d.symbolsMightBeDiagonallyConnected(sym1, sym2, binary) {
				// Determine direction
				from, to := d.determineConnectionDirection(sym1, sym2)
				
				conn := Connection{
					From:           from,
					To:             to,
					ConnectionType: "solid",
					Properties:     make(map[string]interface{}),
				}
				
				connections = append(connections, conn)
			}
		}
	}
	
	return connections
}

// alreadyConnected checks if two symbols are already connected
func (d *Detector) alreadyConnected(connections []Connection, sym1, sym2 *Symbol) bool {
	for _, conn := range connections {
		if (conn.From == sym1 && conn.To == sym2) || (conn.From == sym2 && conn.To == sym1) {
			return true
		}
	}
	return false
}

// symbolsMightBeDiagonallyConnected checks if two symbols might have a diagonal connection
func (d *Detector) symbolsMightBeDiagonallyConnected(sym1, sym2 *Symbol, binary *image.Gray) bool {
	// Skip outer circle and same symbol
	if sym1.Type == OuterCircle || sym2.Type == OuterCircle || sym1 == sym2 {
		return false
	}
	
	// Calculate angle between symbols
	dx := float64(sym2.Position.X - sym1.Position.X)
	dy := float64(sym2.Position.Y - sym1.Position.Y)
	angle := math.Atan2(dy, dx)
	
	// Check if angle is roughly diagonal (45 or -45 degrees)
	diag45 := math.Pi / 4
	diagM45 := -math.Pi / 4
	diag135 := 3 * math.Pi / 4
	diagM135 := -3 * math.Pi / 4
	
	angleThreshold := math.Pi / 8 // 22.5 degrees tolerance
	
	isDiagonal := math.Abs(angle-diag45) < angleThreshold ||
		math.Abs(angle-diagM45) < angleThreshold ||
		math.Abs(angle-diag135) < angleThreshold ||
		math.Abs(angle-diagM135) < angleThreshold
	
	if !isDiagonal {
		return false
	}
	
	// Check if there's a path of dark pixels between the symbols
	return d.hasPixelPath(sym1.Position, sym2.Position, binary)
}

// hasPixelPath checks if there's a path of dark pixels between two points
func (d *Detector) hasPixelPath(start, end Position, binary *image.Gray) bool {
	// Sample points along the line between start and end
	dx := end.X - start.X
	dy := end.Y - start.Y
	steps := int(math.Max(math.Abs(dx), math.Abs(dy)))
	
	if steps == 0 {
		return false
	}
	
	darkPixels := 0
	totalPixels := 0
	
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(float64(start.X) + t*float64(dx))
		y := int(float64(start.Y) + t*float64(dy))
		
		// Check in a small area around the point
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				px := x + dx
				py := y + dy
				
				if px >= 0 && py >= 0 && px < binary.Bounds().Max.X && py < binary.Bounds().Max.Y {
					totalPixels++
					if binary.GrayAt(px, py).Y < 128 { // Dark pixel
						darkPixels++
					}
				}
			}
		}
	}
	
	// Require at least 30% dark pixels along the path
	if totalPixels == 0 {
		return false
	}
	
	ratio := float64(darkPixels) / float64(totalPixels)
	return ratio > 0.3
}

// improvedEdgeDetection performs better edge detection for diagonal lines
func (d *Detector) improvedEdgeDetection(binary *image.Gray) *image.Gray {
	bounds := binary.Bounds()
	edges := image.NewGray(bounds)
	
	// Apply Canny-like edge detection
	// Step 1: Gaussian blur (simplified)
	blurred := d.simpleGaussianBlur(binary)
	
	// Step 2: Gradient calculation with diagonal kernels
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			// Standard Sobel
			gx := int(blurred.GrayAt(x+1, y-1).Y) + 2*int(blurred.GrayAt(x+1, y).Y) + int(blurred.GrayAt(x+1, y+1).Y) -
				int(blurred.GrayAt(x-1, y-1).Y) - 2*int(blurred.GrayAt(x-1, y).Y) - int(blurred.GrayAt(x-1, y+1).Y)
			
			gy := int(blurred.GrayAt(x-1, y+1).Y) + 2*int(blurred.GrayAt(x, y+1).Y) + int(blurred.GrayAt(x+1, y+1).Y) -
				int(blurred.GrayAt(x-1, y-1).Y) - 2*int(blurred.GrayAt(x, y-1).Y) - int(blurred.GrayAt(x+1, y-1).Y)
			
			// Diagonal kernels for better diagonal edge detection
			g45 := int(blurred.GrayAt(x, y-1).Y) + int(blurred.GrayAt(x+1, y).Y) -
				int(blurred.GrayAt(x-1, y).Y) - int(blurred.GrayAt(x, y+1).Y)
			
			gM45 := int(blurred.GrayAt(x-1, y-1).Y) + int(blurred.GrayAt(x, y).Y) -
				int(blurred.GrayAt(x, y).Y) - int(blurred.GrayAt(x+1, y+1).Y)
			
			// Combine all gradients
			magnitude := int(math.Sqrt(float64(gx*gx + gy*gy + g45*g45 + gM45*gM45)))
			
			if magnitude > 40 { // Lower threshold for better sensitivity
				edges.Set(x, y, color.Gray{255})
			} else {
				edges.Set(x, y, color.Gray{0})
			}
		}
	}
	
	return edges
}

// simpleGaussianBlur applies a simple 3x3 Gaussian blur
func (d *Detector) simpleGaussianBlur(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	blurred := image.NewGray(bounds)
	
	// 3x3 Gaussian kernel (approximation)
	kernel := [][]int{
		{1, 2, 1},
		{2, 4, 2},
		{1, 2, 1},
	}
	kernelSum := 16
	
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			sum := 0
			
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := img.GrayAt(x+kx, y+ky).Y
					weight := kernel[ky+1][kx+1]
					sum += int(pixel) * weight
				}
			}
			
			blurred.Set(x, y, color.Gray{uint8(sum / kernelSum)})
		}
	}
	
	// Copy edges
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		blurred.Set(bounds.Min.X, y, img.GrayAt(bounds.Min.X, y))
		blurred.Set(bounds.Max.X-1, y, img.GrayAt(bounds.Max.X-1, y))
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		blurred.Set(x, bounds.Min.Y, img.GrayAt(x, bounds.Min.Y))
		blurred.Set(x, bounds.Max.Y-1, img.GrayAt(x, bounds.Max.Y-1))
	}
	
	return blurred
}