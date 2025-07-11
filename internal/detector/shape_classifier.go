package detector

import (
	"image"
	"math"
)

// classifyShape classifies a contour into different shape types
func (d *Detector) classifyShape(contour Contour) SymbolType {
	// Check for circle first
	if contour.isCircle(d.circleThreshold) {
		// Check if it's the outer circle by size and perimeter
		// Outer circle should be large
		if (contour.Area > 5000 || contour.Perimeter > 500) && d.isOuterCircle(contour) {
			return OuterCircle
		}
		// Check for double circle
		if d.isDoubleCircle(contour) {
			return DoubleCircle
		}
		return Circle
	}
	
	// Check for star shape before polygon approximation
	if d.isStarShape(contour) {
		return Star
	}
	
	// Approximate the contour to a polygon
	approx := d.approximatePolygon(contour)
	vertices := len(approx)
	
	switch vertices {
	case 3:
		return Triangle
	case 4:
		if d.isSquare(approx) {
			return Square
		}
		// Could be other quadrilateral
		return Unknown
	case 5:
		return Pentagon
	case 6:
		return Hexagon
	case 8:
		if d.isStar(approx, 4) {
			return Star
		}
		return Unknown
	case 10:
		if d.isStar(approx, 5) {
			return Star
		}
		return Unknown
	case 12:
		if d.isStar(approx, 6) {
			return SixPointedStar
		}
		return Unknown
	case 16:
		if d.isStar(approx, 8) {
			return EightPointedStar
		}
		return Unknown
	}
	
	// Check for special operator symbols
	if symbolType := d.classifyOperator(contour); symbolType != Unknown {
		return symbolType
	}
	
	return Unknown
}

// approximatePolygon approximates a contour with a polygon
func (d *Detector) approximatePolygon(contour Contour) []image.Point {
	if len(contour.Points) < 3 {
		return contour.Points
	}
	
	// Douglas-Peucker algorithm for polygon approximation
	epsilon := contour.Perimeter * 0.02 // 2% of perimeter
	return d.douglasPeucker(contour.Points, epsilon)
}

// douglasPeucker implements the Douglas-Peucker algorithm
func (d *Detector) douglasPeucker(points []image.Point, epsilon float64) []image.Point {
	if len(points) <= 2 {
		return points
	}
	
	// Find the point with maximum distance from the line
	maxDist := 0.0
	maxIndex := 0
	
	for i := 1; i < len(points)-1; i++ {
		dist := d.perpendicularDistance(points[i], points[0], points[len(points)-1])
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}
	
	// If max distance is greater than epsilon, recursively simplify
	if maxDist > epsilon {
		// Recursively call on both parts
		left := d.douglasPeucker(points[:maxIndex+1], epsilon)
		right := d.douglasPeucker(points[maxIndex:], epsilon)
		
		// Build the result
		result := make([]image.Point, 0, len(left)+len(right)-1)
		result = append(result, left[:len(left)-1]...)
		result = append(result, right...)
		return result
	}
	
	// Otherwise, return just the endpoints
	return []image.Point{points[0], points[len(points)-1]}
}

// perpendicularDistance calculates perpendicular distance from point to line
func (d *Detector) perpendicularDistance(point, lineStart, lineEnd image.Point) float64 {
	// Calculate line parameters
	dx := float64(lineEnd.X - lineStart.X)
	dy := float64(lineEnd.Y - lineStart.Y)
	
	// Handle vertical and horizontal lines
	if dx == 0 && dy == 0 {
		return distance(point, lineStart)
	}
	
	// Calculate perpendicular distance
	numerator := math.Abs(dy*float64(point.X) - dx*float64(point.Y) + 
		float64(lineEnd.X)*float64(lineStart.Y) - float64(lineEnd.Y)*float64(lineStart.X))
	denominator := math.Sqrt(dx*dx + dy*dy)
	
	return numerator / denominator
}

// isSquare checks if a quadrilateral is a square
func (d *Detector) isSquare(vertices []image.Point) bool {
	if len(vertices) != 4 {
		return false
	}
	
	// Check if all sides are approximately equal
	sides := make([]float64, 4)
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		sides[i] = distance(vertices[i], vertices[j])
	}
	
	// Check if all sides are within 10% of each other
	minSide := sides[0]
	maxSide := sides[0]
	for _, side := range sides {
		if side < minSide {
			minSide = side
		}
		if side > maxSide {
			maxSide = side
		}
	}
	
	return (maxSide-minSide)/minSide < 0.1
}

// isStar checks if vertices form a star pattern
func (d *Detector) isStar(vertices []image.Point, expectedPoints int) bool {
	if len(vertices) != expectedPoints*2 {
		return false
	}
	
	// Calculate center
	centerX, centerY := 0, 0
	for _, v := range vertices {
		centerX += v.X
		centerY += v.Y
	}
	center := image.Point{X: centerX / len(vertices), Y: centerY / len(vertices)}
	
	// Calculate distances from center
	distances := make([]float64, len(vertices))
	for i, v := range vertices {
		distances[i] = distance(v, center)
	}
	
	// Check alternating pattern (inner and outer points)
	for i := 0; i < len(distances); i += 2 {
		outer1 := distances[i]
		inner := distances[(i+1)%len(distances)]
		outer2 := distances[(i+2)%len(distances)]
		
		// Inner points should be closer to center than outer points
		if inner > outer1*0.7 || inner > outer2*0.7 {
			return false
		}
	}
	
	return true
}

// isDoubleCircle checks if a circular contour is actually a double circle
func (d *Detector) isDoubleCircle(contour Contour) bool {
	// A double circle will have a specific thickness pattern
	// This is a simplified check - in reality, we'd need to analyze the thickness
	bbox := contour.getBoundingBox()
	
	// Check if the contour has significant thickness relative to its size
	expectedArea := math.Pi * math.Pow(float64(bbox.Dx())/2, 2)
	actualArea := contour.Area
	
	// Double circle will have less area than a filled circle
	return actualArea < expectedArea*0.3 && actualArea > expectedArea*0.1
}

// classifyOperator attempts to classify special operator symbols
func (d *Detector) classifyOperator(contour Contour) SymbolType {
	// This is a placeholder - real implementation would analyze
	// the specific shape patterns for each operator
	
	// Check aspect ratio and area for operator detection
	aspectRatio := contour.getAspectRatio()
	
	// Operators tend to have specific aspect ratios
	if aspectRatio > 2.5 {
		// Possibly a transfer operator (horizontal arrow)
		return Transfer
	}
	
	// More complex operator detection would go here
	return Unknown
}

// isStarShape checks if a contour is star-shaped
func (d *Detector) isStarShape(contour Contour) bool {
	// Method 1: Check vertex count after approximation
	approx := d.approximatePolygon(contour)
	numVertices := len(approx)
	
	// Stars typically have 8-12 vertices (5 points + 5 inner vertices)
	if numVertices >= 8 && numVertices <= 12 {
		return true
	}
	
	// Method 2: Check for significant variation in distances from center
	if len(contour.Points) < 5 {
		return false
	}
	
	center := contour.Center
	distances := make([]float64, len(contour.Points))
	var sum float64
	
	for i, pt := range contour.Points {
		dist := distance(pt, center)
		distances[i] = dist
		sum += dist
	}
	
	mean := sum / float64(len(distances))
	
	// Calculate standard deviation
	var variance float64
	for _, dist := range distances {
		variance += (dist - mean) * (dist - mean)
	}
	stdDev := math.Sqrt(variance / float64(len(distances)))
	
	// Star has significant variation (std > 15% of mean)
	return stdDev > mean*0.15
}