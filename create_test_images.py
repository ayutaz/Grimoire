#!/usr/bin/env python3
"""Create test images for the Grimoire compiler"""

import cv2
import numpy as np
import os


def create_simple_addition():
    """Create a simple addition magic circle"""
    # Create white background
    img = np.ones((600, 600, 3), dtype=np.uint8) * 255
    
    # Draw outer circle (black)
    cv2.circle(img, (300, 300), 280, (0, 0, 0), 3)
    
    # Draw main entry (double circle)
    cv2.circle(img, (300, 150), 30, (0, 0, 0), 2)
    cv2.circle(img, (300, 150), 25, (0, 0, 0), 2)
    
    # Draw number 1 (square with dot)
    cv2.rectangle(img, (170, 220), (230, 280), (0, 0, 0), 2)
    cv2.circle(img, (200, 250), 3, (0, 0, 0), -1)
    
    # Draw number 2 (square with two dots)
    cv2.rectangle(img, (370, 220), (430, 280), (0, 0, 0), 2)
    cv2.circle(img, (390, 250), 3, (0, 0, 0), -1)
    cv2.circle(img, (410, 250), 3, (0, 0, 0), -1)
    
    # Draw convergence operator (lines meeting)
    cv2.line(img, (230, 250), (270, 250), (0, 0, 0), 2)
    cv2.line(img, (330, 250), (370, 250), (0, 0, 0), 2)
    cv2.circle(img, (300, 250), 5, (0, 0, 0), -1)
    
    # Draw output star
    # Simple 5-pointed star
    pts = np.array([
        [300, 330],
        [310, 360],
        [340, 360],
        [315, 380],
        [325, 410],
        [300, 390],
        [275, 410],
        [285, 380],
        [260, 360],
        [290, 360]
    ], np.int32)
    pts = pts.reshape((-1, 1, 2))
    cv2.polylines(img, [pts], True, (0, 0, 0), 2)
    
    # Draw connections
    cv2.line(img, (300, 180), (200, 220), (0, 0, 0), 1)
    cv2.line(img, (300, 255), (300, 330), (0, 0, 0), 1)
    
    cv2.imwrite("test_simple_addition.png", img)
    print("Created: test_simple_addition.png")


def create_loop_example():
    """Create a loop magic circle"""
    # Create white background
    img = np.ones((600, 600, 3), dtype=np.uint8) * 255
    
    # Draw outer circle
    cv2.circle(img, (300, 300), 280, (0, 0, 0), 3)
    
    # Draw main entry (double circle)
    cv2.circle(img, (300, 100), 30, (0, 0, 0), 2)
    cv2.circle(img, (300, 100), 25, (0, 0, 0), 2)
    
    # Draw counter (square with 3 dots)
    cv2.rectangle(img, (220, 170), (280, 230), (0, 0, 0), 2)
    cv2.circle(img, (235, 200), 3, (0, 0, 0), -1)
    cv2.circle(img, (250, 200), 3, (0, 0, 0), -1)
    cv2.circle(img, (265, 200), 3, (0, 0, 0), -1)
    
    # Draw pentagon (loop)
    pentagon = []
    for i in range(5):
        angle = i * 72 * np.pi / 180 - np.pi / 2
        x = int(300 + 60 * np.cos(angle))
        y = int(300 + 60 * np.sin(angle))
        pentagon.append([x, y])
    pentagon = np.array(pentagon, np.int32)
    cv2.polylines(img, [pentagon], True, (0, 0, 0), 2)
    
    # Draw output inside loop
    cv2.polylines(img, [np.array([
        [300, 280],
        [310, 300],
        [330, 300],
        [315, 315],
        [320, 335],
        [300, 320],
        [280, 335],
        [285, 315],
        [270, 300],
        [290, 300]
    ], np.int32)], True, (0, 0, 0), 2)
    
    # Connections
    cv2.line(img, (300, 130), (250, 170), (0, 0, 0), 1)
    cv2.line(img, (250, 230), (300, 240), (0, 0, 0), 1)
    
    cv2.imwrite("test_loop.png", img)
    print("Created: test_loop.png")


def create_conditional_example():
    """Create a conditional (if-else) magic circle"""
    img = np.ones((600, 600, 3), dtype=np.uint8) * 255
    
    # Draw outer circle
    cv2.circle(img, (300, 300), 280, (0, 0, 0), 3)
    
    # Draw main entry
    cv2.circle(img, (300, 100), 30, (0, 0, 0), 2)
    cv2.circle(img, (300, 100), 25, (0, 0, 0), 2)
    
    # Draw triangle (conditional)
    triangle = np.array([[300, 200], [250, 280], [350, 280]], np.int32)
    cv2.polylines(img, [triangle], True, (0, 0, 0), 2)
    
    # Draw condition (comparison)
    cv2.rectangle(img, (270, 150), (330, 190), (0, 0, 0), 2)
    cv2.line(img, (290, 170), (310, 170), (0, 0, 0), 2)
    cv2.polylines(img, [np.array([[310, 165], [320, 170], [310, 175]], np.int32)], False, (0, 0, 0), 2)
    
    # Draw left branch (then)
    cv2.polylines(img, [np.array([
        [200, 320],
        [210, 340],
        [230, 340],
        [215, 355],
        [220, 375],
        [200, 360],
        [180, 375],
        [185, 355],
        [170, 340],
        [190, 340]
    ], np.int32)], True, (0, 0, 0), 2)
    
    # Draw right branch (else)
    cv2.polylines(img, [np.array([
        [400, 320],
        [410, 340],
        [430, 340],
        [415, 355],
        [420, 375],
        [400, 360],
        [380, 375],
        [385, 355],
        [370, 340],
        [390, 340]
    ], np.int32)], True, (0, 0, 0), 2)
    
    # Connections
    cv2.line(img, (300, 130), (300, 150), (0, 0, 0), 1)
    cv2.line(img, (300, 190), (300, 200), (0, 0, 0), 1)
    cv2.line(img, (250, 280), (200, 320), (0, 0, 0), 1)
    cv2.line(img, (350, 280), (400, 320), (0, 0, 0), 1)
    
    cv2.imwrite("test_conditional.png", img)
    print("Created: test_conditional.png")


if __name__ == "__main__":
    print("Creating test images for Grimoire compiler...")
    create_simple_addition()
    create_loop_example()
    create_conditional_example()
    print("\nAll test images created!")