"""AST Visualizer for Grimoire - Provides visual representation of AST"""

from typing import Optional, List, Dict, Any
import json
from .ast_nodes import *


class ASTVisualizer:
    """Visualize AST structure for debugging"""
    
    def __init__(self):
        self.indent_size = 2
        
    def visualize(self, node: ASTNode, indent: int = 0) -> str:
        """Convert AST node to visual string representation"""
        if node is None:
            return self._indent(indent) + "None"
            
        method_name = f"_visualize_{node.__class__.__name__.lower()}"
        method = getattr(self, method_name, self._visualize_default)
        return method(node, indent)
    
    def _indent(self, level: int) -> str:
        """Create indentation string"""
        return " " * (level * self.indent_size)
    
    def _visualize_default(self, node: ASTNode, indent: int) -> str:
        """Default visualization for unknown nodes"""
        return f"{self._indent(indent)}{node.__class__.__name__}"
    
    def _visualize_program(self, node: Program, indent: int) -> str:
        """Visualize Program node"""
        lines = [
            f"{self._indent(indent)}📦 Program",
            f"{self._indent(indent+1)}├─ 🔮 has_outer_circle: {node.has_outer_circle}"
        ]
        
        if node.main_entry:
            lines.append(f"{self._indent(indent+1)}├─ 🚪 main_entry:")
            lines.append(self.visualize(node.main_entry, indent+2))
        
        if node.functions:
            lines.append(f"{self._indent(indent+1)}├─ 🔧 functions: [{len(node.functions)}]")
            for func in node.functions:
                lines.append(self.visualize(func, indent+2))
        
        if node.globals:
            lines.append(f"{self._indent(indent+1)}└─ 🌍 globals: [{len(node.globals)}]")
            for stmt in node.globals:
                lines.append(self.visualize(stmt, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_functiondef(self, node: FunctionDef, indent: int) -> str:
        """Visualize FunctionDef node"""
        lines = [
            f"{self._indent(indent)}🔧 Function: {node.name}",
            f"{self._indent(indent+1)}├─ 📝 params: {', '.join(param.name for param in node.parameters)}"
        ]
        
        if node.body:
            lines.append(f"{self._indent(indent+1)}└─ 📋 body: [{len(node.body)}]")
            for stmt in node.body:
                lines.append(self.visualize(stmt, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_assignment(self, node: Assignment, indent: int) -> str:
        """Visualize Assignment node"""
        lines = [
            f"{self._indent(indent)}📌 Assignment: {node.target.name}",
            f"{self._indent(indent+1)}└─ value:"
        ]
        lines.append(self.visualize(node.value, indent+2))
        return "\n".join(lines)
    
    def _visualize_ifstatement(self, node: IfStatement, indent: int) -> str:
        """Visualize If node"""
        lines = [
            f"{self._indent(indent)}❓ If",
            f"{self._indent(indent+1)}├─ condition:"
        ]
        lines.append(self.visualize(node.condition, indent+2))
        
        if node.then_branch:
            lines.append(f"{self._indent(indent+1)}├─ then: [{len(node.then_branch)}]")
            for stmt in node.then_branch:
                lines.append(self.visualize(stmt, indent+2))
        
        if node.else_branch:
            lines.append(f"{self._indent(indent+1)}└─ else: [{len(node.else_branch)}]")
            for stmt in node.else_branch:
                lines.append(self.visualize(stmt, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_whileloop(self, node: WhileLoop, indent: int) -> str:
        """Visualize While node"""
        lines = [
            f"{self._indent(indent)}🔁 While",
            f"{self._indent(indent+1)}├─ condition:"
        ]
        lines.append(self.visualize(node.condition, indent+2))
        
        if node.body:
            lines.append(f"{self._indent(indent+1)}└─ body: [{len(node.body)}]")
            for stmt in node.body:
                lines.append(self.visualize(stmt, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_forloop(self, node: ForLoop, indent: int) -> str:
        """Visualize For node"""
        lines = [
            f"{self._indent(indent)}🔂 For: {node.counter.name}",
            f"{self._indent(indent+1)}├─ start:"
        ]
        lines.append(self.visualize(node.start, indent+2))
        lines.append(f"{self._indent(indent+1)}├─ end:")
        lines.append(self.visualize(node.end, indent+2))
        
        if node.step:
            lines.append(f"{self._indent(indent+1)}├─ step:")
            lines.append(self.visualize(node.step, indent+2))
        
        if node.body:
            lines.append(f"{self._indent(indent+1)}└─ body: [{len(node.body)}]")
            for stmt in node.body:
                lines.append(self.visualize(stmt, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_binaryop(self, node: BinaryOp, indent: int) -> str:
        """Visualize BinaryOp node"""
        op_symbols = {
            OperatorType.ADD: "+",
            OperatorType.SUBTRACT: "-",
            OperatorType.MULTIPLY: "*",
            OperatorType.DIVIDE: "/",
            OperatorType.EQUAL: "==",
            OperatorType.NOT_EQUAL: "!=",
            OperatorType.LESS_THAN: "<",
            OperatorType.GREATER_THAN: ">",
            OperatorType.LESS_EQUAL: "<=",
            OperatorType.GREATER_EQUAL: ">=",
            OperatorType.AND: "&&",
            OperatorType.OR: "||",
        }
        
        op = op_symbols.get(node.operator, str(node.operator))
        lines = [
            f"{self._indent(indent)}🔢 BinaryOp: {op}",
            f"{self._indent(indent+1)}├─ left:"
        ]
        lines.append(self.visualize(node.left, indent+2))
        lines.append(f"{self._indent(indent+1)}└─ right:")
        lines.append(self.visualize(node.right, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_unaryop(self, node: UnaryOp, indent: int) -> str:
        """Visualize UnaryOp node"""
        op = "!" if node.operator == OperatorType.NOT else str(node.operator)
        lines = [
            f"{self._indent(indent)}🔢 UnaryOp: {op}",
            f"{self._indent(indent+1)}└─ operand:"
        ]
        lines.append(self.visualize(node.operand, indent+2))
        return "\n".join(lines)
    
    def _visualize_literal(self, node: Literal, indent: int) -> str:
        """Visualize Literal node"""
        value_repr = repr(node.value)
        if len(value_repr) > 50:
            value_repr = value_repr[:47] + "..."
        return f"{self._indent(indent)}💎 Literal: {value_repr}"
    
    def _visualize_identifier(self, node: Identifier, indent: int) -> str:
        """Visualize Identifier node"""
        return f"{self._indent(indent)}🏷️  Identifier: {node.name}"
    
    def _visualize_functioncall(self, node: FunctionCall, indent: int) -> str:
        """Visualize FunctionCall node"""
        lines = [
            f"{self._indent(indent)}📞 FunctionCall: {node.name}",
        ]
        
        if node.args:
            lines.append(f"{self._indent(indent+1)}└─ args: [{len(node.args)}]")
            for arg in node.args:
                lines.append(self.visualize(arg, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_returnstatement(self, node: ReturnStatement, indent: int) -> str:
        """Visualize Return node"""
        lines = [
            f"{self._indent(indent)}↩️  Return",
        ]
        
        if node.value:
            lines.append(f"{self._indent(indent+1)}└─ value:")
            lines.append(self.visualize(node.value, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_outputstatement(self, node: OutputStatement, indent: int) -> str:
        """Visualize Print node"""
        lines = [
            f"{self._indent(indent)}🖨️  Output",
            f"{self._indent(indent+1)}└─ value:"
        ]
        lines.append(self.visualize(node.value, indent+2))
        return "\n".join(lines)
    
    def _visualize_parallelblock(self, node: ParallelBlock, indent: int) -> str:
        """Visualize ParallelBlock node"""
        lines = [
            f"{self._indent(indent)}⚡ ParallelBlock: [{len(node.branches)}]"
        ]
        
        for i, branch in enumerate(node.branches):
            prefix = "├─" if i < len(node.branches) - 1 else "└─"
            lines.append(f"{self._indent(indent+1)}{prefix} branch {i+1}: [{len(branch)}]")
            for stmt in branch:
                lines.append(self.visualize(stmt, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_arrayaccess(self, node: ArrayAccess, indent: int) -> str:
        """Visualize ArrayAccess node"""
        lines = [
            f"{self._indent(indent)}📇 ArrayAccess",
            f"{self._indent(indent+1)}├─ array:"
        ]
        lines.append(self.visualize(node.array, indent+2))
        lines.append(f"{self._indent(indent+1)}└─ index:")
        lines.append(self.visualize(node.index, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_arrayliteral(self, node: ArrayLiteral, indent: int) -> str:
        """Visualize ArrayLiteral node"""
        lines = [
            f"{self._indent(indent)}📇 Array: [{len(node.elements)}]"
        ]
        
        for i, elem in enumerate(node.elements):
            prefix = "├─" if i < len(node.elements) - 1 else "└─"
            lines.append(f"{self._indent(indent+1)}{prefix} [{i}]:")
            lines.append(self.visualize(elem, indent+2))
        
        return "\n".join(lines)
    
    def _visualize_mapliteral(self, node: MapLiteral, indent: int) -> str:
        """Visualize MapLiteral node"""
        lines = [
            f"{self._indent(indent)}🗺️  Map: [{len(node.pairs)}]"
        ]
        
        for i, (key, value) in enumerate(node.pairs):
            prefix = "├─" if i < len(node.pairs) - 1 else "└─"
            lines.append(f"{self._indent(indent+1)}{prefix} pair {i+1}:")
            lines.append(f"{self._indent(indent+2)}├─ key:")
            lines.append(self.visualize(key, indent+3))
            lines.append(f"{self._indent(indent+2)}└─ value:")
            lines.append(self.visualize(value, indent+3))
        
        return "\n".join(lines)


def create_execution_trace() -> 'ExecutionTracer':
    """Create a new execution tracer"""
    return ExecutionTracer()


class ExecutionTracer:
    """Trace execution flow for debugging"""
    
    def __init__(self):
        self.trace_entries: List[Dict[str, Any]] = []
        self.current_depth = 0
        
    def enter_function(self, name: str, args: List[Any]):
        """Record function entry"""
        self.trace_entries.append({
            "type": "enter",
            "name": name,
            "args": args,
            "depth": self.current_depth
        })
        self.current_depth += 1
        
    def exit_function(self, name: str, result: Any):
        """Record function exit"""
        self.current_depth -= 1
        self.trace_entries.append({
            "type": "exit",
            "name": name,
            "result": result,
            "depth": self.current_depth
        })
        
    def record_assignment(self, name: str, value: Any):
        """Record variable assignment"""
        self.trace_entries.append({
            "type": "assign",
            "name": name,
            "value": value,
            "depth": self.current_depth
        })
        
    def record_expression(self, expr: str, result: Any):
        """Record expression evaluation"""
        self.trace_entries.append({
            "type": "expr",
            "expr": expr,
            "result": result,
            "depth": self.current_depth
        })
        
    def format_trace(self) -> str:
        """Format trace as readable string"""
        lines = ["🔍 実行トレース:"]
        
        for entry in self.trace_entries:
            indent = "  " * entry["depth"]
            
            if entry["type"] == "enter":
                args_str = ", ".join(str(arg) for arg in entry["args"])
                lines.append(f"{indent}➡️  {entry['name']}({args_str})")
                
            elif entry["type"] == "exit":
                lines.append(f"{indent}⬅️  {entry['name']} → {entry['result']}")
                
            elif entry["type"] == "assign":
                lines.append(f"{indent}📌 {entry['name']} = {entry['value']}")
                
            elif entry["type"] == "expr":
                lines.append(f"{indent}🔢 {entry['expr']} → {entry['result']}")
        
        return "\n".join(lines)