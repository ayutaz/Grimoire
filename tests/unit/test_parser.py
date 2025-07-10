"""
Unit tests for parser module following TDD principles
Based on t-wada's testing approach with AAA pattern
"""

import pytest
from unittest.mock import MagicMock

from grimoire.parser import (
    MagicCircleParser, ParseError, SymbolNode
)
from grimoire.image_recognition import Symbol, SymbolType, Connection
from grimoire.ast_nodes import (
    Program, FunctionDef, Parameter, Statement, Expression,
    Assignment, IfStatement, WhileLoop, ForLoop, ParallelBlock,
    OutputStatement, ExpressionStatement, BinaryOp, Literal,
    Identifier, FunctionCall, DataType, OperatorType
)


class TestParseError:
    """Test ParseError exception"""
    
    def test_parse_error_creation(self):
        """Test creating a parse error"""
        # Arrange
        message = "Test error message"
        
        # Act
        error = ParseError(message)
        
        # Assert
        assert str(error) == message
        assert isinstance(error, Exception)


class TestSymbolNode:
    """Test SymbolNode dataclass"""
    
    def test_symbol_node_creation(self):
        """Test creating a symbol node with defaults"""
        # Arrange
        symbol = Symbol(SymbolType.CIRCLE, (100, 100), 50, 0.9, {})
        
        # Act
        node = SymbolNode(symbol)
        
        # Assert
        assert node.symbol == symbol
        assert node.visited is False
        assert node.ast_node is None
        assert node.parent is None
        assert node.children == []
    
    def test_symbol_node_with_relationships(self):
        """Test symbol node with parent and children"""
        # Arrange
        parent_symbol = Symbol(SymbolType.DOUBLE_CIRCLE, (200, 200), 60, 0.95, {})
        child_symbol = Symbol(SymbolType.SQUARE, (150, 150), 40, 0.88, {})
        
        parent_node = SymbolNode(parent_symbol)
        child_node = SymbolNode(child_symbol)
        
        # Act
        parent_node.children.append(child_node)
        child_node.parent = parent_node
        
        # Assert
        assert len(parent_node.children) == 1
        assert parent_node.children[0] == child_node
        assert child_node.parent == parent_node


class TestMagicCircleParser:
    """Test MagicCircleParser class following TDD principles"""
    
    def setup_method(self):
        """Set up test fixtures (Arrange phase for all tests)"""
        self.parser = MagicCircleParser()
        
    def create_basic_symbols(self):
        """Helper to create basic symbols for testing"""
        return [
            Symbol(SymbolType.OUTER_CIRCLE, (250, 250), 240, 0.98, {"is_double": False}),
            Symbol(SymbolType.DOUBLE_CIRCLE, (250, 200), 50, 0.95, {}),
            Symbol(SymbolType.STAR, (250, 300), 30, 0.90, {})
        ]
    
    def test_parser_initialization(self):
        """Test parser initializes with correct defaults"""
        # Assert
        assert self.parser.symbols == []
        assert self.parser.connections == []
        assert self.parser.symbol_graph == {}
        assert self.parser.symbol_index_map == {}
        assert self.parser.errors == []
    
    def test_parse_without_outer_circle(self):
        """Test parsing fails without outer circle"""
        # Arrange
        symbols = [
            Symbol(SymbolType.CIRCLE, (100, 100), 50, 0.9, {}),
            Symbol(SymbolType.SQUARE, (200, 200), 40, 0.85, {})
        ]
        connections = []
        
        # Act & Assert
        with pytest.raises(ParseError, match="No outer circle found"):
            self.parser.parse(symbols, connections)
    
    def test_parse_minimal_program(self):
        """Test parsing minimal valid program"""
        # Arrange
        symbols = [
            Symbol(SymbolType.OUTER_CIRCLE, (250, 250), 240, 0.98, {"is_double": False})
        ]
        connections = []
        
        # Act
        program = self.parser.parse(symbols, connections)
        
        # Assert
        assert isinstance(program, Program)
        assert program.has_outer_circle is True
        assert program.main_entry is None
        assert len(program.functions) == 0
        assert len(program.globals) == 0
    
    def test_parse_with_main_entry(self):
        """Test parsing program with main entry (double circle)"""
        # Arrange
        symbols = self.create_basic_symbols()
        connections = []
        
        # Act
        program = self.parser.parse(symbols, connections)
        
        # Assert
        assert program.main_entry is not None
        assert isinstance(program.main_entry, FunctionDef)
        assert program.main_entry.is_main is True
    
    def test_build_symbol_graph(self):
        """Test building symbol graph from symbols"""
        # Arrange
        symbols = self.create_basic_symbols()
        self.parser.symbols = symbols
        
        # Act
        self.parser._build_symbol_graph()
        
        # Assert
        assert len(self.parser.symbol_graph) == len(symbols)
        for i, symbol in enumerate(symbols):
            assert i in self.parser.symbol_graph
            assert self.parser.symbol_graph[i].symbol == symbol
    
    def test_build_symbol_graph_with_connections(self):
        """Test building symbol graph with explicit connections"""
        # Arrange
        symbols = [
            Symbol(SymbolType.OUTER_CIRCLE, (250, 250), 240, 0.98, {}),
            Symbol(SymbolType.SQUARE, (200, 200), 40, 0.85, {"pattern": "dot"}),
            Symbol(SymbolType.STAR, (300, 200), 30, 0.90, {})
        ]
        connections = [
            Connection(symbols[1], symbols[2], "solid")
        ]
        
        self.parser.symbols = symbols
        self.parser.connections = connections
        
        # Act
        self.parser._build_symbol_graph()
        
        # Assert
        # Find the connected nodes
        square_node = self.parser.symbol_graph[1]
        star_node = self.parser.symbol_graph[2]
        
        assert star_node in square_node.children
        assert square_node == star_node.parent
    
    def test_get_symbol_index(self):
        """Test getting symbol index by attributes"""
        # Arrange
        symbols = self.create_basic_symbols()
        self.parser.symbols = symbols
        
        # Act
        index = self.parser._get_symbol_index(symbols[1])
        
        # Assert
        assert index == 1
    
    def test_get_symbol_index_not_found(self):
        """Test getting symbol index for non-existent symbol"""
        # Arrange
        symbols = self.create_basic_symbols()
        self.parser.symbols = symbols
        other_symbol = Symbol(SymbolType.PENTAGON, (400, 400), 100, 0.7, {})
        
        # Act
        index = self.parser._get_symbol_index(other_symbol)
        
        # Assert
        assert index is None
    
    def test_find_outer_circle(self):
        """Test finding outer circle from symbols"""
        # Arrange
        symbols = self.create_basic_symbols()
        self.parser.symbols = symbols
        
        # Act
        outer = self.parser._find_outer_circle()
        
        # Assert
        assert outer is not None
        assert outer.type == SymbolType.OUTER_CIRCLE
    
    def test_parse_function_basic(self):
        """Test parsing a basic function"""
        # Arrange
        func_symbol = Symbol(SymbolType.CIRCLE, (150, 150), 40, 0.9, {})
        func_node = SymbolNode(func_symbol)
        
        # Act
        func_def = self.parser._parse_function(func_node)
        
        # Assert
        assert isinstance(func_def, FunctionDef)
        assert func_def.is_main is False
        assert func_node.visited is True
    
    def test_parse_function_with_body(self):
        """Test parsing function with body statements"""
        # Arrange
        func_symbol = Symbol(SymbolType.CIRCLE, (150, 150), 40, 0.9, {})
        star_symbol = Symbol(SymbolType.STAR, (150, 200), 30, 0.85, {})
        
        func_node = SymbolNode(func_symbol)
        star_node = SymbolNode(star_symbol)
        func_node.children.append(star_node)
        
        self.parser.symbol_graph = {0: func_node, 1: star_node}
        
        # Act
        func_def = self.parser._parse_function(func_node)
        
        # Assert
        assert len(func_def.body) > 0
        assert isinstance(func_def.body[0], OutputStatement)
    
    def test_parse_output_statement(self):
        """Test parsing output statement (star)"""
        # Arrange
        star_symbol = Symbol(SymbolType.STAR, (200, 200), 30, 0.9, {})
        star_node = SymbolNode(star_symbol)
        
        # Act
        stmt = self.parser._parse_statement(star_node)
        
        # Assert
        assert isinstance(stmt, OutputStatement)
        assert star_node.visited is True
    
    def test_parse_if_statement(self):
        """Test parsing if statement (triangle)"""
        # Arrange
        triangle_symbol = Symbol(SymbolType.TRIANGLE, (200, 200), 40, 0.9, {})
        triangle_node = SymbolNode(triangle_symbol)
        
        # Act
        stmt = self.parser._parse_statement(triangle_node)
        
        # Assert
        assert isinstance(stmt, IfStatement)
        assert triangle_node.visited is True
    
    def test_parse_loop_statement(self):
        """Test parsing loop statement (pentagon)"""
        # Arrange
        pentagon_symbol = Symbol(SymbolType.PENTAGON, (200, 200), 40, 0.9, {})
        pentagon_node = SymbolNode(pentagon_symbol)
        
        # Act
        stmt = self.parser._parse_statement(pentagon_node)
        
        # Assert
        assert isinstance(stmt, (ForLoop, WhileLoop))
        assert pentagon_node.visited is True
    
    def test_parse_parallel_block(self):
        """Test parsing parallel block (hexagon)"""
        # Arrange
        hexagon_symbol = Symbol(SymbolType.HEXAGON, (200, 200), 40, 0.9, {})
        hexagon_node = SymbolNode(hexagon_symbol)
        
        # Act
        stmt = self.parser._parse_statement(hexagon_node)
        
        # Assert
        assert isinstance(stmt, ParallelBlock)
        assert hexagon_node.visited is True
    
    def test_parse_binary_operation(self):
        """Test parsing binary operations"""
        # Arrange
        op_symbol = Symbol(SymbolType.CONVERGENCE, (200, 200), 20, 0.9, {})
        op_node = SymbolNode(op_symbol)
        
        # Create operands
        left_symbol = Symbol(SymbolType.SQUARE, (150, 200), 30, 0.85, {"pattern": "dot"})
        right_symbol = Symbol(SymbolType.SQUARE, (250, 200), 30, 0.85, {"pattern": "double_dot"})
        
        left_node = SymbolNode(left_symbol)
        right_node = SymbolNode(right_symbol)
        
        # Set up relationships
        self.parser.symbol_graph = {0: op_node, 1: left_node, 2: right_node}
        
        # Act
        expr = self.parser._parse_binary_op(op_node)
        
        # Assert
        assert isinstance(expr, BinaryOp)
        assert expr.operator == OperatorType.ADD
    
    def test_parse_literal_from_properties(self):
        """Test parsing literals from symbol properties"""
        # Arrange
        test_cases = [
            ({"pattern": "dot"}, 1, DataType.INTEGER),
            ({"pattern": "double_dot"}, 2, DataType.INTEGER),
            ({"pattern": "unknown"}, 0, DataType.INTEGER),
        ]
        
        for properties, expected_value, expected_type in test_cases:
            # Arrange
            symbol = Symbol(SymbolType.SQUARE, (100, 100), 30, 0.9, properties)
            node = SymbolNode(symbol)
            
            # Act
            literal = self.parser._parse_literal_from_properties(node)
            
            # Assert
            assert isinstance(literal, Literal)
            assert literal.value == expected_value
            assert literal.literal_type == expected_type
    
    def test_pattern_to_datatype(self):
        """Test pattern to datatype conversion"""
        # Arrange
        test_cases = [
            ("dot", DataType.INTEGER),
            ("double_dot", DataType.FLOAT),
            ("lines", DataType.STRING),
            ("half_circle", DataType.BOOLEAN),
            ("stars", DataType.ARRAY),
            ("grid", DataType.MAP),
            ("unknown", DataType.INTEGER),  # Default
        ]
        
        for pattern, expected_type in test_cases:
            # Act
            result = self.parser._pattern_to_datatype(pattern)
            
            # Assert
            assert result == expected_type
    
    def test_get_parents(self):
        """Test getting parent nodes"""
        # Arrange
        parent_symbol = Symbol(SymbolType.DOUBLE_CIRCLE, (200, 200), 60, 0.95, {})
        child_symbol = Symbol(SymbolType.SQUARE, (150, 150), 40, 0.88, {})
        
        parent_node = SymbolNode(parent_symbol)
        child_node = SymbolNode(child_symbol)
        parent_node.children.append(child_node)
        
        self.parser.symbol_graph = {0: parent_node, 1: child_node}
        
        # Act
        parents = self.parser._get_parents(child_node)
        
        # Assert
        assert len(parents) == 1
        assert parents[0] == parent_node
    
    def test_group_children_by_angle(self):
        """Test grouping children by angular position"""
        # Arrange
        center_symbol = Symbol(SymbolType.HEXAGON, (200, 200), 50, 0.9, {})
        center_node = SymbolNode(center_symbol)
        
        # Create children at different angles
        for i, angle in enumerate([0, 60, 120, 180, 240, 300]):
            x = 200 + 100 * np.cos(np.radians(angle))
            y = 200 + 100 * np.sin(np.radians(angle))
            child_symbol = Symbol(SymbolType.SQUARE, (int(x), int(y)), 20, 0.85, {})
            child_node = SymbolNode(child_symbol)
            center_node.children.append(child_node)
        
        # Act
        groups = self.parser._group_children_by_angle(center_node)
        
        # Assert
        assert len(groups) > 0
        assert all(isinstance(group, list) for group in groups)
    
    def test_infer_connections_basic(self):
        """Test inferring connections when none are explicit"""
        # Arrange
        symbols = [
            Symbol(SymbolType.OUTER_CIRCLE, (250, 250), 240, 0.98, {}),
            Symbol(SymbolType.DOUBLE_CIRCLE, (250, 200), 50, 0.95, {}),
            Symbol(SymbolType.STAR, (250, 300), 30, 0.90, {})
        ]
        
        self.parser.symbols = symbols
        self.parser._build_symbol_graph()
        
        # Act
        self.parser._infer_connections()
        
        # Assert
        # Main entry should connect to star below it
        main_node = self.parser.symbol_graph[1]
        star_node = self.parser.symbol_graph[2]
        assert star_node in main_node.children
    
    def test_parse_assignment(self):
        """Test parsing assignment statement"""
        # Arrange
        square_symbol = Symbol(SymbolType.SQUARE, (200, 200), 40, 0.85, {"pattern": "dot"})
        square_node = SymbolNode(square_symbol)
        
        # Act
        stmt = self.parser._parse_assignment(square_node)
        
        # Assert
        assert isinstance(stmt, Assignment)
        assert isinstance(stmt.target, Identifier)
        assert isinstance(stmt.value, Expression)
    
    # Error cases and edge scenarios
    
    def test_parse_empty_symbols(self):
        """Test parsing with empty symbols list"""
        # Arrange
        symbols = []
        connections = []
        
        # Act & Assert
        with pytest.raises(ParseError, match="No outer circle found"):
            self.parser.parse(symbols, connections)
    
    def test_parse_already_visited_node(self):
        """Test parsing already visited node"""
        # Arrange
        symbol = Symbol(SymbolType.STAR, (200, 200), 30, 0.9, {})
        node = SymbolNode(symbol)
        node.visited = True
        
        # Act
        stmt = self.parser._parse_statement(node)
        
        # Assert
        assert stmt is None
    
    def test_parse_with_multiple_main_entries(self):
        """Test parsing with multiple double circles (should use first)"""
        # Arrange
        symbols = [
            Symbol(SymbolType.OUTER_CIRCLE, (250, 250), 240, 0.98, {}),
            Symbol(SymbolType.DOUBLE_CIRCLE, (200, 200), 50, 0.95, {}),
            Symbol(SymbolType.DOUBLE_CIRCLE, (300, 200), 50, 0.95, {})
        ]
        connections = []
        
        # Act
        program = self.parser.parse(symbols, connections)
        
        # Assert
        assert program.main_entry is not None
        # First double circle should be used as main
    
    def test_parse_complex_expression(self):
        """Test parsing complex nested expression"""
        # Arrange
        symbols = [
            Symbol(SymbolType.OUTER_CIRCLE, (250, 250), 240, 0.98, {}),
            Symbol(SymbolType.SQUARE, (150, 200), 30, 0.85, {"pattern": "dot"}),
            Symbol(SymbolType.SQUARE, (250, 200), 30, 0.85, {"pattern": "double_dot"}),
            Symbol(SymbolType.CONVERGENCE, (200, 200), 20, 0.9, {}),
            Symbol(SymbolType.AMPLIFICATION, (200, 150), 20, 0.9, {}),
            Symbol(SymbolType.STAR, (200, 100), 30, 0.9, {})
        ]
        
        # Complex connections: (square1 + square2) * result -> star
        connections = [
            Connection(symbols[1], symbols[3], "solid"),
            Connection(symbols[2], symbols[3], "solid"),
            Connection(symbols[3], symbols[4], "solid"),
            Connection(symbols[4], symbols[5], "solid")
        ]
        
        # Act
        program = self.parser.parse(symbols, connections)
        
        # Assert
        assert len(program.globals) > 0


# Import numpy for angle calculations
import numpy as np