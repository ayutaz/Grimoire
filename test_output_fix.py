#!/usr/bin/env python3
"""Test output statement parsing with debug info"""

from grimoire.image_recognition import Symbol, SymbolType, Connection
from grimoire.parser import MagicCircleParser

# Create symbols
symbols = [
    Symbol(type=SymbolType.OUTER_CIRCLE, position=(300, 300), size=300, confidence=1.0, properties={'is_double': False}),
    Symbol(type=SymbolType.SQUARE, position=(300, 200), size=50, confidence=1.0, properties={'pattern': 'triple_line'}),
    Symbol(type=SymbolType.STAR, position=(300, 400), size=50, confidence=1.0, properties={'points': 5})
]

# Create connection
connections = [
    Connection(from_symbol=symbols[1], to_symbol=symbols[2], connection_type='solid')
]

# Parse with debug
parser = MagicCircleParser()
parser.parse(symbols, connections)

# Check symbol graph
print("=== Symbol Graph ===")
for i, node in parser.symbol_graph.items():
    print(f"Node {i}: {node.symbol.type.value}")
    print(f"  visited: {node.visited}")
    print(f"  children: {[parser.symbols.index(child.symbol) for child in node.children]}")
    print(f"  parent: {parser.symbols.index(node.parent.symbol) if node.parent else None}")
    print()