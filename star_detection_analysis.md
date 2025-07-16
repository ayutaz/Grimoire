# Star Detection Analysis

## Problem Summary
The star symbol is not being detected in BOTH WASM and CLI for the hello-world.png image due to a distance filter.

## Key Findings

### 1. Star Detection Details
- **Star position**: (496, 449) - detected as `six_pointed_star` in debug mode
- **Star type conversion**: Lines 239-241 convert `SixPointedStar` to `Star` for output
- **Star area**: 709.50, Circularity: 0.06
- **Pattern detected**: "dot" pattern inside the star

### 2. Outer Circle Details
- **Outer circle position**: (300, 299)
- **Outer circle area**: 195310
- **Outer circle size (radius)**: √(195310) ≈ 441.9 pixels
- **Distance from star to center**: √((496-300)² + (449-299)²) ≈ 246.8 pixels

### 3. Star Filter Logic (detector.go lines 262-266)
```go
// For stars, only accept those near the center
if symbolType == Star {
    if centerDist < outerCircle.Size*0.3 { // Within 30% of radius from center
        symbols = append(symbols, symbol)
    }
}
```

### 4. Why Star is Filtered Out
- Filter threshold: 441.9 × 0.3 = 132.57 pixels
- Star distance: 246.8 pixels
- Since 246.8 > 132.57, the star is filtered out in BOTH CLI and WASM

### 5. Root Cause
The star at (496,449) is legitimately detected during contour analysis but is intentionally filtered out because it's too far from the center of the magic circle. This is by design to prevent false positive star detections outside the main program area.

### 6. Why the Confusion
The initial belief that CLI detects the star but WASM doesn't appears to be incorrect. Both implementations use the same detection code and both filter out the star. The debug output confirms:
- CLI output shows 13 symbols detected (no star)
- WASM would show the same 13 symbols

### 7. Potential Solutions
1. **Increase the star filter threshold** from 30% to 60% to allow stars further from center
2. **Remove the distance filter for stars** entirely (but this may cause false positives)
3. **Add a special case** for stars with specific patterns (like dot pattern)
4. **Consider this behavior correct** - stars far from center might not be part of the main program