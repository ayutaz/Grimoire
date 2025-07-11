package detector

import (
	"image"
	"image/color"
	"math"
)

// detectConnections detects connections between symbols
func (d *Detector) detectConnections(binary *image.Gray, symbols []*Symbol) []Connection {
	connections := []Connection{}
	
	// Find lines using edge detection
	edges := d.detectEdges(binary)
	lines := d.detectLines(edges)
	
	// For each line, check if it connects symbols
	for _, line := range lines {
		// Find symbols near line endpoints
		fromSymbol := d.findNearestSymbol(line.Start, symbols)
		toSymbol := d.findNearestSymbol(line.End, symbols)
		
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

// detectEdges performs edge detection
func (d *Detector) detectEdges(binary *image.Gray) *image.Gray {
	bounds := binary.Bounds()
	edges := image.NewGray(bounds)
	
	// Simple edge detection using gradient
	for y := bounds.Min.Y + 1; y < bounds.Max.Y - 1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X - 1; x++ {
			// Sobel operator
			gx := int(binary.GrayAt(x+1, y-1).Y) + 2*int(binary.GrayAt(x+1, y).Y) + int(binary.GrayAt(x+1, y+1).Y) -
				int(binary.GrayAt(x-1, y-1).Y) - 2*int(binary.GrayAt(x-1, y).Y) - int(binary.GrayAt(x-1, y+1).Y)
			
			gy := int(binary.GrayAt(x-1, y+1).Y) + 2*int(binary.GrayAt(x, y+1).Y) + int(binary.GrayAt(x+1, y+1).Y) -
				int(binary.GrayAt(x-1, y-1).Y) - 2*int(binary.GrayAt(x, y-1).Y) - int(binary.GrayAt(x+1, y-1).Y)
			
			magnitude := int(math.Sqrt(float64(gx*gx + gy*gy)))
			if magnitude > 100 {
				edges.Set(x, y, color.Gray{255})
			} else {
				edges.Set(x, y, color.Gray{0})
			}
		}
	}
	
	return edges
}

// detectLines detects lines using Hough transform
func (d *Detector) detectLines(edges *image.Gray) []Line {
	lines := []Line{}
	bounds := edges.Bounds()
	
	// Simplified line detection - scan for horizontal and vertical lines
	// Horizontal lines
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		lineStart := -1
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if edges.GrayAt(x, y).Y > 128 {
				if lineStart == -1 {
					lineStart = x
				}
			} else {
				if lineStart != -1 && x - lineStart > 20 { // Min line length
					lines = append(lines, Line{
						Start: image.Point{X: lineStart, Y: y},
						End:   image.Point{X: x - 1, Y: y},
					})
				}
				lineStart = -1
			}
		}
	}
	
	// Vertical lines
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		lineStart := -1
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if edges.GrayAt(x, y).Y > 128 {
				if lineStart == -1 {
					lineStart = y
				}
			} else {
				if lineStart != -1 && y - lineStart > 20 { // Min line length
					lines = append(lines, Line{
						Start: image.Point{X: x, Y: lineStart},
						End:   image.Point{X: x, Y: y - 1},
					})
				}
				lineStart = -1
			}
		}
	}
	
	// TODO: Add diagonal line detection
	
	return lines
}

// findNearestSymbol finds the nearest symbol to a point
func (d *Detector) findNearestSymbol(point image.Point, symbols []*Symbol) *Symbol {
	var nearest *Symbol
	minDist := 30.0 // Max distance threshold
	
	for _, symbol := range symbols {
		dist := math.Sqrt(math.Pow(float64(point.X) - symbol.Position.X, 2) +
			math.Pow(float64(point.Y) - symbol.Position.Y, 2))
		
		if dist < minDist {
			minDist = dist
			nearest = symbol
		}
	}
	
	return nearest
}

// isValidConnection validates if a line represents a valid connection
func (d *Detector) isValidConnection(line Line, from, to *Symbol) bool {
	// Check if line endpoints are near symbol edges (not centers)
	startDist := math.Sqrt(math.Pow(float64(line.Start.X) - from.Position.X, 2) +
		math.Pow(float64(line.Start.Y) - from.Position.Y, 2))
	endDist := math.Sqrt(math.Pow(float64(line.End.X) - to.Position.X, 2) +
		math.Pow(float64(line.End.Y) - to.Position.Y, 2))
	
	// Distance should be between 30% and 120% of symbol size
	minDist := from.Size * 0.3
	maxDist := from.Size * 1.2
	
	if startDist < minDist || startDist > maxDist {
		return false
	}
	
	minDist = to.Size * 0.3
	maxDist = to.Size * 1.2
	
	if endDist < minDist || endDist > maxDist {
		return false
	}
	
	return true
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
		x := int(float64(line.Start.X) + t * float64(dx))
		y := int(float64(line.Start.Y) + t * float64(dy))
		
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
	
	// Data -> Operator
	if type1IsData && type2IsOperator {
		return sym1, sym2
	}
	if type2IsData && type1IsOperator {
		return sym2, sym1
	}
	
	// Operator -> Output
	if type1IsOperator && type2IsOutput {
		return sym1, sym2
	}
	if type2IsOperator && type1IsOutput {
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
	if sym1.Position.Y < sym2.Position.Y - 10 {
		return sym1, sym2
	}
	if sym2.Position.Y < sym1.Position.Y - 10 {
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