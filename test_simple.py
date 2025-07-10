#!/usr/bin/env python3
"""Test with a simple programmatic approach"""

from grimoire.image_recognition import Symbol, SymbolType, Connection
from grimoire.parser import MagicCircleParser
from grimoire.interpreter import GrimoireInterpreter
from grimoire.code_generator import PythonCodeGenerator

# Create a simple hello world program programmatically
symbols = [
    Symbol(
        type=SymbolType.OUTER_CIRCLE,
        position=(300, 300),
        size=300,
        confidence=1.0,
        properties={'is_double': False}
    ),
    Symbol(
        type=SymbolType.SQUARE,  # Literal "Hello, World!"
        position=(300, 200),
        size=50,
        confidence=1.0,
        properties={'pattern': 'triple_line'}  # String literal
    ),
    Symbol(
        type=SymbolType.STAR,  # Output
        position=(300, 400),
        size=50,
        confidence=1.0,
        properties={'points': 5}
    )
]

# Create a connection from literal to output
connections = [
    Connection(
        from_symbol=symbols[1],  # Square (literal)
        to_symbol=symbols[2],    # Star (output)
        connection_type="solid"
    )
]

# Parse
parser = MagicCircleParser()
try:
    ast = parser.parse(symbols, connections)
    print("=== AST Parsed Successfully ===")
    
    # Interpret
    interpreter = GrimoireInterpreter()
    result = interpreter.interpret(ast)
    print(f"Output: {result}")
    
    # Generate Python code
    generator = PythonCodeGenerator()
    python_code = generator.generate(ast)
    print("\n=== Generated Python Code ===")
    print(python_code)
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()