#!/usr/bin/env python3
"""Debug image detection"""

import cv2
import numpy as np
from grimoire.image_recognition import MagicCircleDetector

# Load and preprocess image
img = cv2.imread("examples/images/hello_world.png")
gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

# Apply preprocessing
detector = MagicCircleDetector()
binary = detector._preprocess_image(gray)

# Save preprocessed image
cv2.imwrite("debug_binary.png", binary)
print("Saved preprocessed image to debug_binary.png")

# Find all contours
contours, hierarchy = cv2.findContours(binary, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
print(f"\nFound {len(contours)} contours")

# Analyze each contour
for i, contour in enumerate(contours):
    area = cv2.contourArea(contour)
    if area < 50:  # Skip very small contours
        continue
        
    # Get contour properties
    perimeter = cv2.arcLength(contour, True)
    approx = cv2.approxPolyDP(contour, 0.04 * perimeter, True)
    
    # Get center
    M = cv2.moments(contour)
    if M["m00"] != 0:
        cx = int(M["m10"] / M["m00"])
        cy = int(M["m01"] / M["m00"])
        
        print(f"\nContour {i}:")
        print(f"  Area: {area}")
        print(f"  Vertices: {len(approx)}")
        print(f"  Center: ({cx}, {cy})")
        
        # Check circularity
        if area > 0:
            circularity = 4 * np.pi * area / (perimeter * perimeter)
            print(f"  Circularity: {circularity:.2f}")
        
        # Draw contour on debug image
        debug_img = img.copy()
        cv2.drawContours(debug_img, [contour], -1, (0, 255, 0), 2)
        cv2.circle(debug_img, (cx, cy), 5, (255, 0, 0), -1)
        cv2.imwrite(f"debug_contour_{i}.png", debug_img)

# Try to detect star manually
print("\n=== Manual Star Detection ===")
# The star is likely in the center of the image
h, w = gray.shape
center_x, center_y = w // 2, h // 2

# Create ROI around center
roi_size = 200
x1 = max(0, center_x - roi_size // 2)
y1 = max(0, center_y - roi_size // 2)
x2 = min(w, center_x + roi_size // 2)
y2 = min(h, center_y + roi_size // 2)

roi = binary[y1:y2, x1:x2]
cv2.imwrite("debug_roi.png", roi)

# Count white pixels in ROI
white_pixels = cv2.countNonZero(roi)
print(f"White pixels in center ROI: {white_pixels}")
print(f"ROI size: {roi.shape}")

# Run full detection
symbols, connections = detector.detect_symbols("examples/images/hello_world.png")
print(f"\n=== Detection Results ===")
print(f"Symbols detected: {len(symbols)}")
for sym in symbols:
    print(f"  {sym.type.value} at {sym.position}")