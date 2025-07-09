#!/usr/bin/env python3
"""Direct test of the parser with manually created symbols"""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from grimoire.image_recognition import Symbol, SymbolType, Connection
from grimoire.parser import MagicCircleParser
from grimoire.interpreter import GrimoireInterpreter
from grimoire.code_generator import PythonCodeGenerator


def create_addition_symbols():
    """Create symbols for 1 + 2 = 3"""
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
    
    # No connections for now - parser should handle this
    connections = []
    
    return symbols, connections


def test_parser():
    """Test the parser directly"""
    print("="*60)
    print("Testing Grimoire Parser Directly")
    print("="*60)
    
    # Create test symbols
    symbols, connections = create_addition_symbols()
    
    print(f"\nCreated {len(symbols)} symbols:")
    for sym in symbols:
        print(f"  - {sym.type.value} at {sym.position}")
    
    # Parse
    parser = MagicCircleParser()
    try:
        print("\nParsing symbols to AST...")
        ast = parser.parse(symbols, connections)
        
        print(f"\nAST created successfully:")
        print(f"  - Has outer circle: {ast.has_outer_circle}")
        print(f"  - Main entry: {ast.main_entry is not None}")
        print(f"  - Functions: {len(ast.functions)}")
        print(f"  - Global statements: {len(ast.globals)}")
        
        # Interpret
        print("\nInterpreting AST...")
        interpreter = GrimoireInterpreter()
        result = interpreter.interpret(ast)
        print(f"Result: {result}")
        
        # Generate code
        print("\nGenerating Python code...")
        generator = PythonCodeGenerator()
        python_code = generator.generate(ast)
        print("Generated code:")
        print("-" * 40)
        print(python_code)
        print("-" * 40)
        
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()


def create_loop_symbols():
    """Create symbols for a simple loop"""
    symbols = [
        # Outer circle
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
        # Loop counter (3 times)
        Symbol(
            type=SymbolType.SQUARE,
            position=(250, 200),
            size=40,
            confidence=0.9,
            properties={"pattern": "triple_dot"}  # 3
        ),
        # Loop pentagon
        Symbol(
            type=SymbolType.PENTAGON,
            position=(300, 250),
            size=60,
            confidence=0.9,
            properties={}
        ),
        # Output inside loop
        Symbol(
            type=SymbolType.STAR,
            position=(300, 250),  # Inside the loop
            size=30,
            confidence=0.9,
            properties={}
        )
    ]
    
    connections = []
    return symbols, connections


def test_loop():
    """Test loop parsing"""
    print("\n" + "="*60)
    print("Testing Loop Parser")
    print("="*60)
    
    symbols, connections = create_loop_symbols()
    
    parser = MagicCircleParser()
    try:
        ast = parser.parse(symbols, connections)
        
        interpreter = GrimoireInterpreter()
        result = interpreter.interpret(ast)
        print(f"Loop result: {result}")
        
    except Exception as e:
        print(f"Loop error: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    test_parser()
    test_loop()