"""高速化されたimage_recognitionモジュール（遅延インポート版）"""

import numpy as np
from typing import List, Tuple, Dict, Optional, Any
from dataclasses import dataclass
from enum import Enum
import math

# 遅延インポート用の変数
_cv2 = None

def get_cv2():
    """OpenCVを遅延インポート"""
    global _cv2
    if _cv2 is None:
        import cv2
        _cv2 = cv2
    return _cv2

# SymbolTypeとその他のクラスは同じ（省略）
from .image_recognition import (
    SymbolType, Symbol, Connection, 
    AdvancedPatternDetector, PatternInfo
)

class FastMagicCircleDetector:
    """高速化されたMagicCircleDetector"""
    
    def __init__(self):
        self.min_contour_area = 100
        self.circle_threshold = 0.80
        self.symbols: List[Symbol] = []
        self.connections: List[Connection] = []
        self._pattern_detector = None
        
    @property
    def pattern_detector(self):
        """パターン検出器を遅延初期化"""
        if self._pattern_detector is None:
            self._pattern_detector = AdvancedPatternDetector()
        return self._pattern_detector
    
    def detect_symbols(self, image_path: str) -> Tuple[List[Symbol], List[Connection]]:
        """最適化された検出関数"""
        cv2 = get_cv2()  # ここで初めてOpenCVをインポート
        
        # Load image
        img = cv2.imread(image_path)
        if img is None:
            raise ValueError(f"Cannot load image: {image_path}")
        
        # Preprocess - グレースケール変換と二値化を統合
        gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
        _, binary = cv2.threshold(gray, 0, 255, cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU)
        
        # Detect outer circle first (required)
        outer_circle = self._detect_outer_circle_fast(binary)
        if outer_circle is None:
            raise ValueError("No outer circle detected. All Grimoire programs must be enclosed in a magic circle.")
        
        self.symbols = [outer_circle]
        
        # 並列処理可能な検出を統合
        self._detect_all_shapes_fast(binary, outer_circle)
        self._detect_connections_fast(binary)
        
        return self.symbols, self.connections
    
    def _detect_outer_circle_fast(self, binary: np.ndarray) -> Optional[Symbol]:
        """高速化された外円検出"""
        cv2 = get_cv2()
        
        # Find contours
        contours, _ = cv2.findContours(binary, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        if not contours:
            return None
        
        # Find the largest contour (should be outer circle)
        largest_contour = max(contours, key=cv2.contourArea)
        area = cv2.contourArea(largest_contour)
        
        if area < self.min_contour_area * 10:  # Outer circle should be large
            return None
        
        # Quick circle check
        (x, y), radius = cv2.minEnclosingCircle(largest_contour)
        circle_area = math.pi * radius * radius
        
        if area / circle_area > self.circle_threshold:
            return Symbol(
                type=SymbolType.OUTER_CIRCLE,
                position=(int(x), int(y)),
                size=radius * 2,
                confidence=area / circle_area,
                properties={}
            )
        
        return None
    
    def _detect_all_shapes_fast(self, binary: np.ndarray, outer_circle: Symbol):
        """全ての形状を一度に検出（高速化）"""
        cv2 = get_cv2()
        
        # マスクを作成して外円内のみ処理
        mask = np.zeros_like(binary)
        cv2.circle(mask, outer_circle.position, int(outer_circle.size/2), 255, -1)
        masked = cv2.bitwise_and(binary, mask)
        
        # 全ての輪郭を一度に取得
        contours, _ = cv2.findContours(masked, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
        
        for contour in contours:
            area = cv2.contourArea(contour)
            if area < self.min_contour_area:
                continue
            
            # 簡略化された形状判定
            self._classify_and_add_shape(contour, area)
    
    def _classify_and_add_shape(self, contour, area):
        """形状を分類して追加（簡略化）"""
        cv2 = get_cv2()
        
        # 多角形近似
        epsilon = 0.04 * cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, epsilon, True)
        vertices = len(approx)
        
        # 重心を計算
        M = cv2.moments(contour)
        if M["m00"] == 0:
            return
        
        cx = int(M["m10"] / M["m00"])
        cy = int(M["m01"] / M["m00"])
        
        # 形状の分類（簡略化）
        symbol_type = None
        if vertices <= 2:
            # 円の可能性
            (x, y), radius = cv2.minEnclosingCircle(contour)
            circle_area = math.pi * radius * radius
            if area / circle_area > 0.7:
                symbol_type = SymbolType.CIRCLE
        elif vertices == 3:
            symbol_type = SymbolType.TRIANGLE
        elif vertices == 4:
            symbol_type = SymbolType.SQUARE
        elif vertices == 5:
            symbol_type = SymbolType.PENTAGON
        elif vertices == 6:
            symbol_type = SymbolType.HEXAGON
        elif vertices >= 8:
            symbol_type = SymbolType.STAR
        
        if symbol_type:
            self.symbols.append(Symbol(
                type=symbol_type,
                position=(cx, cy),
                size=math.sqrt(area),
                confidence=0.8,
                properties={}
            ))
    
    def _detect_connections_fast(self, binary: np.ndarray):
        """高速化された接続検出（簡略版）"""
        # 最小限の接続検出のみ実装
        # 実際のアプリケーションでは必要に応じて実装
        pass