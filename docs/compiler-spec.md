# Grimoire Compiler Architecture Specification

## Table of Contents
1. [Overview](#overview)
2. [Compilation Pipeline](#compilation-pipeline)
3. [Image Recognition Stage](#image-recognition)
4. [Symbol Extraction](#symbol-extraction)
5. [Topology Analysis](#topology-analysis)
6. [AST Generation](#ast-generation)
7. [Type Inference](#type-inference)
8. [Code Generation](#code-generation)
9. [Optimization](#optimization)
10. [Error Handling](#error-handling)
11. [Debug Support](#debug-support)

## Overview {#overview}

The Grimoire compiler transforms hand-drawn magic circles into executable programs through a multi-stage pipeline combining computer vision, symbolic analysis, and traditional compilation techniques.

### Compiler Components
```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│Image Input  │ --> │CV Processing │ --> │Symbol Extract│
└─────────────┘     └──────────────┘     └──────────────┘
                           |
                           v
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│Code Gen     │ <-- │Type Inference│ <-- │AST Builder   │
└─────────────┘     └──────────────┘     └──────────────┘
```

## Compilation Pipeline {#compilation-pipeline}

### Stage 1: Image Preprocessing
1. Load image (PNG/JPG/SVG)
2. Noise reduction
3. Contrast enhancement
4. Binary thresholding
5. Edge detection

### Stage 2: Shape Recognition
1. Contour detection
2. Shape classification using ML
3. Symbol boundary extraction
4. Text recognition (OCR)

### Stage 3: Semantic Analysis
1. Topology graph construction
2. Symbol relationship mapping
3. Flow direction detection
4. Scope hierarchy building

### Stage 4: Code Generation
1. AST construction
2. Type inference
3. Optimization passes
4. Target code emission

## Image Recognition Stage {#image-recognition}

### Preprocessing Pipeline
```python
def preprocess_image(image):
    # 1. Convert to grayscale
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    
    # 2. Apply Gaussian blur for noise reduction
    blurred = cv2.GaussianBlur(gray, (5, 5), 0)
    
    # 3. Adaptive thresholding for varying lighting
    binary = cv2.adaptiveThreshold(blurred, 255, 
                                  cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                                  cv2.THRESH_BINARY, 11, 2)
    
    # 4. Morphological operations
    kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (3, 3))
    cleaned = cv2.morphologyEx(binary, cv2.MORPH_CLOSE, kernel)
    
    return cleaned
```

### Shape Detection Algorithm
```python
def detect_shapes(contours):
    shapes = []
    for contour in contours:
        # Approximate polygon
        epsilon = 0.02 * cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, epsilon, True)
        
        # Classify by vertex count
        vertices = len(approx)
        if vertices == 3:
            shape_type = "triangle"
        elif vertices == 4:
            shape_type = "square"
        elif vertices == 5:
            shape_type = "pentagon"
        elif vertices == 6:
            shape_type = "hexagon"
        elif vertices > 6:
            # Check if circle
            area = cv2.contourArea(contour)
            perimeter = cv2.arcLength(contour, True)
            circularity = 4 * np.pi * area / (perimeter ** 2)
            if circularity > 0.8:
                shape_type = "circle"
            else:
                shape_type = "star" if self.is_star(approx) else "polygon"
                
        shapes.append(Shape(shape_type, contour, approx))
    return shapes
```

## Symbol Extraction {#symbol-extraction}

### Symbol Recognition Pipeline
1. **Isolation**: Extract symbol regions from shapes
2. **Normalization**: Scale and rotate to standard orientation
3. **Feature Extraction**: 
   - Geometric features (angles, curves)
   - Topological features (holes, intersections)
   - Statistical features (moments, distributions)
4. **Classification**: Neural network or template matching

### Custom Symbol Training
```python
class SymbolRecognizer:
    def __init__(self):
        self.model = self.load_pretrained_model()
        self.custom_symbols = {}
    
    def add_custom_symbol(self, image, meaning):
        features = self.extract_features(image)
        self.custom_symbols[meaning] = features
    
    def recognize(self, symbol_image):
        # Try standard symbols first
        prediction = self.model.predict(symbol_image)
        if prediction.confidence > 0.8:
            return prediction.symbol
            
        # Fall back to custom symbols
        return self.match_custom_symbol(symbol_image)
```

## Topology Analysis {#topology-analysis}

### Graph Construction
```python
class TopologyGraph:
    def __init__(self):
        self.nodes = {}  # shape_id -> Shape
        self.edges = {}  # (shape1_id, shape2_id) -> Connection
    
    def build_from_shapes(self, shapes):
        # 1. Create nodes
        for shape in shapes:
            self.nodes[shape.id] = shape
        
        # 2. Detect connections
        for shape1 in shapes:
            for shape2 in shapes:
                if shape1.id != shape2.id:
                    connection = self.detect_connection(shape1, shape2)
                    if connection:
                        self.edges[(shape1.id, shape2.id)] = connection
```

### Connection Detection
- **Direct touch**: Shapes share boundaries
- **Line connection**: Explicit lines between shapes
- **Containment**: One shape inside another
- **Proximity**: Close shapes with implicit connection

### Flow Analysis
```python
def analyze_flow(topology):
    # Detect flow direction from:
    # 1. Arrow directions
    # 2. Clockwise/counterclockwise patterns
    # 3. Numeric annotations
    # 4. Default top-to-bottom, left-to-right
    
    flow_graph = FlowGraph()
    for edge in topology.edges:
        direction = infer_direction(edge)
        flow_graph.add_directed_edge(edge.source, edge.target, direction)
    
    return flow_graph
```

## AST Generation {#ast-generation}

### AST Node Types
```python
class ASTNode:
    pass

class ProgramNode(ASTNode):
    def __init__(self, main_circle, functions, globals):
        self.main = main_circle
        self.functions = functions
        self.globals = globals

class CircleNode(ASTNode):  # Function/Scope
    def __init__(self, name, params, body):
        self.name = name
        self.params = params
        self.body = body

class SquareNode(ASTNode):  # Variable
    def __init__(self, name, type, value):
        self.name = name
        self.type = type
        self.value = value

class TriangleNode(ASTNode):  # Conditional
    def __init__(self, condition, true_branch, false_branch):
        self.condition = condition
        self.true_branch = true_branch
        self.false_branch = false_branch
```

### AST Construction Algorithm
```python
def build_ast(topology, symbols):
    # 1. Find main entry point
    main = find_main_circle(topology)
    
    # 2. Build function definitions
    functions = []
    for circle in topology.get_circles():
        if circle != main:
            func_ast = build_function_ast(circle, topology)
            functions.append(func_ast)
    
    # 3. Build main execution flow
    main_ast = build_execution_ast(main, topology)
    
    return ProgramNode(main_ast, functions, globals)
```

## Type Inference {#type-inference}

### Type Inference Rules
1. **Shape-based inference**: Square outline style indicates type
2. **Operator-based inference**: Connected operators constrain types
3. **Flow-based inference**: Types propagate through connections
4. **Annotation-based**: Explicit type symbols override inference

### Type Inference Algorithm
```python
class TypeInferencer:
    def infer_types(self, ast):
        # Build constraint graph
        constraints = self.collect_constraints(ast)
        
        # Solve constraints using unification
        substitutions = self.unify(constraints)
        
        # Apply substitutions to AST
        typed_ast = self.apply_types(ast, substitutions)
        
        # Check for conflicts
        self.verify_types(typed_ast)
        
        return typed_ast
```

## Code Generation {#code-generation}

### Backend Options

#### 1. C Backend
```python
class CCodeGenerator:
    def generate(self, ast):
        self.emit("#include <stdio.h>")
        self.emit("#include <stdlib.h>")
        
        # Generate function declarations
        for func in ast.functions:
            self.generate_function_decl(func)
        
        # Generate main
        self.emit("int main() {")
        self.generate_statements(ast.main.body)
        self.emit("return 0;")
        self.emit("}")
```

#### 2. LLVM Backend
```python
class LLVMCodeGenerator:
    def __init__(self):
        self.module = llvm.Module()
        self.builder = llvm.Builder()
    
    def generate(self, ast):
        # Generate LLVM IR
        for func in ast.functions:
            self.generate_function(func)
        
        # Generate main
        main_func = self.module.add_function("main", ...)
        self.generate_main(ast.main)
```

#### 3. Bytecode Backend
```python
class BytecodeGenerator:
    def generate(self, ast):
        bytecode = []
        
        # Generate instructions
        for node in ast.walk():
            instructions = self.generate_instructions(node)
            bytecode.extend(instructions)
        
        return GrimoireBytecode(bytecode)
```

## Optimization {#optimization}

### Visual Optimization Hints
- **Line thickness**: Indicates hot paths
- **Color intensity**: Suggests optimization level
- **Overlapping shapes**: Enable inlining

### Optimization Passes
1. **Dead shape elimination**: Remove unreachable shapes
2. **Shape fusion**: Combine adjacent operations
3. **Loop unrolling**: Based on pentagon annotations
4. **Parallel detection**: Hexagon patterns to thread pools

## Error Handling {#error-handling}

### Compilation Errors
```python
class GrimoireError:
    def __init__(self, shape, message, suggestion=None):
        self.shape = shape
        self.location = shape.bounding_box
        self.message = message
        self.suggestion = suggestion
    
    def visualize(self, image):
        # Draw error highlights on original image
        cv2.rectangle(image, self.location, (0, 0, 255), 3)
        # Add error message
        cv2.putText(image, self.message, ...)
```

### Error Types
1. **Shape Recognition Errors**
   - Ambiguous shapes
   - Incomplete circles
   - Overlapping conflicts

2. **Semantic Errors**
   - Disconnected components
   - Type mismatches
   - Undefined symbols

3. **Logic Errors**
   - Infinite loops without exit
   - Unreachable code
   - Circular dependencies

## Debug Support {#debug-support}

### Debug Information Generation
```python
class DebugInfo:
    def __init__(self):
        self.shape_to_line = {}  # Map shapes to generated code lines
        self.breakpoints = []    # Shapes marked as breakpoints
        self.watches = []        # Shapes marked for watching
    
    def generate_sourcemap(self, shapes, generated_code):
        # Create bidirectional mapping between visual and textual code
        pass
```

### Visual Debugger Integration
1. **Execution visualization**: Highlight currently executing shape
2. **Variable inspection**: Show values near shapes
3. **Step-through debugging**: Shape-by-shape execution
4. **Time-travel debugging**: Replay execution visually

### Debug Compilation Mode
```bash
grimoire compile --debug circle.png
# Generates:
# - circle.grim.debug (debug symbols)
# - circle.grim.map (shape-to-code mapping)
# - circle.grim.trace (execution trace format)
```

---

*This specification defines the technical architecture of the Grimoire compiler.*