#!/usr/bin/env python3
"""Test script for Grimoire compiler"""

import sys
import os
from pathlib import Path

# Add src to path so we can import grimoire
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from grimoire.compiler import GrimoireCompiler
import generate_sample_images
from generate_sample_images import MagicCircleDrawer, SampleImageGenerator


def create_test_image():
    """Create a simple test magic circle"""
    print("Creating test magic circle image...")
    
    generator = SampleImageGenerator()
    drawer = generator.drawer
    
    # Create a simple program: 1 + 2 with output
    drawer.setup(600, 600)
    
    # Draw outer circle (mandatory)
    drawer.draw_outer_circle()
    
    # Draw main entry point
    main_pos = (300, 200)
    drawer.draw_double_circle(*main_pos, label="")
    
    # Draw numbers
    one_pos = (200, 300)
    drawer.draw_square(*one_pos, "•")  # 1
    
    two_pos = (400, 300)
    drawer.draw_square(*two_pos, "••")  # 2
    
    # Draw addition operator
    add_pos = (300, 300)
    drawer.draw_convergence(*add_pos)
    
    # Draw output
    output_pos = (300, 400)
    drawer.draw_star(*output_pos)
    
    # Draw connections
    drawer.draw_connection(*main_pos, *one_pos)
    drawer.draw_connection(*one_pos, *add_pos)
    drawer.draw_connection(*two_pos, *add_pos)
    drawer.draw_connection(*add_pos, *output_pos)
    
    # Save image
    test_path = "test_addition.png"
    drawer.save(test_path)
    print(f"Test image saved to: {test_path}")
    return test_path


def test_compiler():
    """Test the compiler with a simple magic circle"""
    # Create test image
    image_path = create_test_image()
    
    print("\n" + "="*50)
    print("Testing Grimoire Compiler")
    print("="*50 + "\n")
    
    # Create compiler
    compiler = GrimoireCompiler()
    
    try:
        # Test 1: Debug mode to see what's happening
        print("1. Running in debug mode...")
        print("-" * 30)
        ast, result = compiler.debug(image_path)
        print(f"Debug output: {result}")
        
        # Test 2: Compile and run
        print("\n2. Compiling and running...")
        print("-" * 30)
        output = compiler.compile_and_run(image_path)
        print(f"Program output: {output}")
        
        # Test 3: Generate Python code
        print("\n3. Generating Python code...")
        print("-" * 30)
        python_code = compiler.compile_to_python(image_path)
        print("Generated Python code:")
        print(python_code)
        
        # Test 4: Save generated code
        output_file = "test_addition.py"
        compiler.compile_to_python(image_path, output_file)
        print(f"\nGenerated code saved to: {output_file}")
        
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()
    
    # Clean up
    if os.path.exists(image_path):
        os.remove(image_path)
        print(f"\nCleaned up test image: {image_path}")


if __name__ == "__main__":
    test_compiler()