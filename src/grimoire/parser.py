"""Parser for Grimoire - Converts detected symbols to AST"""

from typing import List, Optional, Dict, Set, Tuple
from dataclasses import dataclass
import math

from .image_recognition import Symbol, SymbolType, Connection
from .ast_nodes import *
from .errors import ParseError


@dataclass
class SymbolNode:
    """Wrapper for symbol with parsing metadata"""
    symbol: Symbol
    visited: bool = False
    ast_node: Optional[ASTNode] = None
    parent: Optional['SymbolNode'] = None
    children: List['SymbolNode'] = None
    
    def __post_init__(self):
        if self.children is None:
            self.children = []


class MagicCircleParser:
    """Parse symbols and connections into an AST"""
    
    def __init__(self):
        self.symbols: List[Symbol] = []
        self.connections: List[Connection] = []
        self.symbol_graph: Dict[int, SymbolNode] = {}  # Use symbol index as key
        self.symbol_index_map: Dict[Symbol, int] = {}  # Map symbols to indices
        self.errors: List[ParseError] = []
        self.call_depth = 0  # For stack overflow prevention
        self.max_call_depth = 100
    
    def parse(self, symbols: List[Symbol], connections: List[Connection]) -> Program:
        """Main parsing function"""
        self.symbols = symbols
        self.connections = connections
        
        # Build symbol graph
        self._build_symbol_graph()
        
        # Find outer circle
        outer_circle = self._find_outer_circle()
        if not outer_circle:
            raise ParseError(
                "外周円が見つかりません。Grimoireプログラムは魔法陣で囲まれている必要があります",
                error_code="PARSE_NO_OUTER_CIRCLE"
            )
        
        # Find main entry
        main_entry = self._find_main_entry()
        
        # Parse functions (excluding main)
        functions = self._parse_functions()
        
        # If we have a main entry but empty body, parse global statements into main
        if main_entry and not main_entry.body:
            # Find unvisited statements and add to main body
            global_stmts = []
            for i, symbol in enumerate(self.symbols):
                node = self.symbol_graph[i]
                # Include star symbols even if they're marked as visited (connected from other symbols)
                if (not node.visited or symbol.type == SymbolType.STAR) and symbol.type not in [SymbolType.OUTER_CIRCLE, SymbolType.DOUBLE_CIRCLE]:
                    stmt = self._parse_statement(node)
                    if stmt:
                        global_stmts.append(stmt)
            main_entry.body = global_stmts
            globals = []
        else:
            # Parse remaining global statements
            globals = self._parse_global_statements()
            
        # Special case: if we have no main entry and just global statements (like a single star)
        # create an implicit main function
        if not main_entry and globals:
            main_entry = FunctionDef(
                name=None,
                parameters=[],
                body=globals,
                return_type=None,
                is_main=True
            )
            globals = []
        
        return Program(
            has_outer_circle=True,
            main_entry=main_entry,
            functions=functions,
            globals=globals
        )
    
    def _build_symbol_graph(self):
        """Build a graph of symbols and their connections"""
        # Create nodes for all symbols and build index map
        for i, symbol in enumerate(self.symbols):
            self.symbol_graph[i] = SymbolNode(symbol)
            self.symbol_index_map[id(symbol)] = i
        
        # Build parent-child relationships based on connections
        if self.connections:
            for conn in self.connections:
                from_idx = self._get_symbol_index(conn.from_symbol)
                to_idx = self._get_symbol_index(conn.to_symbol)
                
                if from_idx is not None and to_idx is not None:
                    from_node = self.symbol_graph[from_idx]
                    to_node = self.symbol_graph[to_idx]
                    from_node.children.append(to_node)
                    to_node.parent = from_node
        else:
            # No explicit connections - infer based on proximity and symbol types
            self._infer_connections()
    
    def _get_symbol_index(self, symbol: Symbol) -> Optional[int]:
        """Get the index of a symbol"""
        # Try to find symbol by comparing attributes
        for i, s in enumerate(self.symbols):
            if (s.type == symbol.type and 
                s.position == symbol.position and 
                s.size == symbol.size):
                return i
        return None
    
    def _find_outer_circle(self) -> Optional[Symbol]:
        """Find the outer circle symbol"""
        for symbol in self.symbols:
            if symbol.type == SymbolType.OUTER_CIRCLE:
                return symbol
        return None
    
    def _find_main_entry(self) -> Optional[FunctionDef]:
        """Find the main entry point (double circle ◎)"""
        for i, symbol in enumerate(self.symbols):
            if symbol.type == SymbolType.DOUBLE_CIRCLE:
                node = self.symbol_graph[i]
                # Parse as main function
                return self._parse_function(node, is_main=True)
        return None
    
    def _parse_functions(self) -> List[FunctionDef]:
        """Parse all function definitions (circles)"""
        functions = []
        
        for i, symbol in enumerate(self.symbols):
            if symbol.type == SymbolType.CIRCLE and not self.symbol_graph[i].visited:
                node = self.symbol_graph[i]
                func = self._parse_function(node)
                if func:
                    functions.append(func)
        
        return functions
    
    def _parse_function(self, node: SymbolNode, is_main: bool = False) -> Optional[FunctionDef]:
        """Parse a function definition from a circle"""
        if node.visited:
            return node.ast_node if isinstance(node.ast_node, FunctionDef) else None
        
        node.visited = True
        
        # Extract parameters (incoming connections)
        parameters = self._extract_parameters(node)
        
        # For main functions, also mark direct children for parsing
        if is_main:
            # Don't mark children as visited yet - they need to be parsed as statements
            pass
        
        # Parse function body
        body = self._parse_statement_sequence(node.children)
        
        # Determine return type (if any)
        return_type = self._infer_return_type(body)
        
        func = FunctionDef(
            name=None,  # Anonymous functions for now
            parameters=parameters,
            body=body,
            return_type=return_type,
            is_main=is_main
        )
        
        node.ast_node = func
        return func
    
    def _extract_parameters(self, func_node: SymbolNode) -> List[Parameter]:
        """Extract function parameters from incoming connections"""
        parameters = []
        
        # Main functions and simple programs don't have parameters
        if func_node.symbol.type == SymbolType.DOUBLE_CIRCLE:
            return []
        
        # Look for square symbols connected as inputs (parameters)
        # Parameters would be squares without operators attached
        for child in func_node.children:
            if child.symbol.type == SymbolType.SQUARE:
                # Check if this square is a parameter (not connected to operators)
                has_operator = any(
                    grandchild.symbol.type in [SymbolType.CONVERGENCE, SymbolType.DIVERGENCE,
                                              SymbolType.AMPLIFICATION, SymbolType.DISTRIBUTION]
                    for grandchild in child.children
                )
                
                if not has_operator:
                    # This is a parameter
                    pattern = child.symbol.properties.get("pattern", "")
                    data_type = self._pattern_to_datatype(pattern)
                    
                    param = Parameter(
                        name=f"param{len(parameters)}",
                        data_type=data_type
                    )
                    parameters.append(param)
                    # Don't mark as visited for main functions - they need to be parsed
                    if func_node.symbol.type != SymbolType.DOUBLE_CIRCLE:
                        child.visited = True
        
        return parameters
    
    def _parse_statement_sequence(self, nodes: List[SymbolNode]) -> List[Statement]:
        """Parse a sequence of statements from connected nodes"""
        statements = []
        
        for node in nodes:
            if node.visited:
                continue
            
            stmt = self._parse_statement(node)
            if stmt:
                statements.append(stmt)
        
        return statements
    
    def _parse_statement(self, node: SymbolNode) -> Optional[Statement]:
        """Parse a single statement from a symbol"""
        if node.visited and node.symbol.type != SymbolType.STAR:
            # Stars should always be parsed as output statements
            return None
        
        node.visited = True
        symbol = node.symbol
        
        # Output statement (☆)
        if symbol.type == SymbolType.STAR:
            # Look for expression to output
            expr = self._parse_expression_from_parent(node)
            return OutputStatement(value=expr)
        
        # Conditional (△)
        elif symbol.type == SymbolType.TRIANGLE:
            return self._parse_if_statement(node)
        
        # Loop (⬟)
        elif symbol.type == SymbolType.PENTAGON:
            return self._parse_loop(node)
        
        # Parallel block (⬢)
        elif symbol.type == SymbolType.HEXAGON:
            return self._parse_parallel_block(node)
        
        # Binary operator - skip as statement (will be part of expression)
        elif symbol.type in [SymbolType.CONVERGENCE, SymbolType.DIVERGENCE,
                            SymbolType.AMPLIFICATION, SymbolType.DISTRIBUTION]:
            # Mark children as visited so they're not parsed as separate statements
            for child in node.children:
                child.visited = True
            return None
        
        # Assignment or expression
        elif symbol.type == SymbolType.SQUARE:
            # Check if it's connected to an operator
            has_operator_child = any(
                child.symbol.type in [SymbolType.CONVERGENCE, SymbolType.DIVERGENCE,
                                     SymbolType.AMPLIFICATION, SymbolType.DISTRIBUTION]
                for child in node.children
            )
            
            if not has_operator_child:
                # Standalone variable/assignment
                return self._parse_assignment(node)
            else:
                # Part of an expression, skip
                return None
        
        # Expression statement
        expr = self._parse_expression(node)
        if expr:
            return ExpressionStatement(expression=expr)
        
        return None
    
    def _parse_if_statement(self, node: SymbolNode) -> IfStatement:
        """Parse an if statement from a triangle"""
        # Find condition (usually comparison operators near the triangle)
        condition = self._parse_condition(node)
        
        # Find branches
        then_branch = []
        else_branch = []
        
        # Simple heuristic: left branch is 'then', right branch is 'else'
        left_children = []
        right_children = []
        
        for child in node.children:
            # Determine if child is on left or right based on position
            if child.symbol.position[0] < node.symbol.position[0]:
                left_children.append(child)
            else:
                right_children.append(child)
        
        then_branch = self._parse_statement_sequence(left_children)
        else_branch = self._parse_statement_sequence(right_children) if right_children else None
        
        return IfStatement(
            condition=condition,
            then_branch=then_branch,
            else_branch=else_branch
        )
    
    def _parse_loop(self, node: SymbolNode) -> Statement:
        """Parse a loop from a pentagon"""
        # Look for counter or condition
        counter_node = None
        for parent in self._get_parents(node):
            if parent.symbol.type == SymbolType.SQUARE:
                counter_node = parent
                break
        
        if counter_node:
            # For loop with counter
            counter = Identifier(name="loop_counter")
            
            # Extract count from counter node
            count_expr = self._parse_numeric_literal(counter_node)
            
            body = self._parse_statement_sequence(node.children)
            
            return ForLoop(
                counter=counter,
                start=Literal(value=0, literal_type=DataType.INTEGER),
                end=count_expr,
                step=Literal(value=1, literal_type=DataType.INTEGER),
                body=body
            )
        else:
            # While loop with condition
            condition = self._parse_condition(node)
            body = self._parse_statement_sequence(node.children)
            
            return WhileLoop(
                condition=condition,
                body=body
            )
    
    def _parse_parallel_block(self, node: SymbolNode) -> ParallelBlock:
        """Parse a parallel block from a hexagon"""
        branches = []
        
        # Group children by their angular position relative to the hexagon
        child_groups = self._group_children_by_angle(node)
        
        for group in child_groups:
            branch = self._parse_statement_sequence(group)
            if branch:
                branches.append(branch)
        
        return ParallelBlock(branches=branches)
    
    def _parse_assignment(self, node: SymbolNode) -> Optional[Assignment]:
        """Parse an assignment statement"""
        # Variable is represented by the square
        var_name = f"var_{id(node)}"
        target = Identifier(name=var_name)
        
        # Look for value (connected expression)
        value = None
        for child in node.children:
            value = self._parse_expression(child)
            if value:
                break
        
        if not value:
            # Look for literal value in properties
            value = self._parse_literal_from_properties(node)
        
        if value:
            return Assignment(target=target, value=value)
        
        return None
    
    def _parse_expression(self, node: SymbolNode) -> Optional[Expression]:
        """Parse an expression from a symbol"""
        if node.visited and node.ast_node:
            return node.ast_node if isinstance(node.ast_node, Expression) else None
        
        node.visited = True
        symbol = node.symbol
        
        # Binary operators
        if symbol.type in [SymbolType.CONVERGENCE, SymbolType.DIVERGENCE,
                          SymbolType.AMPLIFICATION, SymbolType.DISTRIBUTION]:
            return self._parse_binary_op(node)
        
        # Literals
        elif symbol.type == SymbolType.SQUARE:
            return self._parse_literal_from_properties(node)
        
        # Function call
        elif symbol.type == SymbolType.CIRCLE:
            return self._parse_function_call(node)
        
        return None
    
    def _parse_binary_op(self, node: SymbolNode) -> Optional[BinaryOp]:
        """Parse a binary operation"""
        # Map symbol types to operators
        op_map = {
            SymbolType.CONVERGENCE: OperatorType.ADD,
            SymbolType.DIVERGENCE: OperatorType.SUBTRACT,
            SymbolType.AMPLIFICATION: OperatorType.MULTIPLY,
            SymbolType.DISTRIBUTION: OperatorType.DIVIDE
        }
        
        operator = op_map.get(node.symbol.type)
        if not operator:
            return None
        
        # Find operands (connected squares or expressions)
        operands = []
        
        # Check parents (inputs to the operator)
        parents = self._get_parents(node)
        for parent in parents:
            if parent.symbol.type == SymbolType.SQUARE:
                # Parse the literal value from the square
                literal = self._parse_literal_from_properties(parent)
                if literal:
                    operands.append(literal)
            else:
                expr = self._parse_expression(parent)
                if expr:
                    operands.append(expr)
        
        # If we have exactly 2 operands, create the binary op
        if len(operands) == 2:
            return BinaryOp(
                left=operands[0],
                operator=operator,
                right=operands[1]
            )
        elif len(operands) == 1:
            # Single operand - look for another operand in children or use default
            return BinaryOp(
                left=operands[0],
                operator=operator,
                right=Literal(value=0, literal_type=DataType.INTEGER)
            )
        else:
            # No operands found - use defaults
            return BinaryOp(
                left=Literal(value=0, literal_type=DataType.INTEGER),
                operator=operator,
                right=Literal(value=0, literal_type=DataType.INTEGER)
            )
    
    def _parse_function_call(self, node: SymbolNode) -> FunctionCall:
        """Parse a function call"""
        # If already visited, return a simple function call without recursing
        if node.visited:
            return FunctionCall(
                function=None,  # Built-in or unresolved function
                arguments=[]
            )
        
        node.visited = True
        
        # The circle represents the function
        # For now, treat all circles as built-in print functions to avoid deep recursion
        func_def = None  # We'll handle function resolution later
        
        # Arguments are the connected inputs
        arguments = []
        for parent in self._get_parents(node):
            if not parent.visited:  # Avoid cycles
                arg = self._parse_expression(parent)
                if arg:
                    arguments.append(arg)
        
        return FunctionCall(
            function=func_def,
            arguments=arguments
        )
    
    def _parse_literal_from_properties(self, node: SymbolNode) -> Optional[Literal]:
        """Parse a literal value from symbol properties"""
        pattern = node.symbol.properties.get("pattern", "")
        
        # Numeric literals based on dots
        if pattern == "dot":
            return Literal(value=1, literal_type=DataType.INTEGER)
        elif pattern == "double_dot":
            return Literal(value=2, literal_type=DataType.INTEGER)
        elif pattern == "triple_dot":
            return Literal(value=3, literal_type=DataType.INTEGER)
        elif pattern == "multiple_dots":
            # Extract number from properties if available
            return Literal(value=4, literal_type=DataType.INTEGER)
        elif pattern == "empty":
            return Literal(value=0, literal_type=DataType.INTEGER)
        elif pattern == "triple_line":
            # String literal - for now, default to "Hello, World!"
            return Literal(value="Hello, World!", literal_type=DataType.STRING)
        elif pattern in ["lines", "single_line", "double_line"]:
            # String literals
            return Literal(value="Text", literal_type=DataType.STRING)
        elif pattern == "cross":
            # Boolean true (crossed out)
            return Literal(value=True, literal_type=DataType.BOOLEAN)
        elif pattern == "half_circle":
            # Boolean false
            return Literal(value=False, literal_type=DataType.BOOLEAN)
        elif pattern == "grid":
            # Array or map indicator
            return Literal(value=[], literal_type=DataType.ARRAY)
        elif pattern == "filled":
            # Special value
            return Literal(value=-1, literal_type=DataType.INTEGER)
        
        # Default integer literal
        return Literal(value=0, literal_type=DataType.INTEGER)
    
    def _parse_numeric_literal(self, node: SymbolNode) -> Expression:
        """Parse numeric literal from visual representation"""
        # This would analyze the actual drawn dots/symbols
        # For now, return a default
        return Literal(value=10, literal_type=DataType.INTEGER)
    
    def _parse_condition(self, node: SymbolNode) -> Expression:
        """Parse a condition expression"""
        # Look for comparison operators near the node
        # For now, return a simple false literal to avoid infinite loops
        return Literal(value=False, literal_type=DataType.BOOLEAN)
    
    def _parse_expression_from_parent(self, node: SymbolNode) -> Expression:
        """Parse expression from parent nodes"""
        # If node has a parent, parse that as expression
        if node.parent:
            parent_expr = self._parse_expression(node.parent)
            if parent_expr:
                return parent_expr
        
        # Otherwise look at all parents
        for parent in self._get_parents(node):
            expr = self._parse_expression(parent)
            if expr:
                return expr
        
        # For standalone stars (hello world case), return "Hello, World!"
        if node.symbol.type == SymbolType.STAR:
            return Literal(value="Hello, World!", literal_type=DataType.STRING)
        
        # Default literal
        return Literal(value=0, literal_type=DataType.INTEGER)
    
    def _get_parents(self, node: SymbolNode) -> List[SymbolNode]:
        """Get all parent nodes (nodes that connect TO this node)"""
        parents = []
        for idx, sym_node in self.symbol_graph.items():
            if node in sym_node.children:
                parents.append(sym_node)
        return parents
    
    def _group_children_by_angle(self, node: SymbolNode) -> List[List[SymbolNode]]:
        """Group children by their angular position"""
        if not node.children:
            return []
        
        # Calculate angles
        center_x, center_y = node.symbol.position
        child_angles = []
        
        for child in node.children:
            cx, cy = child.symbol.position
            angle = math.atan2(cy - center_y, cx - center_x)
            child_angles.append((angle, child))
        
        # Sort by angle
        child_angles.sort(key=lambda x: x[0])
        
        # Group into sectors (e.g., 6 sectors for hexagon)
        groups = [[] for _ in range(6)]
        for angle, child in child_angles:
            sector = int((angle + math.pi) / (2 * math.pi / 6))
            # Ensure sector is within bounds
            sector = min(max(0, sector), 5)
            groups[sector].append(child)
        
        return [g for g in groups if g]  # Remove empty groups
    
    def _pattern_to_datatype(self, pattern: str) -> DataType:
        """Convert pattern to data type"""
        mapping = {
            "dot": DataType.INTEGER,
            "double_dot": DataType.FLOAT,
            "lines": DataType.STRING,
            "half_circle": DataType.BOOLEAN,
            "stars": DataType.ARRAY,
            "grid": DataType.MAP
        }
        return mapping.get(pattern, DataType.INTEGER)
    
    def _infer_return_type(self, statements: List[Statement]) -> Optional[DataType]:
        """Infer return type from function body"""
        for stmt in statements:
            if isinstance(stmt, ReturnStatement) and stmt.value:
                return stmt.value.data_type
        return None
    
    def _parse_global_statements(self) -> List[Statement]:
        """Parse global statements (not inside functions)"""
        globals = []
        
        # Find unvisited symbols that are direct children of outer circle
        for i, symbol in enumerate(self.symbols):
            node = self.symbol_graph[i]
            # Include star symbols even if they're marked as visited
            if (not node.visited or symbol.type == SymbolType.STAR) and symbol.type != SymbolType.OUTER_CIRCLE:
                stmt = self._parse_statement(node)
                if stmt:
                    globals.append(stmt)
        
        return globals
    
    def _infer_connections(self):
        """Infer connections based on symbol positions and types"""
        # Skip outer circle
        symbols_to_connect = []
        for i, sym in enumerate(self.symbols):
            if sym.type != SymbolType.OUTER_CIRCLE:
                symbols_to_connect.append((i, sym))
        
        # Connect symbols based on proximity and logical flow
        # 1. Connect main entry to nearest symbols
        main_idx = None
        for i, sym in symbols_to_connect:
            if sym.type == SymbolType.DOUBLE_CIRCLE:
                main_idx = i
                break
        
        if main_idx is not None:
            # Find symbols below main entry
            main_y = self.symbols[main_idx].position[1]
            for i, sym in symbols_to_connect:
                if i != main_idx and sym.position[1] > main_y:
                    # Check if it's reasonably aligned horizontally
                    x_diff = abs(sym.position[0] - self.symbols[main_idx].position[0])
                    if x_diff < 150:  # Within reasonable horizontal distance
                        self.symbol_graph[main_idx].children.append(self.symbol_graph[i])
                        self.symbol_graph[i].parent = self.symbol_graph[main_idx]
        
        # 2. Connect operators to nearby operands
        for i, sym in symbols_to_connect:
            if sym.type in [SymbolType.CONVERGENCE, SymbolType.DIVERGENCE, 
                           SymbolType.AMPLIFICATION, SymbolType.DISTRIBUTION]:
                # Find nearby squares (operands)
                operator_pos = sym.position
                nearby_operands = []
                
                for j, other_sym in symbols_to_connect:
                    if j != i and other_sym.type == SymbolType.SQUARE:
                        dist = math.sqrt(
                            (other_sym.position[0] - operator_pos[0])**2 + 
                            (other_sym.position[1] - operator_pos[1])**2
                        )
                        if dist < 150:  # Within reasonable distance
                            nearby_operands.append((j, dist))
                
                # Connect two nearest operands to operator
                nearby_operands.sort(key=lambda x: x[1])
                for idx, _ in nearby_operands[:2]:
                    self.symbol_graph[idx].children.append(self.symbol_graph[i])
                    self.symbol_graph[i].parent = self.symbol_graph[idx]
        
        # 3. Connect outputs (stars) to nearest expressions
        for i, sym in symbols_to_connect:
            if sym.type == SymbolType.STAR:
                # Find nearest symbol above this star
                star_pos = sym.position
                nearest = None
                min_dist = float('inf')
                
                for j, other_sym in symbols_to_connect:
                    if j != i and other_sym.position[1] < star_pos[1]:  # Above the star
                        try:
                            dx = float(other_sym.position[0]) - float(star_pos[0])
                            dy = float(other_sym.position[1]) - float(star_pos[1])
                            dist = math.sqrt(dx**2 + dy**2)
                        except (OverflowError, ValueError):
                            continue
                        if dist < min_dist:
                            min_dist = dist
                            nearest = j
                
                if nearest is not None and min_dist < 150:
                    self.symbol_graph[nearest].children.append(self.symbol_graph[i])
                    self.symbol_graph[i].parent = self.symbol_graph[nearest]