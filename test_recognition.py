#!/usr/bin/env python3
"""Test image recognition directly"""

from grimoire.image_recognition import MagicCircleDetector

def test_image(image_path):
    print(f"\n=== Testing {image_path} ===")
    detector = MagicCircleDetector()
    try:
        symbols, connections = detector.detect_symbols(image_path)
        print(f"Detected {len(symbols)} symbols and {len(connections)} connections")
        for symbol in symbols:
            print(f"  - {symbol.type.value} at {symbol.position}")
            if hasattr(symbol, 'properties') and symbol.properties:
                print(f"    properties: {symbol.properties}")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_image("examples/images/hello_world.png")
    test_image("examples/images/calculator.png")