"""Interpreter for Grimoire - Executes AST"""

from typing import Any, Dict, List, Optional
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
import threading

from .ast_nodes import *


class InterpreterError(Exception):
    """Runtime error during interpretation"""
    pass


class ReturnValue(Exception):
    """Used to implement return statements"""
    def __init__(self, value):
        self.value = value
        super().__init__()


class Environment:
    """Variable environment for scoping"""
    
    def __init__(self, parent: Optional['Environment'] = None):
        self.values: Dict[str, Any] = {}
        self.parent = parent
    
    def define(self, name: str, value: Any):
        """Define a variable in current scope"""
        self.values[name] = value
    
    def get(self, name: str) -> Any:
        """Get variable value, checking parent scopes"""
        if name in self.values:
            return self.values[name]
        if self.parent:
            return self.parent.get(name)
        raise InterpreterError(f"Undefined variable: {name}")
    
    def set(self, name: str, value: Any):
        """Set variable value, checking parent scopes"""
        if name in self.values:
            self.values[name] = value
        elif self.parent:
            self.parent.set(name, value)
        else:
            raise InterpreterError(f"Undefined variable: {name}")


class GrimoireInterpreter(ASTVisitor):
    """Interpreter that executes Grimoire AST"""
    
    def __init__(self):
        self.global_env = Environment()
        self.current_env = self.global_env
        self.output_buffer = []
        self.functions = {}  # Store function definitions
    
    def interpret(self, program: Program) -> str:
        """Interpret a program and return output"""
        self.output_buffer = []
        
        # Check outer circle
        if not program.has_outer_circle:
            raise InterpreterError("Program must be enclosed in a magic circle (outer circle)")
        
        # Store function definitions
        for func in program.functions:
            self.functions[id(func)] = func
        
        if program.main_entry:
            self.functions['main'] = program.main_entry
        
        # Execute global statements
        for stmt in program.globals:
            self.execute(stmt)
        
        # Execute main entry if present
        if program.main_entry:
            try:
                self.execute_function(program.main_entry, [])
            except ReturnValue:
                pass  # Main function returned
        
        return '\n'.join(str(x) for x in self.output_buffer)
    
    def execute(self, node: ASTNode) -> Any:
        """Execute a node and return its value"""
        return node.accept(self)
    
    def execute_function(self, func: FunctionDef, args: List[Any]) -> Any:
        """Execute a function with given arguments"""
        # Create new environment for function
        func_env = Environment(self.global_env)
        
        # Bind parameters
        for i, param in enumerate(func.parameters):
            if i < len(args):
                func_env.define(param.name, args[i])
            elif param.default_value:
                func_env.define(param.name, self.execute(param.default_value))
            else:
                raise InterpreterError(f"Missing argument for parameter {param.name}")
        
        # Save current environment
        prev_env = self.current_env
        self.current_env = func_env
        
        try:
            # Execute function body
            for stmt in func.body:
                self.execute(stmt)
            return None
        except ReturnValue as ret:
            return ret.value
        finally:
            # Restore environment
            self.current_env = prev_env
    
    # Visitor methods
    
    def visit_program(self, node: Program) -> None:
        # Program is handled by interpret()
        pass
    
    def visit_function_def(self, node: FunctionDef) -> None:
        # Function definitions are stored, not executed
        pass
    
    def visit_parameter(self, node: Parameter) -> None:
        # Parameters are handled during function calls
        pass
    
    def visit_expression_statement(self, node: ExpressionStatement) -> None:
        self.execute(node.expression)
    
    def visit_assignment(self, node: Assignment) -> None:
        value = self.execute(node.value)
        self.current_env.define(node.target.name, value)
    
    def visit_if_statement(self, node: IfStatement) -> None:
        condition = self.execute(node.condition)
        
        if self._is_truthy(condition):
            for stmt in node.then_branch:
                self.execute(stmt)
        elif node.else_branch:
            for stmt in node.else_branch:
                self.execute(stmt)
    
    def visit_while_loop(self, node: WhileLoop) -> None:
        while self._is_truthy(self.execute(node.condition)):
            for stmt in node.body:
                self.execute(stmt)
    
    def visit_for_loop(self, node: ForLoop) -> None:
        # Initialize counter
        start = self.execute(node.start)
        end = self.execute(node.end)
        step = self.execute(node.step) if node.step else 1
        
        # Create loop variable
        self.current_env.define(node.counter.name, start)
        
        # Execute loop
        current = start
        while (step > 0 and current < end) or (step < 0 and current > end):
            for stmt in node.body:
                self.execute(stmt)
            
            current += step
            self.current_env.set(node.counter.name, current)
    
    def visit_parallel_block(self, node: ParallelBlock) -> None:
        """Execute branches in parallel"""
        results = []
        
        with ThreadPoolExecutor(max_workers=len(node.branches)) as executor:
            # Submit each branch for execution
            futures = []
            for branch in node.branches:
                future = executor.submit(self._execute_branch, branch)
                futures.append(future)
            
            # Wait for all branches to complete
            for future in as_completed(futures):
                try:
                    result = future.result()
                    results.append(result)
                except Exception as e:
                    raise InterpreterError(f"Error in parallel branch: {e}")
        
        # Merge results (for now, just add all outputs)
        for result in results:
            if result:
                self.output_buffer.extend(result)
    
    def _execute_branch(self, statements: List[Statement]) -> List[Any]:
        """Execute a branch in its own environment"""
        # Create a new interpreter instance for thread safety
        branch_interpreter = GrimoireInterpreter()
        branch_interpreter.global_env = self.global_env  # Share globals
        branch_interpreter.functions = self.functions
        branch_interpreter.current_env = Environment(self.global_env)
        
        for stmt in statements:
            branch_interpreter.execute(stmt)
        
        return branch_interpreter.output_buffer
    
    def visit_return_statement(self, node: ReturnStatement) -> None:
        value = self.execute(node.value) if node.value else None
        raise ReturnValue(value)
    
    def visit_output_statement(self, node: OutputStatement) -> None:
        value = self.execute(node.value)
        self.output_buffer.append(self._value_to_string(value))
    
    def visit_binary_op(self, node: BinaryOp) -> Any:
        left = self.execute(node.left)
        right = self.execute(node.right)
        
        op_map = {
            OperatorType.ADD: lambda l, r: l + r,
            OperatorType.SUBTRACT: lambda l, r: l - r,
            OperatorType.MULTIPLY: lambda l, r: l * r,
            OperatorType.DIVIDE: lambda l, r: l / r if r != 0 else self._error("Division by zero"),
            OperatorType.EQUAL: lambda l, r: l == r,
            OperatorType.NOT_EQUAL: lambda l, r: l != r,
            OperatorType.LESS_THAN: lambda l, r: l < r,
            OperatorType.GREATER_THAN: lambda l, r: l > r,
            OperatorType.LESS_EQUAL: lambda l, r: l <= r,
            OperatorType.GREATER_EQUAL: lambda l, r: l >= r,
            OperatorType.AND: lambda l, r: self._is_truthy(l) and self._is_truthy(r),
            OperatorType.OR: lambda l, r: self._is_truthy(l) or self._is_truthy(r),
        }
        
        op_func = op_map.get(node.operator)
        if op_func:
            return op_func(left, right)
        else:
            raise InterpreterError(f"Unknown operator: {node.operator}")
    
    def visit_unary_op(self, node: UnaryOp) -> Any:
        operand = self.execute(node.operand)
        
        if node.operator == OperatorType.NOT:
            return not self._is_truthy(operand)
        else:
            raise InterpreterError(f"Unknown unary operator: {node.operator}")
    
    def visit_literal(self, node: Literal) -> Any:
        return node.value
    
    def visit_identifier(self, node: Identifier) -> Any:
        return self.current_env.get(node.name)
    
    def visit_function_call(self, node: FunctionCall) -> Any:
        # Evaluate arguments
        args = [self.execute(arg) for arg in node.arguments]
        
        # Get function
        if isinstance(node.function, FunctionDef):
            return self.execute_function(node.function, args)
        elif isinstance(node.function, Identifier):
            func_name = node.function.name
            if func_name in self.functions:
                return self.execute_function(self.functions[func_name], args)
            else:
                # Check if it's a built-in function
                return self._call_builtin(func_name, args)
        else:
            raise InterpreterError("Invalid function call")
    
    def visit_array_literal(self, node: ArrayLiteral) -> List[Any]:
        return [self.execute(elem) for elem in node.elements]
    
    def visit_map_literal(self, node: MapLiteral) -> Dict[Any, Any]:
        result = {}
        for key_expr, value_expr in node.pairs:
            key = self.execute(key_expr)
            value = self.execute(value_expr)
            result[key] = value
        return result
    
    def visit_array_access(self, node: ArrayAccess) -> Any:
        array = self.execute(node.array)
        index = self.execute(node.index)
        
        if not isinstance(array, list):
            raise InterpreterError("Array access on non-array")
        
        if not isinstance(index, int):
            raise InterpreterError("Array index must be integer")
        
        if 0 <= index < len(array):
            return array[index]
        else:
            raise InterpreterError(f"Array index out of bounds: {index}")
    
    def visit_map_access(self, node: MapAccess) -> Any:
        map_value = self.execute(node.map_expr)
        key = self.execute(node.key)
        
        if not isinstance(map_value, dict):
            raise InterpreterError("Map access on non-map")
        
        if key in map_value:
            return map_value[key]
        else:
            raise InterpreterError(f"Key not found in map: {key}")
    
    # Helper methods
    
    def _is_truthy(self, value: Any) -> bool:
        """Determine if a value is truthy"""
        if isinstance(value, bool):
            return value
        elif value is None:
            return False
        elif isinstance(value, (int, float)):
            return value != 0
        elif isinstance(value, (str, list, dict)):
            return len(value) > 0
        else:
            return True
    
    def _value_to_string(self, value: Any) -> str:
        """Convert value to string for output"""
        if value is None:
            return "∅"
        elif isinstance(value, bool):
            return "☀" if value else "☾"
        elif isinstance(value, (int, float)):
            # Convert numbers to dot notation
            if isinstance(value, int) and 0 <= value <= 10:
                if value == 0:
                    return "∅"
                elif value == 1:
                    return "•"
                elif value == 2:
                    return "••"
                elif value == 3:
                    return "•••"
                elif value == 10:
                    return "⦿"
            return str(value)
        elif isinstance(value, str):
            return value
        elif isinstance(value, list):
            return f"[{', '.join(self._value_to_string(v) for v in value)}]"
        elif isinstance(value, dict):
            items = [f"{self._value_to_string(k)}→{self._value_to_string(v)}" 
                    for k, v in value.items()]
            return f"{{{', '.join(items)}}}"
        else:
            return str(value)
    
    def _error(self, message: str):
        """Raise an interpreter error"""
        raise InterpreterError(message)
    
    def _call_builtin(self, name: str, args: List[Any]) -> Any:
        """Call built-in function"""
        builtins = {
            'len': lambda x: len(x) if hasattr(x, '__len__') else self._error("len() requires sequence"),
            'range': lambda *args: list(range(*args)),
            'print': lambda *args: self.output_buffer.append(' '.join(str(a) for a in args)),
        }
        
        if name in builtins:
            return builtins[name](*args)
        else:
            raise InterpreterError(f"Unknown function: {name}")