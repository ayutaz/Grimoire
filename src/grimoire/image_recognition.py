"""Image recognition module for Grimoire - Detects symbols from magic circle images"""

import cv2
import numpy as np
from typing import List, Tuple, Dict, Optional, Any
from dataclasses import dataclass
from enum import Enum
import math


class SymbolType(Enum):
    """Types of symbols that can be detected"""
    # Structure elements
    OUTER_CIRCLE = "outer_circle"
    CIRCLE = "circle"
    DOUBLE_CIRCLE = "double_circle"
    SQUARE = "square"
    TRIANGLE = "triangle"
    PENTAGON = "pentagon"
    HEXAGON = "hexagon"
    STAR = "star"
    
    # Operators
    CONVERGENCE = "convergence"  # ⟐
    DIVERGENCE = "divergence"    # ⟑
    AMPLIFICATION = "amplification"  # ✦
    DISTRIBUTION = "distribution"    # ⟠
    
    # Data types (patterns inside shapes)
    DOT = "dot"
    DOUBLE_DOT = "double_dot"
    LINES = "lines"
    HALF_CIRCLE = "half_circle"
    
    # Special symbols
    CONNECTION = "connection"
    ARROW = "arrow"
    LOOP_BACK = "loop_back"


@dataclass
class Symbol:
    """Represents a detected symbol"""
    type: SymbolType
    position: Tuple[int, int]  # (x, y)
    size: float
    confidence: float
    properties: Dict[str, Any]  # Additional properties like pattern, connections


@dataclass
class Connection:
    """Represents a connection between symbols"""
    from_symbol: Symbol
    to_symbol: Symbol
    connection_type: str  # "solid", "dashed", "curved"


class MagicCircleDetector:
    """Detects and analyzes magic circles from images"""
    
    def __init__(self):
        self.min_contour_area = 100
        self.circle_threshold = 0.88  # Even stricter threshold to reject ellipses
        self.symbols: List[Symbol] = []
        self.connections: List[Connection] = []
    
    def detect_symbols(self, image_path: str) -> Tuple[List[Symbol], List[Connection]]:
        """Main detection function"""
        # Load image
        img = cv2.imread(image_path)
        if img is None:
            raise ValueError(f"Cannot load image: {image_path}")
        
        # Preprocess
        gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
        binary = self._preprocess_image(gray)
        
        # Detect outer circle first (required)
        outer_circle = self._detect_outer_circle(binary)
        if outer_circle is None:
            raise ValueError("No outer circle detected. All Grimoire programs must be enclosed in a magic circle.")
        
        self.symbols = [outer_circle]
        
        # Detect other symbols within the outer circle
        self._detect_circles(binary, outer_circle)
        self._detect_polygons(binary, outer_circle)
        self._detect_stars(binary, outer_circle)
        self._detect_operators(binary, outer_circle)
        self._detect_connections(binary)
        
        return self.symbols, self.connections
    
    def _preprocess_image(self, gray: np.ndarray) -> np.ndarray:
        """Preprocess image for better detection"""
        # Apply Gaussian blur to reduce noise
        blurred = cv2.GaussianBlur(gray, (5, 5), 0)
        
        # Apply adaptive threshold
        binary = cv2.adaptiveThreshold(
            blurred, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
            cv2.THRESH_BINARY_INV, 11, 2
        )
        
        # Morphological operations to clean up
        kernel = np.ones((3, 3), np.uint8)
        binary = cv2.morphologyEx(binary, cv2.MORPH_CLOSE, kernel)
        binary = cv2.morphologyEx(binary, cv2.MORPH_OPEN, kernel)
        
        return binary
    
    def _detect_outer_circle(self, binary: np.ndarray) -> Optional[Symbol]:
        """Detect the outer circle (magic circle boundary)"""
        # Find contours
        contours, _ = cv2.findContours(binary, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        # Find the largest circular contour
        for contour in sorted(contours, key=cv2.contourArea, reverse=True):
            area = cv2.contourArea(contour)
            if area < self.min_contour_area:
                continue
            
            # Check if it's circular
            perimeter = cv2.arcLength(contour, True)
            circularity = 4 * np.pi * area / (perimeter * perimeter)
            
            if circularity > self.circle_threshold:
                # Get circle parameters
                (x, y), radius = cv2.minEnclosingCircle(contour)
                
                # Check if it's near the image edges (outer circle)
                h, w = binary.shape
                margin = min(w, h) * 0.2  # Increased margin tolerance
                # Also check if it's the largest circle found
                if radius > min(w, h) * 0.3:  # At least 30% of image size
                    
                    return Symbol(
                        type=SymbolType.OUTER_CIRCLE,
                        position=(int(x), int(y)),
                        size=radius,
                        confidence=circularity,
                        properties={"is_double": self._is_double_circle(binary, x, y, radius)}
                    )
        
        return None
    
    def _is_double_circle(self, binary: np.ndarray, x: float, y: float, radius: float) -> bool:
        """Check if a circle is actually a double circle"""
        # Look for concentric circles by checking different radii
        inner_found = False
        outer_found = False
        
        # Check for circles at different radii
        for r_factor in [0.7, 0.8, 0.9, 1.0, 1.1]:
            test_radius = int(radius * r_factor)
            # Create a circular mask at this radius
            mask = np.zeros_like(binary)
            cv2.circle(mask, (int(x), int(y)), test_radius, 255, 2)
            
            # Count white pixels on the circle perimeter
            masked = cv2.bitwise_and(binary, mask)
            white_pixels = cv2.countNonZero(masked)
            expected_pixels = 2 * np.pi * test_radius
            
            # If more than 60% of the circle perimeter has white pixels
            if white_pixels > expected_pixels * 0.6:
                if r_factor < 0.9:
                    inner_found = True
                else:
                    outer_found = True
        
        return inner_found and outer_found
    
    def _detect_circles(self, binary: np.ndarray, outer_circle: Symbol):
        """Detect circles within the outer circle"""
        # Use HoughCircles for better circle detection
        circles = cv2.HoughCircles(
            binary, cv2.HOUGH_GRADIENT, 1, 20,
            param1=50, param2=20, minRadius=10, maxRadius=int(outer_circle.size * 0.5)
        )
        
        if circles is not None:
            circles = np.uint16(np.around(circles))
            for circle in circles[0]:
                x, y, radius = circle
                
                # Check if inside outer circle
                try:
                    dx = float(x) - float(outer_circle.position[0])
                    dy = float(y) - float(outer_circle.position[1])
                    dist = np.sqrt(dx**2 + dy**2)
                except (OverflowError, ValueError):
                    continue
                if dist + radius < outer_circle.size:
                    # Check if it's a double circle
                    is_double = self._is_double_circle(binary, x, y, radius)
                    
                    symbol = Symbol(
                        type=SymbolType.DOUBLE_CIRCLE if is_double else SymbolType.CIRCLE,
                        position=(x, y),
                        size=radius,
                        confidence=0.9,
                        properties={"pattern": self._detect_internal_pattern(binary, x, y, radius)}
                    )
                    self.symbols.append(symbol)
    
    def _detect_polygons(self, binary: np.ndarray, outer_circle: Symbol):
        """Detect polygons (triangles, squares, pentagons, hexagons)"""
        # Find contours
        contours, _ = cv2.findContours(binary, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
        
        for contour in contours:
            area = cv2.contourArea(contour)
            if area < self.min_contour_area:
                continue
            
            # Approximate polygon
            epsilon = 0.04 * cv2.arcLength(contour, True)
            approx = cv2.approxPolyDP(contour, epsilon, True)
            
            # Get center
            M = cv2.moments(contour)
            if M["m00"] != 0:
                cx = int(M["m10"] / M["m00"])
                cy = int(M["m01"] / M["m00"])
                
                # Check if inside outer circle
                dist = np.sqrt((cx - outer_circle.position[0])**2 + (cy - outer_circle.position[1])**2)
                if dist < outer_circle.size * 0.8:
                    # Identify shape by number of vertices
                    vertices = len(approx)
                    symbol_type = None
                    
                    if vertices == 3:
                        symbol_type = SymbolType.TRIANGLE
                    elif vertices == 4:
                        symbol_type = SymbolType.SQUARE
                    elif vertices == 5:
                        symbol_type = SymbolType.PENTAGON
                    elif vertices == 6:
                        symbol_type = SymbolType.HEXAGON
                    
                    if symbol_type:
                        # Calculate size
                        _, (width, height), _ = cv2.minAreaRect(contour)
                        size = max(width, height)
                        
                        symbol = Symbol(
                            type=symbol_type,
                            position=(cx, cy),
                            size=size,
                            confidence=0.8,
                            properties={"vertices": vertices, "pattern": self._detect_internal_pattern(binary, cx, cy, size/2)}
                        )
                        self.symbols.append(symbol)
    
    def _detect_stars(self, binary: np.ndarray, outer_circle: Symbol):
        """Detect star shapes"""
        # Stars are more complex - look for shapes with alternating distances from center
        contours, _ = cv2.findContours(binary, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
        
        for contour in contours:
            area = cv2.contourArea(contour)
            if area < self.min_contour_area * 2:  # Stars need larger area
                continue
            
            # Get center
            M = cv2.moments(contour)
            if M["m00"] != 0:
                cx = int(M["m10"] / M["m00"])
                cy = int(M["m01"] / M["m00"])
                
                # Check if inside outer circle
                dist = np.sqrt((cx - outer_circle.position[0])**2 + (cy - outer_circle.position[1])**2)
                if dist < outer_circle.size * 0.8:
                    # Check for star pattern (simplified)
                    if self._is_star_shape(contour, cx, cy):
                        _, (width, height), _ = cv2.minAreaRect(contour)
                        size = max(width, height)
                        
                        symbol = Symbol(
                            type=SymbolType.STAR,
                            position=(cx, cy),
                            size=size,
                            confidence=0.7,
                            properties={"points": 5}
                        )
                        self.symbols.append(symbol)
    
    def _is_star_shape(self, contour, cx: int, cy: int) -> bool:
        """Check if contour is star-shaped (simplified)"""
        # Calculate distances from center to contour points
        distances = []
        for point in contour:
            x, y = point[0]
            dist = np.sqrt((x - cx)**2 + (y - cy)**2)
            distances.append(dist)
        
        if len(distances) < 10:
            return False
        
        # Check for alternating pattern of distances
        # (simplified star detection)
        distances = np.array(distances)
        mean_dist = np.mean(distances)
        variations = np.abs(distances - mean_dist)
        
        # Star should have significant variations
        return np.std(variations) > mean_dist * 0.2
    
    def _detect_operators(self, binary: np.ndarray, outer_circle: Symbol):
        """Detect operator symbols (convergence, divergence, etc.)"""
        # This is simplified - in reality would need more sophisticated pattern matching
        # For now, detect based on specific patterns of lines
        
        # Example: Detect convergence pattern (lines meeting at a point)
        # This would require line detection and intersection analysis
        pass
    
    def _detect_connections(self, binary: np.ndarray):
        """Detect connections between symbols"""
        # Use Hough Line Transform to detect lines
        edges = cv2.Canny(binary, 30, 100)
        lines = cv2.HoughLinesP(edges, 1, np.pi/180, 30, minLineLength=20, maxLineGap=15)
        
        if lines is not None:
            for line in lines:
                x1, y1, x2, y2 = line[0]
                
                # Find symbols near line endpoints
                start_symbol = self._find_nearest_symbol(x1, y1)
                end_symbol = self._find_nearest_symbol(x2, y2)
                
                if start_symbol and end_symbol and start_symbol != end_symbol:
                    connection = Connection(
                        from_symbol=start_symbol,
                        to_symbol=end_symbol,
                        connection_type="solid"
                    )
                    self.connections.append(connection)
    
    def _find_nearest_symbol(self, x: int, y: int, max_dist: float = 30) -> Optional[Symbol]:
        """Find the nearest symbol to a point"""
        nearest = None
        min_dist = max_dist
        
        for symbol in self.symbols:
            dist = np.sqrt((x - symbol.position[0])**2 + (y - symbol.position[1])**2)
            if dist < min_dist:
                min_dist = dist
                nearest = symbol
        
        return nearest
    
    def _detect_internal_pattern(self, binary: np.ndarray, x: int, y: int, radius: float) -> Optional[str]:
        """Detect pattern inside a shape (for data types)"""
        # Create ROI around the shape
        roi_size = int(radius * 2)
        x1 = max(0, x - roi_size // 2)
        y1 = max(0, y - roi_size // 2)
        x2 = min(binary.shape[1], x + roi_size // 2)
        y2 = min(binary.shape[0], y + roi_size // 2)
        
        roi = binary[y1:y2, x1:x2]
        
        # Count non-zero pixels (simplified pattern detection)
        white_pixels = cv2.countNonZero(roi)
        total_pixels = roi.shape[0] * roi.shape[1]
        
        if total_pixels == 0:
            return None
        
        fill_ratio = white_pixels / total_pixels
        
        # Simple classification based on fill ratio
        if fill_ratio < 0.1:
            return "empty"
        elif fill_ratio < 0.2:
            return "dot"  # Single dot
        elif fill_ratio < 0.3:
            return "double_dot"  # Double dot
        elif fill_ratio < 0.5:
            return "lines"  # Lines pattern
        else:
            return "filled"
        
        return None