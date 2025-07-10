#!/usr/bin/env python3
"""Test hello world generation"""

from grimoire.compiler import compile_magic_circle

# Test with a simple hello world program
result = compile_magic_circle("samples/hello_world.png", debug=True)
print("Generated code:")
print(result)