#!/usr/bin/env python3
"""Debug parser behavior"""

from grimoire.image_recognition import MagicCircleDetector
from grimoire.parser import MagicCircleParser
from grimoire.ast_visualizer import ASTVisualizer

# Detect symbols
detector = MagicCircleDetector()
symbols, connections = detector.detect_symbols("examples/images/hello_world.png")

print("=== Symbols ===")
for i, sym in enumerate(symbols):
    print(f"{i}: {sym.type.value} at {sym.position}")
    print(f"   Properties: {sym.properties}")

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
    print(f"From {from_idx} to {to_idx}")

# Parse
parser = MagicCircleParser()
ast = parser.parse(symbols, connections)

print("\n=== Symbol Graph (after parsing) ===")
for i, node in parser.symbol_graph.items():
    print(f"Node {i}: {node.symbol.type.value}")
    print(f"  visited: {node.visited}")
    print(f"  has parent: {node.parent is not None}")
    print(f"  children: {len(node.children)}")

print("\n=== AST Visualization ===")
visualizer = ASTVisualizer()
visualizer.visualize(ast)