"""Advanced pattern recognition for Grimoire symbols"""

import cv2
import numpy as np
from typing import Tuple, Optional, List
from dataclasses import dataclass


@dataclass
class PatternInfo:
    """Information about detected pattern"""
    pattern_type: str
    value: Optional[int] = None
    confidence: float = 0.0


class AdvancedPatternDetector:
    """Advanced pattern detection within symbols"""
    
    def detect_dot_pattern(self, roi: np.ndarray) -> PatternInfo:
        """Detect dot patterns (1 dot = 1, 2 dots = 2, etc.)"""
        # Find connected components (dots)
        num_labels, labels, stats, centroids = cv2.connectedComponentsWithStats(roi, connectivity=8)
        
        # Filter out background and very small components
        min_area = 5
        dots = []
        for i in range(1, num_labels):  # Skip background (0)
            area = stats[i, cv2.CC_STAT_AREA]
            if area >= min_area:
                # Check if component is roughly circular
                width = stats[i, cv2.CC_STAT_WIDTH]
                height = stats[i, cv2.CC_STAT_HEIGHT]
                aspect_ratio = width / height if height > 0 else 0
                
                if 0.7 <= aspect_ratio <= 1.3:  # Roughly circular
                    dots.append({
                        'area': area,
                        'center': centroids[i],
                        'bbox': stats[i]
                    })
        
        num_dots = len(dots)
        
        if num_dots == 0:
            return PatternInfo("empty", None, 0.9)
        elif num_dots == 1:
            return PatternInfo("dot", 1, 0.9)
        elif num_dots == 2:
            return PatternInfo("double_dot", 2, 0.9)
        elif num_dots == 3:
            return PatternInfo("triple_dot", 3, 0.85)
        else:
            return PatternInfo("multiple_dots", num_dots, 0.8)
    
    def detect_line_pattern(self, roi: np.ndarray) -> PatternInfo:
        """Detect line patterns (horizontal, vertical, cross)"""
        # Use Hough Line Transform
        edges = cv2.Canny(roi, 50, 150)
        lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=20, minLineLength=10, maxLineGap=5)
        
        if lines is None:
            return PatternInfo("empty", None, 0.7)
        
        # Classify lines by angle
        horizontal_lines = 0
        vertical_lines = 0
        diagonal_lines = 0
        
        for line in lines:
            x1, y1, x2, y2 = line[0]
            angle = np.abs(np.arctan2(y2 - y1, x2 - x1) * 180 / np.pi)
            
            if angle < 15 or angle > 165:  # Horizontal
                horizontal_lines += 1
            elif 75 < angle < 105:  # Vertical
                vertical_lines += 1
            else:  # Diagonal
                diagonal_lines += 1
        
        total_lines = len(lines)
        
        if total_lines == 1:
            if horizontal_lines > 0:
                return PatternInfo("single_line", None, 0.85)
            elif vertical_lines > 0:
                return PatternInfo("vertical_line", None, 0.85)
            else:
                return PatternInfo("diagonal_line", None, 0.85)
        elif total_lines == 2:
            if horizontal_lines > 0 and vertical_lines > 0:
                return PatternInfo("cross", None, 0.9)
            else:
                return PatternInfo("double_line", None, 0.85)
        elif total_lines == 3:
            return PatternInfo("triple_line", None, 0.85)
        else:
            return PatternInfo("lines", None, 0.8)
    
    def detect_shape_pattern(self, roi: np.ndarray) -> PatternInfo:
        """Detect more complex patterns (half circle, grid, etc.)"""
        # Check for arc/half circle
        contours, _ = cv2.findContours(roi, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        if contours:
            largest_contour = max(contours, key=cv2.contourArea)
            
            # Fit ellipse to check for arc
            if len(largest_contour) >= 5:
                ellipse = cv2.fitEllipse(largest_contour)
                # Check if it's a partial ellipse (arc)
                hull = cv2.convexHull(largest_contour)
                hull_area = cv2.contourArea(hull)
                contour_area = cv2.contourArea(largest_contour)
                
                if contour_area > 0:
                    fill_ratio = contour_area / hull_area
                    if 0.4 < fill_ratio < 0.7:
                        return PatternInfo("half_circle", None, 0.8)
        
        # Check for grid pattern
        # Simple check: look for regular spacing of lines
        horizontal_spacing = self._check_grid_spacing(roi, axis=0)
        vertical_spacing = self._check_grid_spacing(roi, axis=1)
        
        if horizontal_spacing and vertical_spacing:
            return PatternInfo("grid", None, 0.85)
        
        return PatternInfo("unknown", None, 0.5)
    
    def _check_grid_spacing(self, roi: np.ndarray, axis: int) -> bool:
        """Check if there's regular spacing along an axis (for grid detection)"""
        # Project pixels along axis
        projection = np.sum(roi, axis=axis)
        
        # Find peaks (lines)
        threshold = np.max(projection) * 0.5
        peaks = []
        for i, val in enumerate(projection):
            if val > threshold:
                if not peaks or i - peaks[-1] > 5:  # Min spacing
                    peaks.append(i)
        
        if len(peaks) < 3:
            return False
        
        # Check if spacing is regular
        spacings = [peaks[i+1] - peaks[i] for i in range(len(peaks)-1)]
        if spacings:
            avg_spacing = np.mean(spacings)
            std_spacing = np.std(spacings)
            # Regular if standard deviation is small relative to mean
            return std_spacing < avg_spacing * 0.3
        
        return False
    
    def analyze_pattern(self, binary: np.ndarray, x: int, y: int, radius: float) -> PatternInfo:
        """Main pattern analysis function"""
        # Extract ROI - smaller to avoid borders
        roi_size = int(radius * 0.8)  # Much smaller to focus on interior only
        x1 = max(0, x - roi_size // 2)
        y1 = max(0, y - roi_size // 2)
        x2 = min(binary.shape[1], x + roi_size // 2)
        y2 = min(binary.shape[0], y + roi_size // 2)
        
        roi = binary[y1:y2, x1:x2].copy()
        
        if roi.size == 0:
            return PatternInfo("empty", None, 0.5)
        
        # Create circular mask to focus on center area only
        mask = np.zeros_like(roi)
        center = (roi.shape[1] // 2, roi.shape[0] // 2)
        mask_radius = min(roi.shape[0], roi.shape[1]) // 3  # Smaller mask
        cv2.circle(mask, center, mask_radius, 255, -1)
        
        # Apply mask to get interior only
        interior = cv2.bitwise_and(roi, mask)
        
        # Count white pixels
        white_pixels = cv2.countNonZero(interior)
        total_pixels = cv2.countNonZero(mask)
        
        if total_pixels == 0:
            return PatternInfo("empty", None, 0.5)
        
        fill_ratio = white_pixels / total_pixels
        
        # First check if mostly empty
        if fill_ratio < 0.05:
            return PatternInfo("empty", None, 0.9)
        
        # Try to detect dots
        dot_info = self.detect_dot_pattern(interior)
        if dot_info.pattern_type in ["dot", "double_dot", "triple_dot"]:
            return dot_info
        
        # Try to detect lines
        line_info = self.detect_line_pattern(interior)
        if line_info.confidence > 0.8:
            return line_info
        
        # Try to detect other shapes
        shape_info = self.detect_shape_pattern(interior)
        if shape_info.confidence > 0.7:
            return shape_info
        
        # Default based on fill ratio
        if fill_ratio > 0.7:
            return PatternInfo("filled", None, 0.7)
        else:
            return PatternInfo("pattern", None, 0.6)