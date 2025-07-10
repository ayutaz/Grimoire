"""Image recognition module for Grimoire - Detects symbols from magic circle images"""

import cv2
import numpy as np
from typing import List, Tuple, Dict, Optional, Any
from dataclasses import dataclass
from enum import Enum
import math
from .advanced_recognition import AdvancedPatternDetector, PatternInfo


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
        self.circle_threshold = 0.80  # Threshold for circle detection
        self.symbols: List[Symbol] = []
        self.connections: List[Connection] = []
        self.pattern_detector = AdvancedPatternDetector()
    
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
        
        # Track detected star positions to avoid duplicates
        detected_positions = set()
        
        for contour in contours:
            area = cv2.contourArea(contour)
            if area < self.min_contour_area:  # Minimum area check
                continue
            
            # Skip the outer circle contour
            if area > outer_circle.size * outer_circle.size * 2:  # Too large to be a star
                continue
            
            # Get center
            M = cv2.moments(contour)
            if M["m00"] != 0:
                cx = int(M["m10"] / M["m00"])
                cy = int(M["m01"] / M["m00"])
                
                # Check if we already detected a star at this position
                pos_key = (cx // 10, cy // 10)  # Grid-based deduplication
                if pos_key in detected_positions:
                    continue
                
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
                        detected_positions.add(pos_key)
    
    def _is_star_shape(self, contour, cx: int, cy: int) -> bool:
        """Check if contour is star-shaped (simplified)"""
        # Use approximated polygon to check for star shape
        epsilon = 0.02 * cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, epsilon, True)
        
        # Stars typically have 8-12 vertices when approximated
        # (5 points + 5 inner vertices)
        num_vertices = len(approx)
        if 8 <= num_vertices <= 12:
            return True
            
        # Alternative: check for significant variation in distances
        distances = []
        for point in contour:
            x, y = point[0]
            dist = np.sqrt((x - cx)**2 + (y - cy)**2)
            distances.append(dist)
        
        if len(distances) < 5:
            return False
        
        distances = np.array(distances)
        # Check if there's significant variation (star-like pattern)
        return np.std(distances) > np.mean(distances) * 0.15
    
    def _detect_operators(self, binary: np.ndarray, outer_circle: Symbol):
        """Detect operator symbols (convergence, divergence, etc.)"""
        # Find contours that might be operators
        contours, _ = cv2.findContours(binary, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
        
        for contour in contours:
            area = cv2.contourArea(contour)
            if area < self.min_contour_area or area > 5000:  # Operators are medium-sized
                continue
            
            # Get center
            M = cv2.moments(contour)
            if M["m00"] == 0:
                continue
                
            cx = int(M["m10"] / M["m00"])
            cy = int(M["m01"] / M["m00"])
            
            # Check if inside outer circle
            dist = np.sqrt((cx - outer_circle.position[0])**2 + (cy - outer_circle.position[1])**2)
            if dist >= outer_circle.size * 0.8:
                continue
            
            # Analyze the shape to determine operator type
            operator_type = self._identify_operator(contour, binary, cx, cy)
            if operator_type:
                _, (width, height), _ = cv2.minAreaRect(contour)
                size = max(width, height)
                
                symbol = Symbol(
                    type=operator_type,
                    position=(cx, cy),
                    size=size,
                    confidence=0.8,
                    properties={}
                )
                self.symbols.append(symbol)
    
    def _identify_operator(self, contour, binary: np.ndarray, cx: int, cy: int) -> Optional[SymbolType]:
        """Identify specific operator type from contour"""
        # Get bounding box
        x, y, w, h = cv2.boundingRect(contour)
        
        # Extract ROI
        roi = binary[y:y+h, x:x+w]
        
        if roi.size == 0:
            return None
        
        # Detect line patterns to identify operators
        edges = cv2.Canny(roi, 50, 150)
        lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=10, minLineLength=5, maxLineGap=3)
        
        if lines is None:
            return None
        
        # Analyze line directions
        converging = 0
        diverging = 0
        crossing = 0
        
        # Group lines by their angles and positions
        line_angles = []
        for line in lines:
            x1, y1, x2, y2 = line[0]
            angle = np.arctan2(y2 - y1, x2 - x1)
            line_angles.append(angle)
        
        if len(lines) >= 2:
            # Check for convergence pattern (⟐): lines meeting at a point
            # Check for divergence pattern (⟑): lines spreading from a point
            # This is simplified - real implementation would need more sophisticated analysis
            
            # For now, use simple heuristics
            if self._check_convergence_pattern(roi):
                return SymbolType.CONVERGENCE
            elif self._check_divergence_pattern(roi):
                return SymbolType.DIVERGENCE
            elif self._check_amplification_pattern(roi):
                return SymbolType.AMPLIFICATION
            elif self._check_distribution_pattern(roi):
                return SymbolType.DISTRIBUTION
        
        return None
    
    def _check_convergence_pattern(self, roi: np.ndarray) -> bool:
        """Check for convergence pattern (⟐)"""
        # Simple check: more pixels on one side than the other
        h, w = roi.shape
        left_half = roi[:, :w//2]
        right_half = roi[:, w//2:]
        
        left_pixels = cv2.countNonZero(left_half)
        right_pixels = cv2.countNonZero(right_half)
        
        # Convergence has more pixels on the wide side
        return left_pixels > right_pixels * 1.5 or right_pixels > left_pixels * 1.5
    
    def _check_divergence_pattern(self, roi: np.ndarray) -> bool:
        """Check for divergence pattern (⟑)"""
        # Similar to convergence but inverted
        return self._check_convergence_pattern(roi)
    
    def _check_amplification_pattern(self, roi: np.ndarray) -> bool:
        """Check for amplification pattern (✦)"""
        # Check for star-like or cross pattern
        h, w = roi.shape
        center_y, center_x = h // 2, w // 2
        
        # Check for lines radiating from center
        # Simplified check: pixels along diagonals and axes
        diagonal1 = np.diagonal(roi)
        diagonal2 = np.diagonal(np.fliplr(roi))
        horizontal = roi[center_y, :]
        vertical = roi[:, center_x]
        
        # Count non-zero pixels along these lines
        total_line_pixels = (np.count_nonzero(diagonal1) + np.count_nonzero(diagonal2) +
                           np.count_nonzero(horizontal) + np.count_nonzero(vertical))
        
        total_pixels = cv2.countNonZero(roi)
        
        # Amplification has most pixels along the main axes
        return total_pixels > 0 and total_line_pixels / total_pixels > 0.6
    
    def _check_distribution_pattern(self, roi: np.ndarray) -> bool:
        """Check for distribution pattern (⟠)"""
        # Check for circular or hexagonal pattern
        # Simplified: check if pixels form a ring
        h, w = roi.shape
        center_y, center_x = h // 2, w // 2
        
        # Create masks for inner and outer regions
        mask_outer = np.zeros_like(roi)
        mask_inner = np.zeros_like(roi)
        
        cv2.circle(mask_outer, (center_x, center_y), min(h, w) // 2 - 2, 255, -1)
        cv2.circle(mask_inner, (center_x, center_y), min(h, w) // 4, 255, -1)
        
        ring_mask = cv2.bitwise_xor(mask_outer, mask_inner)
        ring_pixels = cv2.bitwise_and(roi, ring_mask)
        
        ring_pixel_count = cv2.countNonZero(ring_pixels)
        total_pixels = cv2.countNonZero(roi)
        
        # Distribution has most pixels in a ring pattern
        return total_pixels > 0 and ring_pixel_count / total_pixels > 0.5
    
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
                    # Skip connections that are too close to symbols themselves
                    # (avoid detecting symbol borders as connections)
                    if self._is_valid_connection(start_symbol, end_symbol, x1, y1, x2, y2):
                        # Determine connection direction based on symbol types
                        from_sym, to_sym = self._determine_connection_direction(
                            start_symbol, end_symbol, x1, y1, x2, y2
                        )
                        
                        connection = Connection(
                            from_symbol=from_sym,
                            to_symbol=to_sym,
                            connection_type="solid"
                        )
                        self.connections.append(connection)
    
    def _is_valid_connection(self, sym1: Symbol, sym2: Symbol, x1: int, y1: int, x2: int, y2: int) -> bool:
        """Check if a detected line is a valid connection between symbols"""
        # Calculate distances from line endpoints to symbol centers
        dist1_to_center = np.sqrt((x1 - sym1.position[0])**2 + (y1 - sym1.position[1])**2)
        dist2_to_center = np.sqrt((x2 - sym2.position[0])**2 + (y2 - sym2.position[1])**2)
        
        # Connection should start/end near symbol edges, not centers
        min_dist = min(sym1.size, sym2.size) * 0.3
        max_dist = max(sym1.size, sym2.size) * 1.2
        
        return (min_dist < dist1_to_center < max_dist and 
                min_dist < dist2_to_center < max_dist)
    
    def _determine_connection_direction(self, sym1: Symbol, sym2: Symbol, 
                                      x1: int, y1: int, x2: int, y2: int) -> Tuple[Symbol, Symbol]:
        """Determine the direction of connection based on symbol types and positions"""
        # Rules for connection direction:
        # 1. Squares/circles (data) -> operators
        # 2. Operators -> outputs (stars)
        # 3. Functions -> outputs
        # 4. Main (double circle) -> statements
        
        # Check symbol types
        data_types = [SymbolType.SQUARE, SymbolType.CIRCLE]
        operator_types = [SymbolType.CONVERGENCE, SymbolType.DIVERGENCE, 
                         SymbolType.AMPLIFICATION, SymbolType.DISTRIBUTION]
        output_types = [SymbolType.STAR]
        control_types = [SymbolType.TRIANGLE, SymbolType.PENTAGON, SymbolType.HEXAGON]
        
        # Data flows to operators
        if sym1.type in data_types and sym2.type in operator_types:
            return (sym1, sym2)
        elif sym2.type in data_types and sym1.type in operator_types:
            return (sym2, sym1)
        
        # Operators flow to outputs
        if sym1.type in operator_types and sym2.type in output_types:
            return (sym1, sym2)
        elif sym2.type in operator_types and sym1.type in output_types:
            return (sym2, sym1)
        
        # Main entry flows to everything
        if sym1.type == SymbolType.DOUBLE_CIRCLE:
            return (sym1, sym2)
        elif sym2.type == SymbolType.DOUBLE_CIRCLE:
            return (sym2, sym1)
        
        # Control flow based on position (top to bottom, left to right)
        if sym1.position[1] < sym2.position[1] - 20:  # sym1 is above sym2
            return (sym1, sym2)
        elif sym2.position[1] < sym1.position[1] - 20:  # sym2 is above sym1
            return (sym2, sym1)
        elif sym1.position[0] < sym2.position[0]:  # sym1 is left of sym2
            return (sym1, sym2)
        else:
            return (sym2, sym1)
    
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
        # Use advanced pattern detector
        pattern_info = self.pattern_detector.analyze_pattern(binary, x, y, radius)
        
        # Map pattern info to string for backward compatibility
        pattern_map = {
            "dot": "dot",
            "double_dot": "double_dot",
            "triple_dot": "triple_line",  # Map 3 dots to triple_line
            "single_line": "lines",
            "double_line": "double_line",
            "triple_line": "triple_line",
            "cross": "cross",
            "half_circle": "half_circle",
            "grid": "grid",
            "filled": "filled",
            "empty": "empty"
        }
        
        return pattern_map.get(pattern_info.pattern_type, pattern_info.pattern_type)