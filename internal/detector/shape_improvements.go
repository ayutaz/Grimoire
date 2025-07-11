package detector

import (
	"image"
	"math"
)

// improveSquareDetection adds additional checks for better square detection
func (d *Detector) improveSquareDetection(contour Contour) bool {
	// Method 1: Corner detection
	if d.hasSquareCorners(contour) {
		return true
	}

	// Method 2: Edge straightness check
	if d.hasStraightEdges(contour) {
		return true
	}

	// Method 3: Symmetry check
	if d.hasSquareSymmetry(contour) {
		return true
	}

	return false
}

// hasSquareCorners checks if contour has 4 distinct corners
func (d *Detector) hasSquareCorners(contour Contour) bool {
	if len(contour.Points) < 8 {
		return false
	}

	// Find corner points using curvature
	corners := d.findCorners(contour.Points)

	// Should have exactly 4 corners for a square
	if len(corners) != 4 {
		return false
	}

	// Check if corners form a square
	return d.cornersFormSquare(corners)
}

// findCorners detects corner points in a contour
func (d *Detector) findCorners(points []image.Point) []image.Point {
	if len(points) < 5 {
		return nil
	}

	var corners []image.Point
	n := len(points)

	// Calculate curvature at each point
	for i := 0; i < n; i++ {
		prev := points[(i-2+n)%n]
		curr := points[i]
		next := points[(i+2)%n]

		// Calculate angle change
		angle1 := math.Atan2(float64(curr.Y-prev.Y), float64(curr.X-prev.X))
		angle2 := math.Atan2(float64(next.Y-curr.Y), float64(next.X-curr.X))

		angleChange := math.Abs(angle2 - angle1)
		if angleChange > math.Pi {
			angleChange = 2*math.Pi - angleChange
		}

		// Corner detection threshold
		if angleChange > math.Pi/3 { // 60 degrees
			corners = append(corners, curr)
		}
	}

	// Remove corners that are too close to each other
	return d.filterCloseCorners(corners, 10)
}

// filterCloseCorners removes corners that are too close to each other
func (d *Detector) filterCloseCorners(corners []image.Point, minDist float64) []image.Point {
	if len(corners) <= 4 {
		return corners
	}

	var filtered []image.Point
	for i, corner := range corners {
		tooClose := false
		for j := i + 1; j < len(corners); j++ {
			if distance(corner, corners[j]) < minDist {
				tooClose = true
				break
			}
		}
		if !tooClose {
			filtered = append(filtered, corner)
		}
	}

	return filtered
}

// cornersFormSquare checks if 4 corners form a square shape
func (d *Detector) cornersFormSquare(corners []image.Point) bool {
	if len(corners) != 4 {
		return false
	}

	// Calculate all sides
	sides := make([]float64, 4)
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		sides[i] = distance(corners[i], corners[j])
	}

	// Check if all sides are similar
	avgSide := (sides[0] + sides[1] + sides[2] + sides[3]) / 4
	for _, side := range sides {
		if math.Abs(side-avgSide)/avgSide > 0.25 { // 25% tolerance
			return false
		}
	}

	// Check diagonals are similar
	diag1 := distance(corners[0], corners[2])
	diag2 := distance(corners[1], corners[3])
	if math.Abs(diag1-diag2)/diag1 > 0.2 { // 20% tolerance
		return false
	}

	return true
}

// hasStraightEdges checks if contour has straight edges
func (d *Detector) hasStraightEdges(contour Contour) bool {
	if len(contour.Points) < 8 {
		return false
	}

	// Divide contour into 4 segments
	n := len(contour.Points)
	segmentSize := n / 4

	straightCount := 0
	for i := 0; i < 4; i++ {
		start := i * segmentSize
		end := (i + 1) * segmentSize
		if end >= n {
			end = n - 1
		}

		if d.isSegmentStraight(contour.Points[start:end]) {
			straightCount++
		}
	}

	// At least 3 out of 4 segments should be straight
	return straightCount >= 3
}

// isSegmentStraight checks if a segment of points forms a straight line
func (d *Detector) isSegmentStraight(points []image.Point) bool {
	if len(points) < 3 {
		return true
	}

	// Fit a line to the points
	start := points[0]
	end := points[len(points)-1]

	// Check deviation of intermediate points from the line
	maxDeviation := 0.0
	for i := 1; i < len(points)-1; i++ {
		deviation := d.perpendicularDistance(points[i], start, end)
		if deviation > maxDeviation {
			maxDeviation = deviation
		}
	}

	// Line is straight if max deviation is small
	lineLength := distance(start, end)
	return maxDeviation < lineLength*0.1 // 10% of line length
}

// hasSquareSymmetry checks if contour has square-like symmetry
func (d *Detector) hasSquareSymmetry(contour Contour) bool {
	// Check aspect ratio
	aspectRatio := contour.getAspectRatio()
	if aspectRatio < 0.85 || aspectRatio > 1.15 {
		return false
	}

	// Check horizontal and vertical symmetry
	center := contour.Center
	bbox := contour.getBoundingBox()

	// Count points in each quadrant
	q1, q2, q3, q4 := 0, 0, 0, 0
	for _, pt := range contour.Points {
		if pt.X >= center.X && pt.Y >= center.Y {
			q1++ // Bottom-right
		} else if pt.X < center.X && pt.Y >= center.Y {
			q2++ // Bottom-left
		} else if pt.X < center.X && pt.Y < center.Y {
			q3++ // Top-left
		} else {
			q4++ // Top-right
		}
	}

	// Check if quadrants have similar number of points
	total := float64(q1 + q2 + q3 + q4)

	// Each quadrant should have roughly 25% of points
	for _, count := range []int{q1, q2, q3, q4} {
		ratio := float64(count) / total
		if math.Abs(ratio-0.25) > 0.1 { // 10% tolerance
			return false
		}
	}

	// Check fill ratio
	bboxArea := float64(bbox.Dx() * bbox.Dy())
	fillRatio := contour.Area / bboxArea

	// Squares typically fill 85-95% of their bounding box
	return fillRatio > 0.85 && fillRatio < 0.95
}
