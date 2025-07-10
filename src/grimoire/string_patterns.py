"""String pattern recognition for Grimoire symbols"""

import cv2
import numpy as np
from typing import Optional, Dict, List, Tuple
from dataclasses import dataclass


@dataclass
class StringPattern:
    """Represents a string pattern"""
    pattern_type: str
    text: str
    confidence: float = 0.0


class StringPatternRecognizer:
    """Recognizes string patterns from visual symbols"""
    
    def __init__(self):
        # Define common string patterns
        self.patterns = {
            # Line patterns to characters
            "vertical_line": "I",
            "horizontal_line": "-",
            "cross": "+",
            "diagonal_slash": "/",
            "diagonal_backslash": "\\",
            "circle": "O",
            "dot": ".",
            
            # Multiple lines to text
            "triple_line": "Hello, World!",  # Default for now
            "double_line": "Hi",
            "single_line": "A",
            
            # Special patterns
            "wavy_lines": "~",
            "zigzag": "W",
            "spiral": "@",
            "grid": "#",
        }
        
        # Common words based on pattern combinations
        self.word_patterns = {
            ("vertical_line", "horizontal_line"): "IT",
            ("circle", "vertical_line"): "OI",
            ("dot", "dot", "dot"): "...",
            ("horizontal_line", "horizontal_line"): "==",
        }
    
    def recognize_string_from_pattern(self, pattern_type: str, roi: np.ndarray = None) -> StringPattern:
        """Recognize string from pattern type"""
        # Direct pattern mapping
        if pattern_type in self.patterns:
            return StringPattern(
                pattern_type=pattern_type,
                text=self.patterns[pattern_type],
                confidence=0.9
            )
        
        # If ROI provided, try more sophisticated analysis
        if roi is not None:
            text = self._analyze_roi_for_text(roi)
            if text:
                return StringPattern(
                    pattern_type="custom",
                    text=text,
                    confidence=0.7
                )
        
        # Default fallback
        return StringPattern(
            pattern_type=pattern_type,
            text="Text",
            confidence=0.5
        )
    
    def _analyze_roi_for_text(self, roi: np.ndarray) -> Optional[str]:
        """Analyze ROI to extract text pattern"""
        if roi.size == 0:
            return None
        
        # Count strokes/lines
        edges = cv2.Canny(roi, 50, 150)
        lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=10, minLineLength=5, maxLineGap=3)
        
        if lines is None:
            return None
        
        num_lines = len(lines)
        
        # Map number of lines to possible characters
        line_to_char = {
            1: "I",
            2: "T",
            3: "E",
            4: "M",
            5: "W",
        }
        
        if num_lines in line_to_char:
            return line_to_char[num_lines]
        
        # Check for specific patterns
        if self._has_circular_pattern(roi):
            return "O"
        elif self._has_triangular_pattern(roi):
            return "A"
        
        return None
    
    def _has_circular_pattern(self, roi: np.ndarray) -> bool:
        """Check if ROI contains circular pattern"""
        # Find contours
        contours, _ = cv2.findContours(roi, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        for contour in contours:
            area = cv2.contourArea(contour)
            if area < 10:
                continue
            
            perimeter = cv2.arcLength(contour, True)
            if perimeter > 0:
                circularity = 4 * np.pi * area / (perimeter * perimeter)
                if circularity > 0.7:
                    return True
        
        return False
    
    def _has_triangular_pattern(self, roi: np.ndarray) -> bool:
        """Check if ROI contains triangular pattern"""
        contours, _ = cv2.findContours(roi, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        for contour in contours:
            # Approximate polygon
            epsilon = 0.04 * cv2.arcLength(contour, True)
            approx = cv2.approxPolyDP(contour, epsilon, True)
            
            if len(approx) == 3:
                return True
        
        return False
    
    def combine_patterns(self, patterns: List[str]) -> str:
        """Combine multiple patterns into a word or phrase"""
        # Check if pattern combination exists in word patterns
        pattern_tuple = tuple(patterns)
        if pattern_tuple in self.word_patterns:
            return self.word_patterns[pattern_tuple]
        
        # Otherwise concatenate individual characters
        result = ""
        for pattern in patterns:
            if pattern in self.patterns:
                result += self.patterns[pattern]
            else:
                result += "?"
        
        return result if result else "Text"