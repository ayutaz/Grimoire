"""Abstract Syntax Tree nodes for Grimoire"""

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import List, Optional, Any, Union, Tuple
from enum import Enum


class DataType(Enum):
    """Data types in Grimoire"""
    INTEGER = "integer"      # •
    FLOAT = "float"         # ••
    STRING = "string"       # ≡
    BOOLEAN = "boolean"     # ◐
    ARRAY = "array"         # ※
    MAP = "map"            # ⊞
    VOID = "void"          # ∅


class OperatorType(Enum):
    """Operator types"""
    ADD = "add"            # ⟐ (convergence)
    SUBTRACT = "subtract"  # ⟑ (divergence)
    MULTIPLY = "multiply"  # ✦ (amplification)
    DIVIDE = "divide"      # ⟠ (distribution)
    ASSIGN = "assign"      # ⟷ (transfer)
    
    # Comparison
    EQUAL = "equal"
    NOT_EQUAL = "not_equal"
    LESS_THAN = "less_than"
    GREATER_THAN = "greater_than"
    LESS_EQUAL = "less_equal"
    GREATER_EQUAL = "greater_equal"
    
    # Logical
    AND = "and"
    OR = "or"
    NOT = "not"
    LOGICAL_AND = "logical_and"
    LOGICAL_OR = "logical_or"
    LOGICAL_NOT = "logical_not"
    LOGICAL_XOR = "logical_xor"


class ASTNode(ABC):
    """Base class for all AST nodes"""
    
    @abstractmethod
    def accept(self, visitor):
        """Accept a visitor for the visitor pattern"""
        pass


@dataclass
class Program(ASTNode):
    """Root node representing a magic circle program"""
    has_outer_circle: bool
    main_entry: Optional['FunctionDef']
    functions: List['FunctionDef']
    globals: List['Statement']
    
    def accept(self, visitor):
        return visitor.visit_program(self)


@dataclass
class FunctionDef(ASTNode):
    """Function definition (circle with parameters)"""
    name: Optional[str]  # None for anonymous functions
    parameters: List['Parameter']
    body: List['Statement']
    return_type: Optional[DataType]
    is_main: bool = False
    
    def accept(self, visitor):
        return visitor.visit_function_def(self)


@dataclass
class Parameter(ASTNode):
    """Function parameter"""
    name: str
    data_type: DataType
    default_value: Optional['Expression'] = None
    
    def accept(self, visitor):
        return visitor.visit_parameter(self)


# Statements

class Statement(ASTNode):
    """Base class for statements"""
    pass


@dataclass
class ExpressionStatement(Statement):
    """Expression as a statement"""
    expression: 'Expression'
    
    def accept(self, visitor):
        return visitor.visit_expression_statement(self)


@dataclass
class Assignment(Statement):
    """Assignment statement (⟷)"""
    target: 'Identifier'
    value: 'Expression'
    
    def accept(self, visitor):
        return visitor.visit_assignment(self)


@dataclass
class IfStatement(Statement):
    """Conditional statement (△)"""
    condition: 'Expression'
    then_branch: List[Statement]
    else_branch: Optional[List[Statement]] = None
    
    def accept(self, visitor):
        return visitor.visit_if_statement(self)


@dataclass
class WhileLoop(Statement):
    """While loop (⬟ with condition)"""
    condition: 'Expression'
    body: List[Statement]
    
    def accept(self, visitor):
        return visitor.visit_while_loop(self)


@dataclass
class ForLoop(Statement):
    """For loop (⬟ with counter)"""
    counter: 'Identifier'
    start: 'Expression'
    end: 'Expression'
    step: Optional['Expression']
    body: List[Statement]
    
    def accept(self, visitor):
        return visitor.visit_for_loop(self)


@dataclass
class ParallelBlock(Statement):
    """Parallel execution block (⬢)"""
    branches: List[List[Statement]]
    
    def accept(self, visitor):
        return visitor.visit_parallel_block(self)


@dataclass
class ReturnStatement(Statement):
    """Return statement"""
    value: Optional['Expression']
    
    def accept(self, visitor):
        return visitor.visit_return_statement(self)


@dataclass
class OutputStatement(Statement):
    """Output statement (☆)"""
    value: 'Expression'
    
    def accept(self, visitor):
        return visitor.visit_output_statement(self)


# Expressions

class Expression(ASTNode):
    """Base class for expressions"""
    data_type: Optional[DataType] = None


@dataclass
class BinaryOp(Expression):
    """Binary operation"""
    left: Expression
    operator: OperatorType
    right: Expression
    
    def accept(self, visitor):
        return visitor.visit_binary_op(self)


@dataclass
class UnaryOp(Expression):
    """Unary operation"""
    operator: OperatorType
    operand: Expression
    
    def accept(self, visitor):
        return visitor.visit_unary_op(self)


@dataclass
class Literal(Expression):
    """Literal value"""
    value: Any
    literal_type: DataType
    
    def __post_init__(self):
        self.data_type = self.literal_type
    
    def accept(self, visitor):
        return visitor.visit_literal(self)


@dataclass
class Identifier(Expression):
    """Variable reference"""
    name: str
    
    def accept(self, visitor):
        return visitor.visit_identifier(self)


@dataclass
class FunctionCall(Expression):
    """Function call"""
    function: Union[Identifier, 'FunctionDef']
    arguments: List[Expression]
    
    def accept(self, visitor):
        return visitor.visit_function_call(self)


@dataclass
class ArrayLiteral(Expression):
    """Array literal (※)"""
    elements: List[Expression]
    
    def __post_init__(self):
        self.data_type = DataType.ARRAY
    
    def accept(self, visitor):
        return visitor.visit_array_literal(self)


@dataclass
class MapLiteral(Expression):
    """Map literal (⊞)"""
    pairs: List[Tuple[Expression, Expression]]
    
    def __post_init__(self):
        self.data_type = DataType.MAP
    
    def accept(self, visitor):
        return visitor.visit_map_literal(self)


@dataclass
class ArrayAccess(Expression):
    """Array element access"""
    array: Expression
    index: Expression
    
    def accept(self, visitor):
        return visitor.visit_array_access(self)


@dataclass
class MapAccess(Expression):
    """Map element access"""
    map_expr: Expression
    key: Expression
    
    def accept(self, visitor):
        return visitor.visit_map_access(self)


# Visitor interface

class ASTVisitor(ABC):
    """Visitor interface for traversing the AST"""
    
    @abstractmethod
    def visit_program(self, node: Program): pass
    
    @abstractmethod
    def visit_function_def(self, node: FunctionDef): pass
    
    @abstractmethod
    def visit_parameter(self, node: Parameter): pass
    
    @abstractmethod
    def visit_expression_statement(self, node: ExpressionStatement): pass
    
    @abstractmethod
    def visit_assignment(self, node: Assignment): pass
    
    @abstractmethod
    def visit_if_statement(self, node: IfStatement): pass
    
    @abstractmethod
    def visit_while_loop(self, node: WhileLoop): pass
    
    @abstractmethod
    def visit_for_loop(self, node: ForLoop): pass
    
    @abstractmethod
    def visit_parallel_block(self, node: ParallelBlock): pass
    
    @abstractmethod
    def visit_return_statement(self, node: ReturnStatement): pass
    
    @abstractmethod
    def visit_output_statement(self, node: OutputStatement): pass
    
    @abstractmethod
    def visit_binary_op(self, node: BinaryOp): pass
    
    @abstractmethod
    def visit_unary_op(self, node: UnaryOp): pass
    
    @abstractmethod
    def visit_literal(self, node: Literal): pass
    
    @abstractmethod
    def visit_identifier(self, node: Identifier): pass
    
    @abstractmethod
    def visit_function_call(self, node: FunctionCall): pass
    
    @abstractmethod
    def visit_array_literal(self, node: ArrayLiteral): pass
    
    @abstractmethod
    def visit_map_literal(self, node: MapLiteral): pass
    
    @abstractmethod
    def visit_array_access(self, node: ArrayAccess): pass
    
    @abstractmethod
    def visit_map_access(self, node: MapAccess): pass