#!/usr/bin/env python3
"""Simple test script for Grimoire compiler"""

import sys
import os
from pathlib import Path

# Add src to path so we can import grimoire
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from grimoire.compiler import GrimoireCompiler


def test_compiler():
    """Test the compiler with pre-created magic circle images"""
    
    # Test images
    test_images = [
        "test_simple_addition.png",
        "test_loop.png",
        "test_conditional.png"
    ]
    
    compiler = GrimoireCompiler()
    
    for image_path in test_images:
        if not os.path.exists(image_path):
            print(f"Skipping {image_path} - file not found")
            continue
            
        print(f"\n{'='*60}")
        print(f"Testing: {image_path}")
        print('='*60)
        
        try:
            # Try to compile and run
            print("\nAttempting to compile and run...")
            result = compiler.compile_and_run(image_path)
            print(f"Result: {result}")
            
        except Exception as e:
            print(f"Error: {e}")
            
            # Try debug mode for more info
            print("\nTrying debug mode...")
            try:
                ast, debug_result = compiler.debug(image_path)
                print(f"Debug result: {debug_result}")
            except Exception as e2:
                print(f"Debug error: {e2}")
                import traceback
                traceback.print_exc()


if __name__ == "__main__":
    test_compiler()