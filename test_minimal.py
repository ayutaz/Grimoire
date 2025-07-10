#!/usr/bin/env python3
import time
print(f"Start: {time.time()}")

# Test 1: Just imports
print("Before cv2 import...")
import cv2
print(f"After cv2 import: {time.time()}")

print("Before numpy import...")
import numpy as np  
print(f"After numpy import: {time.time()}")

print("Done!")