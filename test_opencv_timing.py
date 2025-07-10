#!/usr/bin/env python3
import time
import sys

print("Starting timing test...")
start = time.time()

# Time import
import_start = time.time()
import cv2
import numpy as np
import_end = time.time()
print(f"Import time: {import_end - import_start:.3f}s")

# Time image reading
if len(sys.argv) > 1:
    read_start = time.time()
    img = cv2.imread(sys.argv[1])
    read_end = time.time()
    print(f"Image read time: {read_end - read_start:.3f}s")
    print(f"Image shape: {img.shape if img is not None else 'None'}")

# Time basic operations
op_start = time.time()
gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY) if img is not None else None
binary = cv2.adaptiveThreshold(gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY_INV, 11, 2) if gray is not None else None
op_end = time.time()
print(f"Basic operations time: {op_end - op_start:.3f}s")

total_time = time.time() - start
print(f"Total time: {total_time:.3f}s")