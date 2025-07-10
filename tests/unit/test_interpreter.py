"""
Unit tests for interpreter module following TDD principles
Based on t-wada's testing approach with AAA pattern
"""

import pytest
from unittest.mock import patch, MagicMock
from concurrent.futures import ThreadPoolExecutor

from grimoire.interpreter import (
    GrimoireInterpreter, InterpreterError, ReturnValue, Environment
)
from grimoire.ast_nodes import (
    Program, FunctionDef, Parameter, Assignment, IfStatement,
    WhileLoop, ForLoop, ParallelBlock, OutputStatement, ReturnStatement,
    ExpressionStatement, BinaryOp, UnaryOp, Literal, Identifier,
    FunctionCall, ArrayLiteral, MapLiteral, ArrayAccess, MapAccess,
    DataType, OperatorType
)


class TestReturnValue:
    """Test ReturnValue exception for implementing returns"""
    
    def test_return_value_creation(self):
        """Test creating a return value exception"""
        # Arrange
        value = 42
        
        # Act
        ret = ReturnValue(value)
        
        # Assert
        assert ret.value == 42
        assert isinstance(ret, Exception)


class TestEnvironment:
    """Test Environment class for variable scoping"""
    
    def test_environment_creation(self):
        """Test creating an environment"""
        # Arrange & Act
        env = Environment()
        
        # Assert
        assert env.values == {}
        assert env.parent is None
    
    def test_environment_with_parent(self):
        """Test creating environment with parent"""
        # Arrange
        parent_env = Environment()
        
        # Act
        child_env = Environment(parent_env)
        
        # Assert
        assert child_env.parent == parent_env
    
    def test_define_and_get_variable(self):
        """Test defining and getting variables"""
        # Arrange
        env = Environment()
        
        # Act
        env.define("x", 42)
        result = env.get("x")
        
        # Assert
        assert result == 42
    
    def test_get_undefined_variable(self):
        """Test getting undefined variable raises error"""
        # Arrange
        env = Environment()
        
        # Act & Assert
        with pytest.raises(InterpreterError, match="Undefined variable: x"):
            env.get("x")
    
    def test_get_variable_from_parent(self):
        """Test getting variable from parent scope"""
        # Arrange
        parent_env = Environment()
        parent_env.define("x", 42)
        child_env = Environment(parent_env)
        
        # Act
        result = child_env.get("x")
        
        # Assert
        assert result == 42
    
    def test_set_variable(self):
        """Test setting existing variable"""
        # Arrange
        env = Environment()
        env.define("x", 42)
        
        # Act
        env.set("x", 100)
        result = env.get("x")
        
        # Assert
        assert result == 100
    
    def test_set_variable_in_parent(self):
        """Test setting variable in parent scope"""
        # Arrange
        parent_env = Environment()
        parent_env.define("x", 42)
        child_env = Environment(parent_env)
        
        # Act
        child_env.set("x", 100)
        
        # Assert
        assert parent_env.get("x") == 100
    
    def test_set_undefined_variable(self):
        """Test setting undefined variable raises error"""
        # Arrange
        env = Environment()
        
        # Act & Assert
        with pytest.raises(InterpreterError, match="Undefined variable: x"):
            env.set("x", 42)


class TestGrimoireInterpreter:
    """Test GrimoireInterpreter class following TDD principles"""
    
    def setup_method(self):
        """Set up test fixtures (Arrange phase for all tests)"""
        self.interpreter = GrimoireInterpreter()
    
    def test_interpreter_initialization(self):
        """Test interpreter initializes with correct defaults"""
        # Assert
        assert isinstance(self.interpreter.global_env, Environment)
        assert self.interpreter.current_env == self.interpreter.global_env
        assert self.interpreter.output_buffer == []
        assert self.interpreter.functions == {}
    
    def test_interpret_without_outer_circle(self):
        """Test interpretation fails without outer circle"""
        # Arrange
        program = Program(
            has_outer_circle=False,
            main_entry=None,
            functions=[],
            globals=[]
        )
        
        # Act & Assert
        with pytest.raises(InterpreterError, match="must be enclosed in a magic circle"):
            self.interpreter.interpret(program)
    
    def test_interpret_minimal_program(self):
        """Test interpreting minimal valid program"""
        # Arrange
        program = Program(
            has_outer_circle=True,
            main_entry=None,
            functions=[],
            globals=[]
        )
        
        # Act
        result = self.interpreter.interpret(program)
        
        # Assert
        assert result == ""
    
    def test_interpret_hello_world(self):
        """Test interpreting hello world program"""
        # Arrange
        output_stmt = OutputStatement(
            value=Literal(value="Hello, World!", literal_type=DataType.STRING)
        )
        program = Program(
            has_outer_circle=True,
            main_entry=None,
            functions=[],
            globals=[output_stmt]
        )
        
        # Act
        result = self.interpreter.interpret(program)
        
        # Assert
        assert result == "Hello, World!"
    
    def test_interpret_with_main_function(self):
        """Test interpreting program with main function"""
        # Arrange
        output_stmt = OutputStatement(
            value=Literal(value="From main", literal_type=DataType.STRING)
        )
        main_func = FunctionDef(
            name=None,
            parameters=[],
            body=[output_stmt],
            return_type=None,
            is_main=True
        )
        program = Program(
            has_outer_circle=True,
            main_entry=main_func,
            functions=[],
            globals=[]
        )
        
        # Act
        result = self.interpreter.interpret(program)
        
        # Assert
        assert result == "From main"
    
    def test_visit_literal(self):
        """Test visiting literal nodes"""
        # Arrange
        test_cases = [
            (42, DataType.INTEGER, 42),
            (3.14, DataType.FLOAT, 3.14),
            ("hello", DataType.STRING, "hello"),
            (True, DataType.BOOLEAN, True),
            (False, DataType.BOOLEAN, False),
        ]
        
        for value, dtype, expected in test_cases:
            # Arrange
            literal = Literal(value=value, literal_type=dtype)
            
            # Act
            result = self.interpreter.visit_literal(literal)
            
            # Assert
            assert result == expected
    
    def test_visit_identifier(self):
        """Test visiting identifier nodes"""
        # Arrange
        self.interpreter.current_env.define("x", 42)
        identifier = Identifier(name="x")
        
        # Act
        result = self.interpreter.visit_identifier(identifier)
        
        # Assert
        assert result == 42
    
    def test_visit_assignment(self):
        """Test visiting assignment statements"""
        # Arrange
        assignment = Assignment(
            target=Identifier(name="x"),
            value=Literal(value=42, literal_type=DataType.INTEGER)
        )
        
        # Act
        self.interpreter.visit_assignment(assignment)
        
        # Assert
        assert self.interpreter.current_env.get("x") == 42
    
    def test_visit_binary_operations(self):
        """Test visiting binary operations"""
        # Arrange
        test_cases = [
            (OperatorType.ADD, 5, 3, 8),
            (OperatorType.SUBTRACT, 5, 3, 2),
            (OperatorType.MULTIPLY, 5, 3, 15),
            (OperatorType.DIVIDE, 6, 2, 3),
            (OperatorType.EQUAL, 5, 5, True),
            (OperatorType.NOT_EQUAL, 5, 3, True),
            (OperatorType.LESS_THAN, 3, 5, True),
            (OperatorType.GREATER_THAN, 5, 3, True),
            (OperatorType.LESS_EQUAL, 3, 3, True),
            (OperatorType.GREATER_EQUAL, 5, 5, True),
        ]
        
        for op_type, left_val, right_val, expected in test_cases:
            # Arrange
            binary_op = BinaryOp(
                left=Literal(value=left_val, literal_type=DataType.INTEGER),
                operator=op_type,
                right=Literal(value=right_val, literal_type=DataType.INTEGER)
            )
            
            # Act
            result = self.interpreter.visit_binary_op(binary_op)
            
            # Assert
            assert result == expected
    
    def test_visit_binary_op_division_by_zero(self):
        """Test division by zero raises error"""
        # Arrange
        binary_op = BinaryOp(
            left=Literal(value=5, literal_type=DataType.INTEGER),
            operator=OperatorType.DIVIDE,
            right=Literal(value=0, literal_type=DataType.INTEGER)
        )
        
        # Act & Assert
        with pytest.raises(InterpreterError, match="Division by zero"):
            self.interpreter.visit_binary_op(binary_op)
    
    def test_visit_unary_operations(self):
        """Test visiting unary operations"""
        # Arrange
        test_cases = [
            (True, False),
            (False, True),
            (0, True),
            (1, False),
            ("", True),
            ("hello", False),
        ]
        
        for operand_val, expected in test_cases:
            # Arrange
            unary_op = UnaryOp(
                operator=OperatorType.NOT,
                operand=Literal(value=operand_val, literal_type=DataType.BOOLEAN)
            )
            
            # Act
            result = self.interpreter.visit_unary_op(unary_op)
            
            # Assert
            assert result == expected
    
    def test_visit_if_statement_true_branch(self):
        """Test if statement with true condition"""
        # Arrange
        if_stmt = IfStatement(
            condition=Literal(value=True, literal_type=DataType.BOOLEAN),
            then_branch=[
                Assignment(
                    target=Identifier(name="x"),
                    value=Literal(value=42, literal_type=DataType.INTEGER)
                )
            ],
            else_branch=None
        )
        
        # Act
        self.interpreter.visit_if_statement(if_stmt)
        
        # Assert
        assert self.interpreter.current_env.get("x") == 42
    
    def test_visit_if_statement_false_branch(self):
        """Test if statement with false condition"""
        # Arrange
        if_stmt = IfStatement(
            condition=Literal(value=False, literal_type=DataType.BOOLEAN),
            then_branch=[
                Assignment(
                    target=Identifier(name="x"),
                    value=Literal(value=42, literal_type=DataType.INTEGER)
                )
            ],
            else_branch=[
                Assignment(
                    target=Identifier(name="x"),
                    value=Literal(value=100, literal_type=DataType.INTEGER)
                )
            ]
        )
        
        # Act
        self.interpreter.visit_if_statement(if_stmt)
        
        # Assert
        assert self.interpreter.current_env.get("x") == 100
    
    def test_visit_while_loop(self):
        """Test while loop execution"""
        # Arrange
        self.interpreter.current_env.define("counter", 0)
        self.interpreter.current_env.define("sum", 0)
        
        while_loop = WhileLoop(
            condition=BinaryOp(
                left=Identifier(name="counter"),
                operator=OperatorType.LESS_THAN,
                right=Literal(value=3, literal_type=DataType.INTEGER)
            ),
            body=[
                Assignment(
                    target=Identifier(name="sum"),
                    value=BinaryOp(
                        left=Identifier(name="sum"),
                        operator=OperatorType.ADD,
                        right=Identifier(name="counter")
                    )
                ),
                Assignment(
                    target=Identifier(name="counter"),
                    value=BinaryOp(
                        left=Identifier(name="counter"),
                        operator=OperatorType.ADD,
                        right=Literal(value=1, literal_type=DataType.INTEGER)
                    )
                )
            ]
        )
        
        # Act
        self.interpreter.visit_while_loop(while_loop)
        
        # Assert
        assert self.interpreter.current_env.get("sum") == 3  # 0 + 1 + 2
        assert self.interpreter.current_env.get("counter") == 3
    
    def test_visit_for_loop(self):
        """Test for loop execution"""
        # Arrange
        self.interpreter.current_env.define("sum", 0)
        
        for_loop = ForLoop(
            counter=Identifier(name="i"),
            start=Literal(value=0, literal_type=DataType.INTEGER),
            end=Literal(value=5, literal_type=DataType.INTEGER),
            step=Literal(value=1, literal_type=DataType.INTEGER),
            body=[
                Assignment(
                    target=Identifier(name="sum"),
                    value=BinaryOp(
                        left=Identifier(name="sum"),
                        operator=OperatorType.ADD,
                        right=Identifier(name="i")
                    )
                )
            ]
        )
        
        # Act
        self.interpreter.visit_for_loop(for_loop)
        
        # Assert
        assert self.interpreter.current_env.get("sum") == 10  # 0 + 1 + 2 + 3 + 4
    
    def test_visit_output_statement(self):
        """Test output statement"""
        # Arrange
        output_stmt = OutputStatement(
            value=Literal(value="Test output", literal_type=DataType.STRING)
        )
        
        # Act
        self.interpreter.visit_output_statement(output_stmt)
        
        # Assert
        assert "Test output" in self.interpreter.output_buffer
    
    def test_visit_return_statement(self):
        """Test return statement raises ReturnValue"""
        # Arrange
        return_stmt = ReturnStatement(
            value=Literal(value=42, literal_type=DataType.INTEGER)
        )
        
        # Act & Assert
        with pytest.raises(ReturnValue) as exc_info:
            self.interpreter.visit_return_statement(return_stmt)
        
        assert exc_info.value.value == 42
    
    def test_execute_function_with_parameters(self):
        """Test executing function with parameters"""
        # Arrange
        func = FunctionDef(
            name="add",
            parameters=[
                Parameter(name="a", data_type=DataType.INTEGER),
                Parameter(name="b", data_type=DataType.INTEGER)
            ],
            body=[
                ReturnStatement(
                    value=BinaryOp(
                        left=Identifier(name="a"),
                        operator=OperatorType.ADD,
                        right=Identifier(name="b")
                    )
                )
            ],
            return_type=DataType.INTEGER,
            is_main=False
        )
        
        # Act
        result = self.interpreter.execute_function(func, [5, 3])
        
        # Assert
        assert result == 8
    
    def test_visit_array_literal(self):
        """Test visiting array literal"""
        # Arrange
        array_lit = ArrayLiteral(
            elements=[
                Literal(value=1, literal_type=DataType.INTEGER),
                Literal(value=2, literal_type=DataType.INTEGER),
                Literal(value=3, literal_type=DataType.INTEGER)
            ]
        )
        
        # Act
        result = self.interpreter.visit_array_literal(array_lit)
        
        # Assert
        assert result == [1, 2, 3]
    
    def test_visit_array_access(self):
        """Test array access"""
        # Arrange
        self.interpreter.current_env.define("arr", [10, 20, 30])
        array_access = ArrayAccess(
            array=Identifier(name="arr"),
            index=Literal(value=1, literal_type=DataType.INTEGER)
        )
        
        # Act
        result = self.interpreter.visit_array_access(array_access)
        
        # Assert
        assert result == 20
    
    def test_visit_array_access_out_of_bounds(self):
        """Test array access with out of bounds index"""
        # Arrange
        self.interpreter.current_env.define("arr", [10, 20, 30])
        array_access = ArrayAccess(
            array=Identifier(name="arr"),
            index=Literal(value=5, literal_type=DataType.INTEGER)
        )
        
        # Act & Assert
        with pytest.raises(InterpreterError, match="Array index out of bounds"):
            self.interpreter.visit_array_access(array_access)
    
    def test_visit_map_literal(self):
        """Test visiting map literal"""
        # Arrange
        map_lit = MapLiteral(
            pairs=[
                (
                    Literal(value="name", literal_type=DataType.STRING),
                    Literal(value="Alice", literal_type=DataType.STRING)
                ),
                (
                    Literal(value="age", literal_type=DataType.STRING),
                    Literal(value=30, literal_type=DataType.INTEGER)
                )
            ]
        )
        
        # Act
        result = self.interpreter.visit_map_literal(map_lit)
        
        # Assert
        assert result == {"name": "Alice", "age": 30}
    
    def test_visit_map_access(self):
        """Test map access"""
        # Arrange
        self.interpreter.current_env.define("person", {"name": "Bob", "age": 25})
        map_access = MapAccess(
            map_expr=Identifier(name="person"),
            key=Literal(value="name", literal_type=DataType.STRING)
        )
        
        # Act
        result = self.interpreter.visit_map_access(map_access)
        
        # Assert
        assert result == "Bob"
    
    def test_visit_map_access_key_not_found(self):
        """Test map access with non-existent key"""
        # Arrange
        self.interpreter.current_env.define("person", {"name": "Bob"})
        map_access = MapAccess(
            map_expr=Identifier(name="person"),
            key=Literal(value="age", literal_type=DataType.STRING)
        )
        
        # Act & Assert
        with pytest.raises(InterpreterError, match="Key not found in map"):
            self.interpreter.visit_map_access(map_access)
    
    def test_is_truthy(self):
        """Test truthiness evaluation"""
        # Arrange
        test_cases = [
            (True, True),
            (False, False),
            (None, False),
            (0, False),
            (1, True),
            (-1, True),
            (0.0, False),
            (3.14, True),
            ("", False),
            ("hello", True),
            ([], False),
            ([1, 2], True),
            ({}, False),
            ({"a": 1}, True),
        ]
        
        for value, expected in test_cases:
            # Act
            result = self.interpreter._is_truthy(value)
            
            # Assert
            assert result == expected
    
    def test_value_to_string(self):
        """Test value to string conversion"""
        # Arrange
        test_cases = [
            (None, "∅"),
            (True, "☀"),
            (False, "☾"),
            (0, "∅"),
            (1, "•"),
            (2, "••"),
            (3, "•••"),
            (10, "⦿"),
            (42, "42"),
            (3.14, "3.14"),
            ("hello", "hello"),
            ([1, 2, 3], "[•, ••, •••]"),
            ({"a": 1, "b": 2}, "{a→•, b→••}"),
        ]
        
        for value, expected in test_cases:
            # Act
            result = self.interpreter._value_to_string(value)
            
            # Assert
            assert result == expected
    
    def test_parallel_block_execution(self):
        """Test parallel block execution"""
        # Arrange
        # Create shared state
        self.interpreter.global_env.define("results", [])
        
        parallel_block = ParallelBlock(
            branches=[
                # Branch 1
                [
                    OutputStatement(
                        value=Literal(value="Branch 1", literal_type=DataType.STRING)
                    )
                ],
                # Branch 2
                [
                    OutputStatement(
                        value=Literal(value="Branch 2", literal_type=DataType.STRING)
                    )
                ]
            ]
        )
        
        # Act
        self.interpreter.visit_parallel_block(parallel_block)
        
        # Assert
        output_str = '\n'.join(self.interpreter.output_buffer)
        assert "Branch 1" in output_str
        assert "Branch 2" in output_str
    
    def test_call_builtin_functions(self):
        """Test calling built-in functions"""
        # Arrange
        test_cases = [
            ("len", [[1, 2, 3]], 3),
            ("range", [3], [0, 1, 2]),
            ("range", [1, 4], [1, 2, 3]),
        ]
        
        for func_name, args, expected in test_cases:
            # Act
            result = self.interpreter._call_builtin(func_name, args)
            
            # Assert
            assert result == expected
    
    def test_call_unknown_builtin(self):
        """Test calling unknown built-in function"""
        # Arrange & Act & Assert
        with pytest.raises(InterpreterError, match="Unknown function: unknown"):
            self.interpreter._call_builtin("unknown", [])
    
    # Edge cases and error scenarios
    
    def test_nested_function_calls(self):
        """Test nested function calls with proper scoping"""
        # Arrange
        inner_func = FunctionDef(
            name="inner",
            parameters=[Parameter(name="x", data_type=DataType.INTEGER)],
            body=[
                ReturnStatement(
                    value=BinaryOp(
                        left=Identifier(name="x"),
                        operator=OperatorType.MULTIPLY,
                        right=Literal(value=2, literal_type=DataType.INTEGER)
                    )
                )
            ],
            return_type=DataType.INTEGER,
            is_main=False
        )
        
        self.interpreter.functions["inner"] = inner_func
        
        outer_func = FunctionDef(
            name="outer",
            parameters=[Parameter(name="y", data_type=DataType.INTEGER)],
            body=[
                ReturnStatement(
                    value=FunctionCall(
                        function=Identifier(name="inner"),
                        arguments=[
                            BinaryOp(
                                left=Identifier(name="y"),
                                operator=OperatorType.ADD,
                                right=Literal(value=1, literal_type=DataType.INTEGER)
                            )
                        ]
                    )
                )
            ],
            return_type=DataType.INTEGER,
            is_main=False
        )
        
        # Act
        result = self.interpreter.execute_function(outer_func, [5])
        
        # Assert
        assert result == 12  # (5 + 1) * 2
    
    def test_complex_expression_evaluation(self):
        """Test evaluating complex nested expressions"""
        # Arrange
        # (2 + 3) * (4 - 1)
        expr = BinaryOp(
            left=BinaryOp(
                left=Literal(value=2, literal_type=DataType.INTEGER),
                operator=OperatorType.ADD,
                right=Literal(value=3, literal_type=DataType.INTEGER)
            ),
            operator=OperatorType.MULTIPLY,
            right=BinaryOp(
                left=Literal(value=4, literal_type=DataType.INTEGER),
                operator=OperatorType.SUBTRACT,
                right=Literal(value=1, literal_type=DataType.INTEGER)
            )
        )
        
        # Act
        result = self.interpreter.execute(expr)
        
        # Assert
        assert result == 15  # 5 * 3