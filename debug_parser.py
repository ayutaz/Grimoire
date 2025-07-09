#!/usr/bin/env python3
"""Debug parser connection inference"""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from grimoire.image_recognition import Symbol, SymbolType, Connection
from grimoire.parser import MagicCircleParser


def debug_connections():
    """Debug connection inference"""
    symbols = [
        # Outer circle (mandatory)
        Symbol(
            type=SymbolType.OUTER_CIRCLE,
            position=(300, 300),
            size=280,
            confidence=1.0,
            properties={}
        ),
        # Main entry
        Symbol(
            type=SymbolType.DOUBLE_CIRCLE,
            position=(300, 100),
            size=30,
            confidence=0.9,
            properties={}
        ),
        # Number 1
        Symbol(
            type=SymbolType.SQUARE,
            position=(200, 200),
            size=40,
            confidence=0.9,
            properties={"pattern": "dot"}  # Single dot = 1
        ),
        # Number 2
        Symbol(
            type=SymbolType.SQUARE,
            position=(400, 200),
            size=40,
            confidence=0.9,
            properties={"pattern": "double_dot"}  # Two dots = 2
        ),
        # Addition operator
        Symbol(
            type=SymbolType.CONVERGENCE,
            position=(300, 200),
            size=20,
            confidence=0.8,
            properties={}
        ),
        # Output
        Symbol(
            type=SymbolType.STAR,
            position=(300, 300),
            size=40,
            confidence=0.9,
            properties={}
        )
    ]
    
    parser = MagicCircleParser()
    parser.symbols = symbols
    parser.connections = []
    
    # Build graph
    parser._build_symbol_graph()
    
    # Check connections
    print("Symbol Graph Connections:")
    for i, node in parser.symbol_graph.items():
        sym = node.symbol
        print(f"\n{i}: {sym.type.value} at {sym.position}")
        print(f"  Parent: {node.parent.symbol.type.value if node.parent else 'None'}")
        print(f"  Children: {[c.symbol.type.value for c in node.children]}")


if __name__ == "__main__":
    debug_connections()