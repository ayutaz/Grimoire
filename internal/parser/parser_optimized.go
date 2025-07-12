package parser

import (
	"math"
	"sort"
	"sync"

	"github.com/ayutaz/grimoire/internal/detector"
)

// OptimizedParser uses optimized algorithms for parsing
type OptimizedParser struct {
	*Parser
	spatialIndex *SpatialIndex
}

// SpatialIndex provides fast spatial lookups for symbols
type SpatialIndex struct {
	symbols    []*detector.Symbol
	nodes      []*symbolNode
	gridSize   float64
	grid       map[gridKey][]*symbolNode
}

type gridKey struct {
	x, y int
}

// NewOptimizedParser creates a new optimized parser
func NewOptimizedParser() *OptimizedParser {
	return &OptimizedParser{
		Parser: NewParser(),
	}
}

// Parse performs optimized parsing
func (p *OptimizedParser) Parse(symbols []*detector.Symbol, connections []detector.Connection) (*Program, error) {
	p.symbols = symbols
	p.connections = connections

	// Validate input
	if len(symbols) == 0 {
		return nil, p.createError("No symbols to parse", "The input contains no detected symbols")
	}

	// Build spatial index for fast lookups
	p.buildSpatialIndex()

	// Build symbol graph with optimizations
	p.buildOptimizedSymbolGraph()

	// Rest of parsing logic is same as original
	return p.Parser.Parse(symbols, connections)
}

// buildSpatialIndex creates a spatial index for fast symbol lookups
func (p *OptimizedParser) buildSpatialIndex() {
	if len(p.symbols) == 0 {
		return
	}

	// Calculate optimal grid size
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, sym := range p.symbols {
		minX = math.Min(minX, sym.Position.X)
		minY = math.Min(minY, sym.Position.Y)
		maxX = math.Max(maxX, sym.Position.X)
		maxY = math.Max(maxY, sym.Position.Y)
	}

	// Use grid size that creates ~10 cells per dimension
	gridSize := math.Max((maxX-minX)/10, (maxY-minY)/10)
	if gridSize < 50 {
		gridSize = 50
	}

	p.spatialIndex = &SpatialIndex{
		symbols:  p.symbols,
		nodes:    make([]*symbolNode, len(p.symbols)),
		gridSize: gridSize,
		grid:     make(map[gridKey][]*symbolNode),
	}

	// Create nodes and add to grid
	for i, symbol := range p.symbols {
		node := &symbolNode{
			symbol:   symbol,
			children: []*symbolNode{},
		}
		p.symbolGraph[i] = node
		p.spatialIndex.nodes[i] = node

		// Add to grid
		key := p.spatialIndex.getGridKey(symbol.Position)
		p.spatialIndex.grid[key] = append(p.spatialIndex.grid[key], node)
	}
}

// getGridKey returns the grid cell for a position
func (si *SpatialIndex) getGridKey(pos detector.Position) gridKey {
	return gridKey{
		x: int(pos.X / si.gridSize),
		y: int(pos.Y / si.gridSize),
	}
}

// getNearbyNodes returns nodes within a certain distance
func (si *SpatialIndex) getNearbyNodes(pos detector.Position, maxDist float64) []*symbolNode {
	var result []*symbolNode
	
	// Calculate grid cells to check
	cellRadius := int(math.Ceil(maxDist / si.gridSize))
	centerKey := si.getGridKey(pos)
	
	for dx := -cellRadius; dx <= cellRadius; dx++ {
		for dy := -cellRadius; dy <= cellRadius; dy++ {
			key := gridKey{x: centerKey.x + dx, y: centerKey.y + dy}
			if nodes, ok := si.grid[key]; ok {
				for _, node := range nodes {
					dist := distance(pos, node.symbol.Position)
					if dist <= maxDist*maxDist { // Using squared distance
						result = append(result, node)
					}
				}
			}
		}
	}
	
	return result
}

// buildOptimizedSymbolGraph builds the symbol graph with optimizations
func (p *OptimizedParser) buildOptimizedSymbolGraph() {
	// Use explicit connections if available
	if len(p.connections) > 0 {
		p.applyConnectionsOptimized()
	} else {
		// Otherwise use optimized inference
		p.inferConnectionsOptimized()
	}
}

// applyConnectionsOptimized applies connections in parallel
func (p *OptimizedParser) applyConnectionsOptimized() {
	// Build symbol to index map for O(1) lookups
	symbolToIndex := make(map[*detector.Symbol]int, len(p.symbols))
	for i, sym := range p.symbols {
		symbolToIndex[sym] = i
	}

	// Apply connections
	for _, conn := range p.connections {
		if fromIdx, ok := symbolToIndex[conn.From]; ok {
			if toIdx, ok := symbolToIndex[conn.To]; ok {
				fromNode := p.symbolGraph[fromIdx]
				toNode := p.symbolGraph[toIdx]
				fromNode.children = append(fromNode.children, toNode)
				toNode.parent = fromNode
			}
		}
	}
}

// inferConnectionsOptimized uses spatial indexing for faster connection inference
func (p *OptimizedParser) inferConnectionsOptimized() {
	// Skip outer circle
	symbolsToConnect := make([]*symbolNode, 0, len(p.symbols))
	for i, sym := range p.symbols {
		if sym.Type != detector.OuterCircle {
			symbolsToConnect = append(symbolsToConnect, p.symbolGraph[i])
		}
	}

	// Process connections in parallel batches
	var wg sync.WaitGroup
	batchSize := len(symbolsToConnect) / 4
	if batchSize < 10 {
		batchSize = 10
	}

	for start := 0; start < len(symbolsToConnect); start += batchSize {
		end := start + batchSize
		if end > len(symbolsToConnect) {
			end = len(symbolsToConnect)
		}

		wg.Add(1)
		go func(nodes []*symbolNode) {
			defer wg.Done()
			p.processNodeBatch(nodes)
		}(symbolsToConnect[start:end])
	}

	wg.Wait()
}

// processNodeBatch processes a batch of nodes for connections
func (p *OptimizedParser) processNodeBatch(nodes []*symbolNode) {
	for _, node := range nodes {
		switch node.symbol.Type {
		case detector.DoubleCircle:
			p.connectMainEntry(node)
		case detector.Star:
			p.connectStar(node)
		default:
			if isOperator(node.symbol.Type) {
				p.connectOperator(node)
			}
		}
	}
}

// connectMainEntry connects main entry to symbols below it
func (p *OptimizedParser) connectMainEntry(node *symbolNode) {
	mainY := node.symbol.Position.Y
	
	// Find symbols below main using spatial index
	candidates := p.spatialIndex.getNearbyNodes(
		detector.Position{X: node.symbol.Position.X, Y: mainY + 100},
		150,
	)
	
	for _, other := range candidates {
		if other != node && other.symbol.Position.Y > mainY {
			// Check horizontal alignment
			xDiff := math.Abs(other.symbol.Position.X - node.symbol.Position.X)
			if xDiff < 150 {
				node.children = append(node.children, other)
				other.parent = node
			}
		}
	}
}

// connectOperator connects operators to nearby operands
func (p *OptimizedParser) connectOperator(node *symbolNode) {
	// Find nearby squares using spatial index
	nearbyNodes := p.spatialIndex.getNearbyNodes(node.symbol.Position, 150)
	
	var operands []*symbolNode
	for _, other := range nearbyNodes {
		if other != node && other.symbol.Type == detector.Square {
			operands = append(operands, other)
		}
	}
	
	// Sort by distance to get closest operands
	sort.Slice(operands, func(i, j int) bool {
		dist1 := distance(node.symbol.Position, operands[i].symbol.Position)
		dist2 := distance(node.symbol.Position, operands[j].symbol.Position)
		return dist1 < dist2
	})
	
	// Connect closest operands (usually 2 for binary operators)
	maxOperands := 2
	if node.symbol.Type == detector.LogicalNot {
		maxOperands = 1
	}
	
	for i := 0; i < len(operands) && i < maxOperands; i++ {
		operands[i].children = append(operands[i].children, node)
	}
}

// connectStar connects stars to nearest expressions above
func (p *OptimizedParser) connectStar(node *symbolNode) {
	starPos := node.symbol.Position
	
	// Find symbols above using spatial index
	searchPos := detector.Position{X: starPos.X, Y: starPos.Y - 75}
	candidates := p.spatialIndex.getNearbyNodes(searchPos, 150)
	
	var nearest *symbolNode
	minDist := math.MaxFloat64
	
	for _, other := range candidates {
		if other != node && other.symbol.Position.Y < starPos.Y {
			dist := distance(starPos, other.symbol.Position)
			if dist < minDist {
				minDist = dist
				nearest = other
			}
		}
	}
	
	if nearest != nil && minDist < 150*150 { // Using squared distance
		nearest.children = append(nearest.children, node)
		node.parent = nearest
	}
}

// Optimized expression parsing with memoization
type ExpressionCache struct {
	mu    sync.RWMutex
	cache map[*symbolNode]Expression
}

func NewExpressionCache() *ExpressionCache {
	return &ExpressionCache{
		cache: make(map[*symbolNode]Expression),
	}
}

func (ec *ExpressionCache) get(node *symbolNode) (Expression, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	expr, ok := ec.cache[node]
	return expr, ok
}

func (ec *ExpressionCache) set(node *symbolNode, expr Expression) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.cache[node] = expr
}

// createError creates an error with proper formatting
func (p *OptimizedParser) createError(msg, details string) error {
	return p.Parser.errors[0] // Reuse parser's error creation logic
}