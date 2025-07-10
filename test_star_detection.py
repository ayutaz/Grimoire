#!/usr/bin/env python3
"""Test star detection specifically"""

import cv2
import numpy as np
from grimoire.image_recognition import MagicCircleDetector, Symbol, SymbolType

# Create a custom detector to test
detector = MagicCircleDetector()

# Load and preprocess
img = cv2.imread("examples/images/hello_world.png")
gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
binary = detector._preprocess_image(gray)

# First detect outer circle
outer_circle = detector._detect_outer_circle(binary)
print(f"Outer circle: {outer_circle}")

# Now manually check star detection
print("\n=== Testing star detection ===")

# Use RETR_TREE to get all contours including nested ones
contours, hierarchy = cv2.findContours(binary, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
print(f"Total contours found: {len(contours)}")

# Manually check each contour
for i, contour in enumerate(contours):
    area = cv2.contourArea(contour)
    if area < 100:  # Skip tiny contours
        continue
        
    M = cv2.moments(contour)
    if M["m00"] != 0:
        cx = int(M["m10"] / M["m00"])
        cy = int(M["m01"] / M["m00"])
        
        # Check if it could be a star
        epsilon = 0.02 * cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, epsilon, True)
        
        print(f"\nContour {i}:")
        print(f"  Area: {area}")
        print(f"  Center: ({cx}, {cy})")
        print(f"  Vertices: {len(approx)}")
        print(f"  Hierarchy: {hierarchy[0][i] if hierarchy is not None else 'None'}")
        
        # Check if it's inside outer circle and right size
        if outer_circle:
            dist = np.sqrt((cx - outer_circle.position[0])**2 + (cy - outer_circle.position[1])**2)
            print(f"  Distance from outer circle center: {dist}")
            print(f"  Inside outer circle: {dist < outer_circle.size * 0.8}")
            print(f"  Is star shape: {detector._is_star_shape(contour, cx, cy)}")

# Now test the full detection with modified approach
print("\n=== Running full detection ===")
detector.symbols = []  # Reset
detector.connections = []

# Detect all symbols
symbols, connections = detector.detect_symbols("examples/images/hello_world.png")
print(f"\nDetected symbols: {len(symbols)}")
for sym in symbols:
    print(f"  {sym.type.value} at {sym.position}")