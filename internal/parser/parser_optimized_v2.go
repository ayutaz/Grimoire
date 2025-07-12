package parser

import (
	"math"
	"sync"

	"github.com/ayutaz/grimoire/internal/detector"
	grimoireErrors "github.com/ayutaz/grimoire/internal/errors"
)

// OptimizedParserV2 is an improved version of the optimized parser
type OptimizedParserV2 struct {
	*Parser
	spatialIndex   *AdaptiveSpatialIndex
	symbolCache    map[*detector.Symbol]int
	connectionPool *sync.Pool
}

// AdaptiveSpatialIndex adapts grid size based on symbol density
type AdaptiveSpatialIndex struct {
	symbols     []*detector.Symbol
	nodes       []*symbolNode
	gridSize    float64
	grid        map[gridKey][]*symbolNode
	useQuadTree bool
	quadTree    *QuadTree
	symbolCount int
}

// QuadTree for more efficient spatial indexing at large scales
type QuadTree struct {
	bounds    Rectangle
	maxDepth  int
	maxNodes  int
	depth     int
	nodes     []*symbolNode
	northWest *QuadTree
	northEast *QuadTree
	southWest *QuadTree
	southEast *QuadTree
}

type Rectangle struct {
	x, y, width, height float64
}

// NewOptimizedParserV2 creates an improved optimized parser
func NewOptimizedParserV2() *OptimizedParserV2 {
	return &OptimizedParserV2{
		Parser:      NewParser(),
		symbolCache: make(map[*detector.Symbol]int),
		connectionPool: &sync.Pool{
			New: func() interface{} {
				return &symbolNode{
					children: make([]*symbolNode, 0, 4),
				}
			},
		},
	}
}

// Parse performs optimized parsing with adaptive strategies
func (p *OptimizedParserV2) Parse(symbols []*detector.Symbol, connections []detector.Connection) (*Program, error) {
	// For small symbol counts, use the standard parser
	if len(symbols) < 50 {
		return p.Parser.Parse(symbols, connections)
	}

	p.symbols = symbols
	p.connections = connections

	// Validate input
	if len(symbols) == 0 {
		return nil, grimoireErrors.NewError(grimoireErrors.SyntaxError, "No symbols to parse").
			WithDetails("The input contains no detected symbols")
	}

	// Find outer circle first
	var outerCircle *symbolNode
	var outerCircleIdx int = -1
	for i, sym := range p.symbols {
		if sym.Type == detector.OuterCircle {
			outerCircle = &symbolNode{symbol: sym}
			outerCircleIdx = i
			break
		}
	}
	if outerCircle == nil {
		return nil, grimoireErrors.NewError(grimoireErrors.SyntaxError, "No outer circle found").
			WithDetails("All Grimoire programs must be enclosed in a magic circle")
	}

	// Build adaptive spatial index
	p.buildAdaptiveSpatialIndex()

	// Build symbol graph with connections
	if len(p.connections) > 0 {
		p.applyConnections()
	} else {
		p.inferConnectionsOptimized()
	}

	// Make sure outer circle is in the graph
	if outerCircleIdx >= 0 {
		outerCircle = p.symbolGraph[outerCircleIdx]
	}

	// Build and process AST in parallel
	program, err := p.buildASTParallel(outerCircle)
	if err != nil {
		return nil, err
	}

	return program, nil
}

// buildAdaptiveSpatialIndex creates an adaptive spatial index
func (p *OptimizedParserV2) buildAdaptiveSpatialIndex() {
	symbolCount := len(p.symbols)

	// Calculate bounds
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, sym := range p.symbols {
		minX = math.Min(minX, sym.Position.X)
		minY = math.Min(minY, sym.Position.Y)
		maxX = math.Max(maxX, sym.Position.X)
		maxY = math.Max(maxY, sym.Position.Y)
	}

	// Initialize symbol graph
	p.symbolGraph = make(map[int]*symbolNode, symbolCount)

	// Build symbol cache for fast lookups
	for i, sym := range p.symbols {
		p.symbolCache[sym] = i
	}

	// Decide whether to use QuadTree based on symbol count and distribution
	useQuadTree := symbolCount > 200

	p.spatialIndex = &AdaptiveSpatialIndex{
		symbols:     p.symbols,
		nodes:       make([]*symbolNode, symbolCount),
		symbolCount: symbolCount,
		useQuadTree: useQuadTree,
	}

	if useQuadTree {
		// Use QuadTree for large symbol counts
		p.spatialIndex.quadTree = &QuadTree{
			bounds: Rectangle{
				x:      minX - 10,
				y:      minY - 10,
				width:  maxX - minX + 20,
				height: maxY - minY + 20,
			},
			maxDepth: 8,
			maxNodes: 4,
			depth:    0,
			nodes:    make([]*symbolNode, 0),
		}

		// Create nodes and add to QuadTree
		for i, symbol := range p.symbols {
			node := p.connectionPool.Get().(*symbolNode)
			node.symbol = symbol
			node.children = node.children[:0] // Reset slice
			node.parent = nil

			p.symbolGraph[i] = node
			p.spatialIndex.nodes[i] = node
			p.spatialIndex.quadTree.insert(node)
		}
	} else {
		// Use grid-based index for medium symbol counts
		gridDimension := math.Sqrt(float64(symbolCount))
		gridSize := math.Max((maxX-minX)/gridDimension, (maxY-minY)/gridDimension)
		if gridSize < 50 {
			gridSize = 50
		}

		p.spatialIndex.gridSize = gridSize
		p.spatialIndex.grid = make(map[gridKey][]*symbolNode, int(gridDimension*gridDimension))

		// Create nodes and add to grid
		for i, symbol := range p.symbols {
			node := p.connectionPool.Get().(*symbolNode)
			node.symbol = symbol
			node.children = node.children[:0]
			node.parent = nil

			p.symbolGraph[i] = node
			p.spatialIndex.nodes[i] = node

			key := gridKey{
				x: int(symbol.Position.X / gridSize),
				y: int(symbol.Position.Y / gridSize),
			}
			p.spatialIndex.grid[key] = append(p.spatialIndex.grid[key], node)
		}
	}
}

// QuadTree methods
func (qt *QuadTree) insert(node *symbolNode) {
	// If we have subdivisions, insert into appropriate quadrant
	if qt.northWest != nil {
		quadrant := qt.getQuadrant(node.symbol.Position)
		if quadrant != nil {
			quadrant.insert(node)
			return
		}
	}

	// Add to this node
	qt.nodes = append(qt.nodes, node)

	// Subdivide if necessary
	if len(qt.nodes) > qt.maxNodes && qt.depth < qt.maxDepth && qt.northWest == nil {
		qt.subdivide()

		// Redistribute nodes
		oldNodes := qt.nodes
		qt.nodes = make([]*symbolNode, 0)

		for _, n := range oldNodes {
			quadrant := qt.getQuadrant(n.symbol.Position)
			if quadrant != nil {
				quadrant.insert(n)
			} else {
				qt.nodes = append(qt.nodes, n)
			}
		}
	}
}

func (qt *QuadTree) subdivide() {
	halfWidth := qt.bounds.width / 2
	halfHeight := qt.bounds.height / 2
	x := qt.bounds.x
	y := qt.bounds.y

	qt.northWest = &QuadTree{
		bounds:   Rectangle{x, y, halfWidth, halfHeight},
		maxDepth: qt.maxDepth,
		maxNodes: qt.maxNodes,
		depth:    qt.depth + 1,
		nodes:    make([]*symbolNode, 0),
	}

	qt.northEast = &QuadTree{
		bounds:   Rectangle{x + halfWidth, y, halfWidth, halfHeight},
		maxDepth: qt.maxDepth,
		maxNodes: qt.maxNodes,
		depth:    qt.depth + 1,
		nodes:    make([]*symbolNode, 0),
	}

	qt.southWest = &QuadTree{
		bounds:   Rectangle{x, y + halfHeight, halfWidth, halfHeight},
		maxDepth: qt.maxDepth,
		maxNodes: qt.maxNodes,
		depth:    qt.depth + 1,
		nodes:    make([]*symbolNode, 0),
	}

	qt.southEast = &QuadTree{
		bounds:   Rectangle{x + halfWidth, y + halfHeight, halfWidth, halfHeight},
		maxDepth: qt.maxDepth,
		maxNodes: qt.maxNodes,
		depth:    qt.depth + 1,
		nodes:    make([]*symbolNode, 0),
	}
}

func (qt *QuadTree) getQuadrant(pos detector.Position) *QuadTree {
	if qt.northWest == nil {
		return nil
	}

	midX := qt.bounds.x + qt.bounds.width/2
	midY := qt.bounds.y + qt.bounds.height/2

	if pos.X < midX {
		if pos.Y < midY {
			return qt.northWest
		}
		return qt.southWest
	}

	if pos.Y < midY {
		return qt.northEast
	}
	return qt.southEast
}

func (qt *QuadTree) query(searchBounds Rectangle, results *[]*symbolNode) {
	// Check if search bounds intersect with this quadrant
	if !qt.intersects(searchBounds) {
		return
	}

	// Check nodes at this level
	for _, node := range qt.nodes {
		if searchBounds.contains(node.symbol.Position) {
			*results = append(*results, node)
		}
	}

	// Recursively check subdivisions
	if qt.northWest != nil {
		qt.northWest.query(searchBounds, results)
		qt.northEast.query(searchBounds, results)
		qt.southWest.query(searchBounds, results)
		qt.southEast.query(searchBounds, results)
	}
}

func (r Rectangle) contains(pos detector.Position) bool {
	return pos.X >= r.x && pos.X <= r.x+r.width &&
		pos.Y >= r.y && pos.Y <= r.y+r.height
}

func (qt *QuadTree) intersects(other Rectangle) bool {
	return !(other.x > qt.bounds.x+qt.bounds.width ||
		other.x+other.width < qt.bounds.x ||
		other.y > qt.bounds.y+qt.bounds.height ||
		other.y+other.height < qt.bounds.y)
}

// getNearbyNodes returns nodes within a certain distance using the adaptive index
func (asi *AdaptiveSpatialIndex) getNearbyNodes(pos detector.Position, maxDist float64) []*symbolNode {
	results := make([]*symbolNode, 0, 10)

	if asi.useQuadTree && asi.quadTree != nil {
		// Use QuadTree query
		searchBounds := Rectangle{
			x:      pos.X - maxDist,
			y:      pos.Y - maxDist,
			width:  maxDist * 2,
			height: maxDist * 2,
		}
		asi.quadTree.query(searchBounds, &results)

		// Filter by actual distance
		filtered := results[:0]
		maxDistSq := maxDist * maxDist
		for _, node := range results {
			dx := node.symbol.Position.X - pos.X
			dy := node.symbol.Position.Y - pos.Y
			if dx*dx+dy*dy <= maxDistSq {
				filtered = append(filtered, node)
			}
		}
		return filtered
	} else {
		// Use grid-based search
		cellRadius := int(math.Ceil(maxDist / asi.gridSize))
		centerKey := gridKey{
			x: int(pos.X / asi.gridSize),
			y: int(pos.Y / asi.gridSize),
		}

		maxDistSq := maxDist * maxDist

		for dx := -cellRadius; dx <= cellRadius; dx++ {
			for dy := -cellRadius; dy <= cellRadius; dy++ {
				key := gridKey{x: centerKey.x + dx, y: centerKey.y + dy}
				if nodes, ok := asi.grid[key]; ok {
					for _, node := range nodes {
						distX := node.symbol.Position.X - pos.X
						distY := node.symbol.Position.Y - pos.Y
						if distX*distX+distY*distY <= maxDistSq {
							results = append(results, node)
						}
					}
				}
			}
		}
	}

	return results
}

// buildASTParallel builds the AST using parallel processing where safe
func (p *OptimizedParserV2) buildASTParallel(outerCircle *symbolNode) (*Program, error) {
	// Find all top-level nodes (direct children of outer circle)
	topLevelNodes := make([]*symbolNode, 0)
	for _, node := range p.symbolGraph {
		if node.parent == outerCircle {
			topLevelNodes = append(topLevelNodes, node)
		}
	}

	if len(topLevelNodes) == 0 {
		return nil, grimoireErrors.NewError(grimoireErrors.SyntaxError, "Empty program").
			WithDetails("No symbols found inside the outer circle")
	}

	// Process top-level nodes in parallel
	statements := make([]Statement, len(topLevelNodes))

	var wg sync.WaitGroup
	for i, node := range topLevelNodes {
		wg.Add(1)
		go func(idx int, n *symbolNode) {
			defer wg.Done()
			stmt := p.Parser.parseStatement(n)
			statements[idx] = stmt
		}(i, node)
	}
	wg.Wait()

	// Filter out nil statements
	validStatements := make([]Statement, 0, len(statements))
	for _, stmt := range statements {
		if stmt != nil {
			validStatements = append(validStatements, stmt)
		}
	}

	// Build the program using the parent parser's logic
	// Store statements temporarily
	p.Parser.symbolGraph = p.symbolGraph

	// Create a simple program structure
	return &Program{
		HasOuterCircle: true,
		Globals:        validStatements,
	}, nil
}

// applyConnections applies explicit connections
func (p *OptimizedParserV2) applyConnections() {
	// Build symbol to index map
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

// inferConnectionsOptimized infers connections using spatial index
func (p *OptimizedParserV2) inferConnectionsOptimized() {
	for idx, node := range p.symbolGraph {
		if node == nil || node.symbol == nil {
			continue
		}

		// Skip if already has outgoing connections
		if len(node.children) > 0 {
			continue
		}

		// Use adaptive spatial index to find nearby nodes
		nearbyNodes := p.spatialIndex.getNearbyNodes(node.symbol.Position, 150)

		// Find the best connection based on pattern
		var bestConnection *symbolNode
		minDistance := math.MaxFloat64

		for _, candidate := range nearbyNodes {
			if candidate == node || candidate.symbol == nil {
				continue
			}

			// Calculate distance
			dx := candidate.symbol.Position.X - node.symbol.Position.X
			dy := candidate.symbol.Position.Y - node.symbol.Position.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			// Skip if too close
			if distance < 30 {
				continue
			}

			// Update if this is the closest valid connection
			if distance < minDistance {
				minDistance = distance
				bestConnection = candidate
			}
		}

		// Connect to the best candidate
		if bestConnection != nil {
			p.symbolGraph[idx].children = append(p.symbolGraph[idx].children, bestConnection)
		}
	}
}

// Cleanup returns nodes to the pool
func (p *OptimizedParserV2) Cleanup() {
	for _, node := range p.symbolGraph {
		p.connectionPool.Put(node)
	}
}
