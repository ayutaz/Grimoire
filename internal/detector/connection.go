package detector

import (
	"fmt"
	"image"
	"math"
	"os"
)

// detectConnections detects connections between symbols
func (d *Detector) detectConnections(binary *image.Gray, symbols []*Symbol) []Connection {
	connections := []Connection{}

	// Find lines using improved edge detection
	edges := d.improvedEdgeDetection(binary)
	lines := d.detectLines(edges)

	// Debug: save edge detection result
	if os.Getenv("GRIMOIRE_DEBUG") != "" {
		if err := d.DebugSaveImage(edges, "debug_edges.png"); err != nil {
			fmt.Printf("Failed to save debug image: %v\n", err)
		}
		fmt.Printf("Detected %d lines\n", len(lines))
	}

	// For each line, check if it connects symbols
	for _, line := range lines {
		// Find symbols near line endpoints
		fromSymbol := d.findNearestSymbol(line.Start, symbols)
		toSymbol := d.findNearestSymbol(line.End, symbols)

		if os.Getenv("GRIMOIRE_DEBUG") != "" && (fromSymbol != nil || toSymbol != nil) {
			fmt.Printf("Line (%d,%d)->(%d,%d): from=%v, to=%v\n",
				line.Start.X, line.Start.Y, line.End.X, line.End.Y,
				fromSymbol != nil, toSymbol != nil)
		}

		if fromSymbol != nil && toSymbol != nil && fromSymbol != toSymbol {
			// Validate connection
			if d.isValidConnection(line, fromSymbol, toSymbol) {
				// Determine connection type
				connType := d.determineConnectionType(line, binary)

				// Determine direction
				from, to := d.determineConnectionDirection(fromSymbol, toSymbol)

				conn := Connection{
					From:           from,
					To:             to,
					ConnectionType: connType,
					Properties:     make(map[string]interface{}),
				}

				connections = append(connections, conn)

				if os.Getenv("GRIMOIRE_DEBUG") != "" {
					fmt.Printf("Connection added: %s -> %s (%s)\n", from.Type, to.Type, connType)
				}
			} else if os.Getenv("GRIMOIRE_DEBUG") != "" {
				fmt.Printf("Connection invalid between %s and %s\n", fromSymbol.Type, toSymbol.Type)
			}
		}
	}

	return connections
}

// Line represents a detected line
type Line struct {
	Start image.Point
	End   image.Point
}


// detectLines detects lines using simplified Hough transform
func (d *Detector) detectLines(edges *image.Gray) []Line {
	lines := []Line{}

	// Use simple line segment detection
	lines = append(lines, d.detectHorizontalLines(edges)...)
	lines = append(lines, d.detectVerticalLines(edges)...)
	lines = append(lines, d.improvedDetectDiagonalLines(edges)...)

	// Merge connected line segments
	lines = d.mergeConnectedLines(lines)

	return lines
}

// detectHorizontalLines detects horizontal lines
func (d *Detector) detectHorizontalLines(edges *image.Gray) []Line {
	lines := []Line{}
	bounds := edges.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		lineStart := -1
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if edges.GrayAt(x, y).Y > 128 {
				if lineStart == -1 {
					lineStart = x
				}
			} else {
				if lineStart != -1 && x-lineStart > 20 { // Min line length
					lines = append(lines, Line{
						Start: image.Point{X: lineStart, Y: y},
						End:   image.Point{X: x - 1, Y: y},
					})
				}
				lineStart = -1
			}
		}
		// Handle line extending to edge
		if lineStart != -1 && bounds.Max.X-lineStart > 20 {
			lines = append(lines, Line{
				Start: image.Point{X: lineStart, Y: y},
				End:   image.Point{X: bounds.Max.X - 1, Y: y},
			})
		}
	}

	return lines
}

// detectVerticalLines detects vertical lines
func (d *Detector) detectVerticalLines(edges *image.Gray) []Line {
	lines := []Line{}
	bounds := edges.Bounds()

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		lineStart := -1
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if edges.GrayAt(x, y).Y > 128 {
				if lineStart == -1 {
					lineStart = y
				}
			} else {
				if lineStart != -1 && y-lineStart > 20 { // Min line length
					lines = append(lines, Line{
						Start: image.Point{X: x, Y: lineStart},
						End:   image.Point{X: x, Y: y - 1},
					})
				}
				lineStart = -1
			}
		}
		// Handle line extending to edge
		if lineStart != -1 && bounds.Max.Y-lineStart > 20 {
			lines = append(lines, Line{
				Start: image.Point{X: x, Y: lineStart},
				End:   image.Point{X: x, Y: bounds.Max.Y - 1},
			})
		}
	}

	return lines
}

// detectDiagonalLines detects diagonal lines using line following
func (d *Detector) detectDiagonalLines(edges *image.Gray) []Line {
	lines := []Line{}
	bounds := edges.Bounds()
	visited := make(map[image.Point]bool)

	// Scan for diagonal lines
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pt := image.Point{X: x, Y: y}
			if edges.GrayAt(x, y).Y > 128 && !visited[pt] {
				// Try to follow a diagonal line
				line := d.followDiagonalLine(edges, pt, visited)
				if line != nil {
					lines = append(lines, *line)
				}
			}
		}
	}

	return lines
}

// followDiagonalLine follows a diagonal line from a starting point
func (d *Detector) followDiagonalLine(edges *image.Gray, start image.Point, visited map[image.Point]bool) *Line {
	bounds := edges.Bounds()

	// Try each diagonal direction
	directions := []image.Point{
		{X: 1, Y: 1},   // Down-right
		{X: -1, Y: 1},  // Down-left
		{X: 1, Y: -1},  // Up-right
		{X: -1, Y: -1}, // Up-left
	}

	for _, dir := range directions {
		points := []image.Point{start}
		current := start
		visited[current] = true

		// Follow the line in this direction
		for {
			next := image.Point{X: current.X + dir.X, Y: current.Y + dir.Y}

			// Check bounds
			if next.X < bounds.Min.X || next.X >= bounds.Max.X ||
				next.Y < bounds.Min.Y || next.Y >= bounds.Max.Y {
				break
			}

			// Check if there's an edge pixel
			found := false

			// Check exact position and nearby pixels (allow some deviation)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					checkPt := image.Point{X: next.X + dx, Y: next.Y + dy}
					if checkPt.X >= bounds.Min.X && checkPt.X < bounds.Max.X &&
						checkPt.Y >= bounds.Min.Y && checkPt.Y < bounds.Max.Y &&
						edges.GrayAt(checkPt.X, checkPt.Y).Y > 128 && !visited[checkPt] {
						next = checkPt
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found {
				break
			}

			visited[next] = true
			points = append(points, next)
			current = next
		}

		// If we found a diagonal line of sufficient length
		if len(points) > 15 { // Min diagonal line length
			return &Line{
				Start: points[0],
				End:   points[len(points)-1],
			}
		}
	}

	return nil
}

// mergeConnectedLines merges line segments that are connected
func (d *Detector) mergeConnectedLines(lines []Line) []Line {
	if len(lines) == 0 {
		return lines
	}

	merged := []Line{}
	used := make([]bool, len(lines))

	for i := 0; i < len(lines); i++ {
		if used[i] {
			continue
		}

		current := lines[i]
		used[i] = true

		// Try to extend this line by finding connected segments
		extended := true
		for extended {
			extended = false

			for j := 0; j < len(lines); j++ {
				if used[j] {
					continue
				}

				// Check if lines are connected and aligned
				if d.linesConnected(current, lines[j]) && d.linesAligned(current, lines[j]) {
					// Merge the lines
					current = d.mergeLines(current, lines[j])
					used[j] = true
					extended = true
				}
			}
		}

		merged = append(merged, current)
	}

	return merged
}

// linesConnected checks if two lines are connected (endpoints close)
func (d *Detector) linesConnected(l1, l2 Line) bool {
	threshold := 5.0

	// Check all endpoint combinations
	dist1 := distance(l1.End, l2.Start)
	dist2 := distance(l1.End, l2.End)
	dist3 := distance(l1.Start, l2.Start)
	dist4 := distance(l1.Start, l2.End)

	return dist1 < threshold || dist2 < threshold || dist3 < threshold || dist4 < threshold
}

// linesAligned checks if two lines are roughly aligned
func (d *Detector) linesAligned(l1, l2 Line) bool {
	// Calculate angles
	angle1 := math.Atan2(float64(l1.End.Y-l1.Start.Y), float64(l1.End.X-l1.Start.X))
	angle2 := math.Atan2(float64(l2.End.Y-l2.Start.Y), float64(l2.End.X-l2.Start.X))

	// Normalize angle difference
	diff := math.Abs(angle1 - angle2)
	if diff > math.Pi {
		diff = 2*math.Pi - diff
	}

	// Allow 30 degree deviation
	return diff < math.Pi/6
}

// mergeLines merges two connected lines into one
func (d *Detector) mergeLines(l1, l2 Line) Line {
	// Find the two points that are farthest apart
	points := []image.Point{l1.Start, l1.End, l2.Start, l2.End}
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

	return Line{Start: start, End: end}
}

// findNearestSymbol finds the nearest symbol to a point
func (d *Detector) findNearestSymbol(point image.Point, symbols []*Symbol) *Symbol {
	var nearest *Symbol
	minDist := 50.0 // Increased threshold to find more connections

	for _, symbol := range symbols {
		dist := math.Sqrt(math.Pow(float64(point.X)-symbol.Position.X, 2) +
			math.Pow(float64(point.Y)-symbol.Position.Y, 2))

		if dist < minDist {
			minDist = dist
			nearest = symbol
		}
	}

	return nearest
}

// isValidConnection validates if a line represents a valid connection
func (d *Detector) isValidConnection(line Line, from, to *Symbol) bool {
	// Skip connections to/from outer circle
	if from.Type == OuterCircle || to.Type == OuterCircle {
		return false
	}

	// Check if line endpoints are near symbols
	startDist := math.Sqrt(math.Pow(float64(line.Start.X)-from.Position.X, 2) +
		math.Pow(float64(line.Start.Y)-from.Position.Y, 2))
	endDist := math.Sqrt(math.Pow(float64(line.End.X)-to.Position.X, 2) +
		math.Pow(float64(line.End.Y)-to.Position.Y, 2))

	// More lenient distance check
	// Line endpoint should be within reasonable distance of symbol center
	maxDistFrom := math.Max(from.Size*2.0, 60.0)
	maxDistTo := math.Max(to.Size*2.0, 60.0)

	if startDist > maxDistFrom || endDist > maxDistTo {
		return false
	}

	// Ensure line is long enough to be a real connection
	lineLength := math.Sqrt(math.Pow(float64(line.End.X-line.Start.X), 2) +
		math.Pow(float64(line.End.Y-line.Start.Y), 2))
	return lineLength >= 20
}

// determineConnectionType determines the type of connection
func (d *Detector) determineConnectionType(line Line, binary *image.Gray) string {
	// Sample points along the line
	dx := line.End.X - line.Start.X
	dy := line.End.Y - line.Start.Y
	length := math.Sqrt(float64(dx*dx + dy*dy))

	if length == 0 {
		return "solid"
	}

	// Sample 10 points along the line
	transitions := 0
	lastPixel := false

	for i := 0; i <= 10; i++ {
		t := float64(i) / 10.0
		x := int(float64(line.Start.X) + t*float64(dx))
		y := int(float64(line.Start.Y) + t*float64(dy))

		pixel := binary.GrayAt(x, y).Y > 128
		if i > 0 && pixel != lastPixel {
			transitions++
		}
		lastPixel = pixel
	}

	// Classify based on transitions
	if transitions <= 2 {
		return "solid"
	} else if transitions <= 10 {
		return "dashed"
	} else {
		return "dotted"
	}
}

// determineConnectionDirection determines the flow direction
func (d *Detector) determineConnectionDirection(sym1, sym2 *Symbol) (*Symbol, *Symbol) {
	// Rules:
	// 1. Data types (squares/circles) -> Operators
	// 2. Operators -> Outputs (stars)
	// 3. Functions -> Outputs
	// 4. Main (double circle) -> Statements
	// 5. Position-based: top-to-bottom, left-to-right

	// Check symbol types
	type1IsData := sym1.Type == Square || sym1.Type == Circle
	type2IsData := sym2.Type == Square || sym2.Type == Circle

	type1IsOperator := isOperatorType(sym1.Type)
	type2IsOperator := isOperatorType(sym2.Type)

	type1IsOutput := sym1.Type == Star
	type2IsOutput := sym2.Type == Star

	switch {
	// Data -> Operator
	case type1IsData && type2IsOperator:
		return sym1, sym2
	case type2IsData && type1IsOperator:
		return sym2, sym1

	// Operator -> Output
	case type1IsOperator && type2IsOutput:
		return sym1, sym2
	case type2IsOperator && type1IsOutput:
		return sym2, sym1
	}

	// Function -> Output
	if sym1.Type == Circle && type2IsOutput {
		return sym1, sym2
	}
	if sym2.Type == Circle && type1IsOutput {
		return sym2, sym1
	}

	// Main -> Statement
	if sym1.Type == DoubleCircle {
		return sym1, sym2
	}
	if sym2.Type == DoubleCircle {
		return sym2, sym1
	}

	// Default: position-based (top-to-bottom, left-to-right)
	if sym1.Position.Y < sym2.Position.Y-10 {
		return sym1, sym2
	}
	if sym2.Position.Y < sym1.Position.Y-10 {
		return sym2, sym1
	}

	if sym1.Position.X < sym2.Position.X {
		return sym1, sym2
	}

	return sym2, sym1
}

// isOperatorType checks if a symbol type is an operator
func isOperatorType(t SymbolType) bool {
	return t == Convergence || t == Divergence ||
		t == Amplification || t == Distribution ||
		t == Transfer || t == Seal || t == Circulation ||
		t == Equal || t == NotEqual ||
		t == LessThan || t == GreaterThan ||
		t == LessEqual || t == GreaterEqual ||
		t == LogicalAnd || t == LogicalOr ||
		t == LogicalNot || t == LogicalXor
}
