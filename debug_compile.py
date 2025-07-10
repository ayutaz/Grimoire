#!/usr/bin/env python3
"""Debug the compilation process"""

from grimoire.image_recognition import MagicCircleDetector
from grimoire.parser import MagicCircleParser
from grimoire.code_generator import PythonCodeGenerator
from grimoire.ast_visualizer import ASTVisualizer

# Detect symbols
detector = MagicCircleDetector()
symbols, connections = detector.detect_symbols("examples/images/hello_world.png")

print("=== Detected Symbols ===")
for i, sym in enumerate(symbols):
    print(f"{i}: {sym.type.value} at {sym.position}, pattern: {sym.properties}")

print("\n=== Connections ===")
for conn in connections:
    from_idx = None
    to_idx = None
    for i, sym in enumerate(symbols):
        if (sym.type == conn.from_symbol.type and 
            sym.position == conn.from_symbol.position):
            from_idx = i
        if (sym.type == conn.to_symbol.type and 
            sym.position == conn.to_symbol.position):
            to_idx = i
    print(f"Connection: {from_idx} -> {to_idx}")

# Parse to AST
parser = MagicCircleParser()
ast = parser.parse(symbols, connections)

print("\n=== AST ===")
visualizer = ASTVisualizer()
visualizer.visualize(ast)

# Generate code
generator = PythonCodeGenerator()
code = generator.generate(ast)

print("\n=== Generated Code ===")
print(code)