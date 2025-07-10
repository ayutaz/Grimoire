package detector

import (
	"image"
	"math"
	"sort"
)

// ContourPoint represents a point on a contour with direction info
type ContourPoint struct {
	Point     image.Point
	Direction int // 0-7 representing 8 directions
}

// findContours finds all contours in a binary image using improved algorithm
func (d *Detector) findContours(binary *image.Gray) []Contour {
	bounds := binary.Bounds()
	visited := make(map[image.Point]bool)
	var contours []Contour

	// Scan for all contours
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pt := image.Point{X: x, Y: y}
			
			// Skip if already visited or not a foreground pixel
			if visited[pt] || binary.GrayAt(x, y).Y != 255 {
				continue
			}
			
			// Found a new contour, trace it
			contour := d.traceContour(binary, pt, visited)
			if len(contour.Points) >= 10 { // Minimum points for a valid contour
				contour.calculateProperties()
				if contour.Area >= float64(d.minContourArea) {
					contours = append(contours, contour)
				}
			}
		}
	}
	
	// Sort contours by area (largest first)
	sort.Slice(contours, func(i, j int) bool {
		return contours[i].Area > contours[j].Area
	})
	
	// Try to merge contours that form the outer circle
	mergedContours := d.mergeCircularContours(contours, bounds)
	
	return mergedContours
}

// traceContour traces a contour using Moore neighborhood tracing
func (d *Detector) traceContour(binary *image.Gray, start image.Point, visited map[image.Point]bool) Contour {
	bounds := binary.Bounds()
	points := []image.Point{}
	
	// Direction vectors for 8-connectivity (clockwise from right)
	dirs := []image.Point{
		{1, 0}, {1, 1}, {0, 1}, {-1, 1},
		{-1, 0}, {-1, -1}, {0, -1}, {1, -1},
	}
	
	current := start
	points = append(points, current)
	visited[current] = true
	
	// Find initial direction
	var dir int
	for i := 0; i < 8; i++ {
		next := image.Point{X: current.X + dirs[i].X, Y: current.Y + dirs[i].Y}
		if d.isValidContourPoint(binary, next, bounds) {
			dir = i
			break
		}
	}
	
	// Trace the contour
	maxSteps := bounds.Dx() * bounds.Dy() // Prevent infinite loops
	steps := 0
	
	for steps < maxSteps {
		found := false
		
		// Check all 8 directions starting from current direction
		for i := 0; i < 8; i++ {
			checkDir := (dir + i) % 8
			next := image.Point{X: current.X + dirs[checkDir].X, Y: current.Y + dirs[checkDir].Y}
			
			if d.isValidContourPoint(binary, next, bounds) && !visited[next] {
				points = append(points, next)
				visited[next] = true
				current = next
				dir = checkDir
				found = true
				break
			}
		}
		
		if !found {
			// Try to find any connected pixel
			for i := 0; i < 8; i++ {
				next := image.Point{X: current.X + dirs[i].X, Y: current.Y + dirs[i].Y}
				if d.isValidContourPoint(binary, next, bounds) && !visited[next] {
					points = append(points, next)
					visited[next] = true
					current = next
					dir = i
					found = true
					break
				}
			}
		}
		
		if !found {
			break
		}
		
		// Check if we've returned to start
		if len(points) > 3 && distance(current, start) < 2.0 {
			break
		}
		
		steps++
	}
	
	// Fill any remaining connected pixels
	d.fillContourRegion(binary, points, visited, bounds)
	
	return Contour{Points: points}
}

// isValidContourPoint checks if a point is valid for contour tracing
func (d *Detector) isValidContourPoint(binary *image.Gray, pt image.Point, bounds image.Rectangle) bool {
	if pt.X < bounds.Min.X || pt.X >= bounds.Max.X ||
		pt.Y < bounds.Min.Y || pt.Y >= bounds.Max.Y {
		return false
	}
	return binary.GrayAt(pt.X, pt.Y).Y == 255
}

// fillContourRegion fills any remaining pixels in the contour region
func (d *Detector) fillContourRegion(binary *image.Gray, contourPoints []image.Point, visited map[image.Point]bool, bounds image.Rectangle) {
	// Create a bounding box for the contour
	if len(contourPoints) == 0 {
		return
	}
	
	minX, minY := contourPoints[0].X, contourPoints[0].Y
	maxX, maxY := minX, minY
	
	for _, pt := range contourPoints {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	
	// Scan the bounding box for any unvisited white pixels
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			pt := image.Point{X: x, Y: y}
			if !visited[pt] && d.isValidContourPoint(binary, pt, bounds) {
				visited[pt] = true
			}
		}
	}
}

// calculateProperties calculates area, center, and other properties of a contour
func (c *Contour) calculateProperties() {
	if len(c.Points) == 0 {
		return
	}
	
	// Calculate bounding box and center
	minX, minY := c.Points[0].X, c.Points[0].Y
	maxX, maxY := minX, minY
	sumX, sumY := 0, 0
	
	for _, pt := range c.Points {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
		sumX += pt.X
		sumY += pt.Y
	}
	
	c.Center = image.Point{
		X: sumX / len(c.Points),
		Y: sumY / len(c.Points),
	}
	
	// Calculate area using shoelace formula
	area := 0.0
	n := len(c.Points)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += float64(c.Points[i].X * c.Points[j].Y)
		area -= float64(c.Points[j].X * c.Points[i].Y)
	}
	c.Area = math.Abs(area) / 2.0
	
	// Calculate perimeter
	perimeter := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		perimeter += distance(c.Points[i], c.Points[j])
	}
	c.Perimeter = perimeter
	
	// Calculate circularity
	if c.Perimeter > 0 {
		c.Circularity = 4 * math.Pi * c.Area / (c.Perimeter * c.Perimeter)
	}
}

// Contour represents a detected contour with additional properties
type Contour struct {
	Points      []image.Point
	Area        float64
	Center      image.Point
	Perimeter   float64
	Circularity float64
}

// isCircle checks if the contour is circular based on circularity measure
func (c *Contour) isCircle(threshold float64) bool {
	return c.Circularity > threshold
}

// getBoundingBox returns the bounding rectangle of the contour
func (c *Contour) getBoundingBox() image.Rectangle {
	if len(c.Points) == 0 {
		return image.Rectangle{}
	}
	
	minX, minY := c.Points[0].X, c.Points[0].Y
	maxX, maxY := minX, minY
	
	for _, pt := range c.Points {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// getAspectRatio returns the aspect ratio of the bounding box
func (c *Contour) getAspectRatio() float64 {
	bbox := c.getBoundingBox()
	width := float64(bbox.Dx())
	height := float64(bbox.Dy())
	
	if width == 0 || height == 0 {
		return 1.0
	}
	
	return math.Max(width, height) / math.Min(width, height)
}

// findEdgeContour looks for contours that touch the image edges (likely outer circle)
func (d *Detector) findEdgeContour(binary *image.Gray, visited map[image.Point]bool) *Contour {
	bounds := binary.Bounds()
	margin := 5 // pixels from edge
	
	// Check top edge
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Min.Y+margin && y < bounds.Max.Y; y++ {
			pt := image.Point{X: x, Y: y}
			if !visited[pt] && binary.GrayAt(x, y).Y == 255 {
				contour := d.traceContour(binary, pt, visited)
				if len(contour.Points) > 100 { // Large contour
					return &contour
				}
			}
		}
	}
	
	// Check bottom edge
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Max.Y-margin; y < bounds.Max.Y && y >= bounds.Min.Y; y++ {
			pt := image.Point{X: x, Y: y}
			if !visited[pt] && binary.GrayAt(x, y).Y == 255 {
				contour := d.traceContour(binary, pt, visited)
				if len(contour.Points) > 100 {
					return &contour
				}
			}
		}
	}
	
	// Check left edge
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Min.X+margin && x < bounds.Max.X; x++ {
			pt := image.Point{X: x, Y: y}
			if !visited[pt] && binary.GrayAt(x, y).Y == 255 {
				contour := d.traceContour(binary, pt, visited)
				if len(contour.Points) > 100 {
					return &contour
				}
			}
		}
	}
	
	// Check right edge
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Max.X-margin; x < bounds.Max.X && x >= bounds.Min.X; x++ {
			pt := image.Point{X: x, Y: y}
			if !visited[pt] && binary.GrayAt(x, y).Y == 255 {
				contour := d.traceContour(binary, pt, visited)
				if len(contour.Points) > 100 {
					return &contour
				}
			}
		}
	}
	
	return nil
}

// mergeCircularContours attempts to merge contours that form a large circle
func (d *Detector) mergeCircularContours(contours []Contour, bounds image.Rectangle) []Contour {
	if len(contours) == 0 {
		return contours
	}
	
	// Calculate image center and radius
	imageCenter := image.Point{
		X: bounds.Dx() / 2,
		Y: bounds.Dy() / 2,
	}
	imageRadius := float64(min(bounds.Dx(), bounds.Dy())) / 2
	
	// Find contours that could be part of the outer circle
	var circularContours []Contour
	var otherContours []Contour
	
	for _, contour := range contours {
		// Check if contour is near the edge of the image
		isNearEdge := false
		maxDistFromCenter := 0.0
		minDistFromCenter := math.MaxFloat64
		
		for _, pt := range contour.Points {
			// Distance from image center
			dist := distance(pt, imageCenter)
			if dist > maxDistFromCenter {
				maxDistFromCenter = dist
			}
			if dist < minDistFromCenter {
				minDistFromCenter = dist
			}
			
			// Check if near edge
			if pt.X < 10 || pt.X > bounds.Dx()-10 ||
				pt.Y < 10 || pt.Y > bounds.Dy()-10 {
				isNearEdge = true
			}
		}
		
		// If contour is circular and near the expected radius
		if isNearEdge && maxDistFromCenter > imageRadius*0.7 &&
			maxDistFromCenter < imageRadius*1.3 {
			circularContours = append(circularContours, contour)
		} else {
			otherContours = append(otherContours, contour)
		}
	}
	
	// If we found multiple circular contours, merge them
	if len(circularContours) > 1 {
		merged := d.mergeContours(circularContours)
		result := []Contour{merged}
		result = append(result, otherContours...)
		return result
	}
	
	return contours
}

// mergeContours merges multiple contours into one
func (d *Detector) mergeContours(contours []Contour) Contour {
	var allPoints []image.Point
	
	for _, contour := range contours {
		allPoints = append(allPoints, contour.Points...)
	}
	
	merged := Contour{Points: allPoints}
	merged.calculateProperties()
	
	return merged
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}