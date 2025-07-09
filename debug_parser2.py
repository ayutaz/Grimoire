#!/usr/bin/env python3
"""Debug parser statement parsing"""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from grimoire.image_recognition import Symbol, SymbolType, Connection
from grimoire.parser import MagicCircleParser


def debug_parsing():
    """Debug statement parsing"""
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
    
    # Debug parsing step by step
    print("=== PARSING DEBUG ===")
    
    # Parse and show what happens
    ast = parser.parse(symbols, [])
    
    print(f"\nMain entry body statements: {len(ast.main_entry.body if ast.main_entry else 0)}")
    if ast.main_entry and ast.main_entry.body:
        for i, stmt in enumerate(ast.main_entry.body):
            print(f"  Statement {i}: {type(stmt).__name__}")
    
    print(f"\nVisited status after parsing:")
    for i, node in parser.symbol_graph.items():
        print(f"  {i}: {node.symbol.type.value} - visited: {node.visited}")


if __name__ == "__main__":
    debug_parsing()