#!/usr/bin/env python3
"""
Grimoireサンプル画像生成スクリプト - 記号ベース版

このスクリプトは、Grimoire言語の記号ベースプログラムの
サンプル画像を生成します。
"""

from PIL import Image, ImageDraw, ImageFont
from dataclasses import dataclass
from typing import Tuple, Optional, List
from enum import Enum
import os
import math

# 出力ディレクトリ
OUTPUT_DIR = "examples/images"

# カラースキーム
class Colors:
    BACKGROUND = 'white'
    FOREGROUND = 'black'
    LINE_WIDTH = 3
    
# 図形のサイズ定数
class Sizes:
    CIRCLE_RADIUS = 40
    SQUARE_SIZE = 40
    STAR_SIZE = 35
    TRIANGLE_SIZE = 40
    PENTAGON_SIZE = 40
    HEXAGON_SIZE = 50
    DOT_RADIUS = 3
    PATTERN_SPACING = 6

@dataclass
class Point:
    """2D座標を表すクラス"""
    x: float
    y: float
    
    def to_tuple(self) -> Tuple[float, float]:
        return (self.x, self.y)

class PatternType(Enum):
    """データ型パターンの種類"""
    DOT = "dot"                # 整数型（•）
    DOUBLE_DOT = "double_dot"  # 浮動小数点型（••）
    LINES = "lines"            # 文字列型（≡）
    HALF_CIRCLE = "half_circle" # ブール型（◐）
    STARS = "stars"            # 配列型（※）
    GRID = "grid"              # マップ型（⊞）

class ShapeDrawer:
    """図形描画を担当するクラス"""
    
    def __init__(self, draw: ImageDraw.Draw):
        self.draw = draw
    
    def circle(self, center: Point, radius: float, double: bool = False, filled: bool = False):
        """円を描画"""
        x, y = center.to_tuple()
        if double:
            # 二重円
            self.draw.ellipse(
                [x-radius-3, y-radius-3, x+radius+3, y+radius+3], 
                outline=Colors.FOREGROUND, 
                width=Colors.LINE_WIDTH
            )
            self.draw.ellipse(
                [x-radius, y-radius, x+radius, y+radius], 
                outline=Colors.FOREGROUND, 
                width=Colors.LINE_WIDTH
            )
        else:
            if filled:
                self.draw.ellipse(
                    [x-radius, y-radius, x+radius, y+radius], 
                    fill=Colors.FOREGROUND
                )
            else:
                self.draw.ellipse(
                    [x-radius, y-radius, x+radius, y+radius], 
                    outline=Colors.FOREGROUND, 
                    width=Colors.LINE_WIDTH
                )
    
    def square(self, center: Point, size: float, pattern: Optional[PatternType] = None):
        """四角を描画（パターン付き）"""
        x, y = center.to_tuple()
        half = size // 2
        
        # 外枠を描画
        self.draw.rectangle(
            [x-half, y-half, x+half, y+half], 
            outline=Colors.FOREGROUND, 
            width=Colors.LINE_WIDTH
        )
        
        # 内部パターンを描画
        if pattern:
            self._draw_pattern(center, size, pattern)
    
    def _draw_pattern(self, center: Point, size: float, pattern: PatternType):
        """内部パターンを描画"""
        x, y = center.to_tuple()
        half = size // 2
        
        if pattern == PatternType.DOT:
            # 整数型: 単一の点
            self.circle(center, Sizes.DOT_RADIUS, filled=True)
            
        elif pattern == PatternType.DOUBLE_DOT:
            # 浮動小数点型: 二重点
            self.circle(Point(x - Sizes.PATTERN_SPACING, y), Sizes.DOT_RADIUS, filled=True)
            self.circle(Point(x + Sizes.PATTERN_SPACING, y), Sizes.DOT_RADIUS, filled=True)
            
        elif pattern == PatternType.LINES:
            # 文字列型: 三本線
            for offset in [-5, 0, 5]:
                self.draw.line(
                    [x-half+10, y+offset, x+half-10, y+offset], 
                    fill=Colors.FOREGROUND, 
                    width=2
                )
                
        elif pattern == PatternType.HALF_CIRCLE:
            # ブール型: 半円
            self.draw.pieslice(
                [x-10, y-10, x+10, y+10], 
                270, 90, 
                fill=Colors.FOREGROUND
            )
            
        elif pattern == PatternType.STARS:
            # 配列型: 星形配置の点
            positions = [
                (-10, -10), (0, -10), (10, -10), 
                (-10, 0), (10, 0),
                (-10, 10), (0, 10), (10, 10)
            ]
            for px, py in positions[:5]:  # 5つの点で星形を表現
                self.circle(Point(x + px, y + py), 2, filled=True)
    
    def star(self, center: Point, size: float):
        """5点星を描画"""
        points = []
        for i in range(10):
            angle = (i * 36 - 90) * math.pi / 180
            r = size if i % 2 == 0 else size * 0.4
            px = center.x + r * math.cos(angle)
            py = center.y + r * math.sin(angle)
            points.append((px, py))
        
        self.draw.polygon(
            points, 
            outline=Colors.FOREGROUND, 
            fill=Colors.BACKGROUND, 
            width=Colors.LINE_WIDTH
        )
    
    def triangle(self, center: Point, size: float):
        """三角形を描画"""
        h = size * 0.866  # sqrt(3)/2
        points = [
            (center.x, center.y - size * 0.577),  # 上頂点
            (center.x - size/2, center.y + h/2),   # 左下
            (center.x + size/2, center.y + h/2)    # 右下
        ]
        
        self.draw.polygon(
            points, 
            outline=Colors.FOREGROUND, 
            fill=Colors.BACKGROUND, 
            width=Colors.LINE_WIDTH
        )
    
    def pentagon(self, center: Point, size: float):
        """五角形を描画"""
        points = []
        for i in range(5):
            angle = (i * 72 - 90) * math.pi / 180
            px = center.x + size * math.cos(angle)
            py = center.y + size * math.sin(angle)
            points.append((px, py))
        
        self.draw.polygon(
            points, 
            outline=Colors.FOREGROUND, 
            fill=Colors.BACKGROUND, 
            width=Colors.LINE_WIDTH
        )
    
    def hexagon(self, center: Point, size: float):
        """六角形を描画"""
        points = []
        for i in range(6):
            angle = (i * 60) * math.pi / 180
            px = center.x + size * math.cos(angle)
            py = center.y + size * math.sin(angle)
            points.append((px, py))
        
        self.draw.polygon(
            points, 
            outline=Colors.FOREGROUND, 
            fill=Colors.BACKGROUND, 
            width=Colors.LINE_WIDTH
        )
    
    def connection(self, start: Point, end: Point, style: str = "solid", width: int = Colors.LINE_WIDTH):
        """接続線を描画"""
        if style == "solid":
            self.draw.line(
                [start.to_tuple(), end.to_tuple()], 
                fill=Colors.FOREGROUND, 
                width=width
            )
        elif style == "dashed":
            self._draw_dashed_line(start, end, width)
    
    def _draw_dashed_line(self, start: Point, end: Point, width: int):
        """破線を描画"""
        x1, y1 = start.to_tuple()
        x2, y2 = end.to_tuple()
        length = math.sqrt((x2-x1)**2 + (y2-y1)**2)
        dash_len = 10
        gap_len = 5
        
        if length > 0:
            dashes = int(length / (dash_len + gap_len))
            for i in range(dashes):
                t1 = i * (dash_len + gap_len) / length
                t2 = min((i * (dash_len + gap_len) + dash_len) / length, 1)
                px1 = x1 + t1 * (x2 - x1)
                py1 = y1 + t1 * (y2 - y1)
                px2 = x1 + t2 * (x2 - x1)
                py2 = y1 + t2 * (y2 - y1)
                self.draw.line(
                    [(px1, py1), (px2, py2)], 
                    fill=Colors.FOREGROUND, 
                    width=width
                )

class NumericSymbols:
    """数値記号を描画するクラス"""
    
    def __init__(self, drawer: ShapeDrawer):
        self.drawer = drawer
    
    def draw_dots(self, center: Point, count: int):
        """点の数で数値を表現"""
        x, y = center.to_tuple()
        
        if count == 1:
            self.drawer.circle(center, Sizes.DOT_RADIUS, filled=True)
        elif count == 2:
            self.drawer.circle(Point(x - 8, y), Sizes.DOT_RADIUS, filled=True)
            self.drawer.circle(Point(x + 8, y), Sizes.DOT_RADIUS, filled=True)
        elif count == 3:
            self.drawer.circle(Point(x - 10, y), Sizes.DOT_RADIUS, filled=True)
            self.drawer.circle(Point(x, y), Sizes.DOT_RADIUS, filled=True)
            self.drawer.circle(Point(x + 10, y), Sizes.DOT_RADIUS, filled=True)
        elif count == 5:
            # 5つの点を十字配置
            self.drawer.circle(Point(x, y), Sizes.DOT_RADIUS, filled=True)
            for dx, dy in [(-10, 0), (10, 0), (0, -10), (0, 10)]:
                self.drawer.circle(Point(x + dx, y + dy), 2, filled=True)
        elif count == 10:
            # 囲み点
            self.drawer.circle(center, 8, filled=False)
            self.drawer.circle(center, Sizes.DOT_RADIUS, filled=True)

class OperatorSymbols:
    """演算子記号を描画するクラス"""
    
    def __init__(self, draw: ImageDraw.Draw):
        self.draw = draw
        try:
            self.font = ImageFont.load_default()
        except:
            self.font = None
    
    def draw_operator(self, center: Point, operator: str):
        """演算子を描画"""
        if self.font:
            x, y = center.to_tuple()
            bbox = self.draw.textbbox((0, 0), operator, font=self.font)
            text_width = bbox[2] - bbox[0]
            text_height = bbox[3] - bbox[1]
            self.draw.text(
                (x - text_width//2, y - text_height//2), 
                operator, 
                fill=Colors.FOREGROUND, 
                font=self.font
            )

class ProgramGenerator:
    """プログラム画像を生成するクラス"""
    
    def __init__(self, width: int = 600, height: int = 400):
        self.img = Image.new('RGB', (width, height), Colors.BACKGROUND)
        self.draw = ImageDraw.Draw(self.img)
        self.drawer = ShapeDrawer(self.draw)
        self.numbers = NumericSymbols(self.drawer)
        self.operators = OperatorSymbols(self.draw)
    
    def save(self, filename: str):
        """画像を保存"""
        filepath = os.path.join(OUTPUT_DIR, filename)
        self.img.save(filepath)
        return filepath

def create_hello_world() -> Image.Image:
    """Hello World相当（星を表示）"""
    gen = ProgramGenerator(400, 300)
    
    # メインエントリ（二重円）
    gen.drawer.circle(Point(200, 100), Sizes.CIRCLE_RADIUS, double=True)
    
    # 接続線
    gen.drawer.connection(Point(200, 140), Point(200, 200))
    
    # 出力星
    gen.drawer.star(Point(200, 200), Sizes.STAR_SIZE)
    
    return gen.img

def create_fibonacci() -> Image.Image:
    """フィボナッチプログラム（記号版）"""
    gen = ProgramGenerator(600, 700)
    
    # 関数定義円
    gen.drawer.circle(Point(300, 100), Sizes.CIRCLE_RADIUS)
    
    # パラメータ（整数型）
    gen.drawer.square(Point(400, 100), Sizes.SQUARE_SIZE, PatternType.DOT)
    gen.drawer.connection(Point(340, 100), Point(380, 100))
    
    # 条件分岐
    gen.drawer.triangle(Point(300, 200), Sizes.TRIANGLE_SIZE)
    gen.drawer.connection(Point(300, 140), Point(300, 170))
    
    # 条件記号
    gen.operators.draw_operator(Point(270, 195), "≤")
    gen.numbers.draw_dots(Point(300, 200), 1)
    
    # true分岐
    gen.drawer.connection(Point(270, 220), Point(200, 300))
    gen.drawer.star(Point(200, 300), 25)
    
    # false分岐（再帰）
    gen.drawer.connection(Point(330, 220), Point(400, 300))
    
    # 再帰呼び出し
    gen.drawer.circle(Point(350, 350), 25)
    gen.operators.draw_operator(Point(340, 345), "-")
    gen.numbers.draw_dots(Point(360, 350), 1)
    
    gen.drawer.circle(Point(450, 350), 25)
    gen.operators.draw_operator(Point(440, 345), "-")
    gen.numbers.draw_dots(Point(460, 350), 2)
    
    # 加算
    gen.drawer.connection(Point(350, 375), Point(400, 420))
    gen.drawer.connection(Point(450, 375), Point(400, 420))
    gen.operators.draw_operator(Point(395, 415), "+")
    
    # 出力
    gen.drawer.connection(Point(400, 440), Point(400, 480))
    gen.drawer.star(Point(400, 500), 25)
    
    # メインエントリ
    gen.drawer.circle(Point(300, 600), Sizes.CIRCLE_RADIUS, double=True)
    
    return gen.img

def create_variables() -> Image.Image:
    """変数の例（記号版）"""
    gen = ProgramGenerator(500, 600)
    
    # メインエントリ
    gen.drawer.circle(Point(250, 80), Sizes.CIRCLE_RADIUS, double=True)
    
    y_pos = 180
    
    # 整数変数
    gen.drawer.connection(Point(250, 120), Point(250, y_pos - 20))
    gen.drawer.square(Point(250, y_pos), Sizes.SQUARE_SIZE, PatternType.DOT)
    gen.operators.draw_operator(Point(300, y_pos - 10), "=")
    gen.numbers.draw_dots(Point(340, y_pos), 5)
    
    # 浮動小数点変数
    y_pos += 80
    gen.drawer.connection(Point(250, y_pos - 60), Point(250, y_pos - 20))
    gen.drawer.square(Point(250, y_pos), Sizes.SQUARE_SIZE, PatternType.DOUBLE_DOT)
    
    # 文字列変数
    y_pos += 80
    gen.drawer.connection(Point(250, y_pos - 60), Point(250, y_pos - 20))
    gen.drawer.square(Point(250, y_pos), Sizes.SQUARE_SIZE, PatternType.LINES)
    
    # ブール変数（true）
    y_pos += 80
    gen.drawer.connection(Point(250, y_pos - 60), Point(250, y_pos - 20))
    gen.drawer.square(Point(250, y_pos), Sizes.SQUARE_SIZE, PatternType.HALF_CIRCLE)
    
    # 配列
    y_pos += 80
    gen.drawer.connection(Point(250, y_pos - 60), Point(250, y_pos - 20))
    gen.drawer.square(Point(250, y_pos), Sizes.SQUARE_SIZE, PatternType.STARS)
    
    return gen.img

def create_parallel() -> Image.Image:
    """並列処理の例（記号版）"""
    gen = ProgramGenerator(600, 500)
    
    # メインエントリ
    gen.drawer.circle(Point(300, 80), Sizes.CIRCLE_RADIUS, double=True)
    
    # 六角形（並列処理）
    gen.drawer.hexagon(Point(300, 180), Sizes.HEXAGON_SIZE)
    gen.drawer.connection(Point(300, 120), Point(300, 130))
    
    # 並列タスク
    gen.drawer.connection(Point(250, 200), Point(150, 280))
    gen.drawer.circle(Point(150, 280), 30)
    gen.drawer.connection(Point(150, 310), Point(150, 340))
    gen.drawer.star(Point(150, 340), 20)
    
    gen.drawer.connection(Point(300, 230), Point(300, 280))
    gen.drawer.circle(Point(300, 280), 30)
    gen.drawer.connection(Point(300, 310), Point(300, 340))
    gen.operators.draw_operator(Point(290, 330), "♪")
    
    gen.drawer.connection(Point(350, 200), Point(450, 280))
    gen.drawer.circle(Point(450, 280), 30)
    gen.drawer.connection(Point(450, 310), Point(450, 340))
    gen.operators.draw_operator(Point(440, 330), "✉")
    
    # 結合
    gen.drawer.connection(Point(150, 360), Point(250, 400))
    gen.drawer.connection(Point(300, 360), Point(300, 400))
    gen.drawer.connection(Point(450, 360), Point(350, 400))
    
    # 下の六角形（同期）
    gen.drawer.hexagon(Point(300, 420), Sizes.HEXAGON_SIZE)
    
    # 完了マーク
    gen.drawer.connection(Point(300, 470), Point(300, 500))
    gen.operators.draw_operator(Point(290, 490), "✓")
    
    return gen.img

def create_calculator() -> Image.Image:
    """計算プログラム（記号版）"""
    gen = ProgramGenerator(400, 400)
    
    # メインエントリ
    gen.drawer.circle(Point(200, 80), Sizes.CIRCLE_RADIUS, double=True)
    gen.drawer.connection(Point(200, 120), Point(200, 180))
    
    # 変数a（整数型、値10）
    gen.drawer.square(Point(150, 180), 35, PatternType.DOT)
    gen.operators.draw_operator(Point(130, 210), "=")
    gen.numbers.draw_dots(Point(150, 220), 10)
    
    # 変数b（整数型、値20）
    gen.drawer.square(Point(250, 180), 35, PatternType.DOT)
    gen.operators.draw_operator(Point(230, 210), "=")
    gen.numbers.draw_dots(Point(250, 220), 10)
    gen.numbers.draw_dots(Point(250, 235), 10)
    
    # 加算
    gen.drawer.connection(Point(150, 215), Point(200, 280))
    gen.drawer.connection(Point(250, 215), Point(200, 280))
    gen.operators.draw_operator(Point(195, 270), "+")
    
    # 乗算
    gen.drawer.connection(Point(200, 290), Point(200, 320))
    gen.operators.draw_operator(Point(195, 310), "×")
    
    # 出力
    gen.drawer.connection(Point(200, 330), Point(200, 360))
    gen.drawer.star(Point(200, 360), 25)
    
    return gen.img

def main():
    """メイン処理"""
    # 出力ディレクトリを作成
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    
    # サンプル画像を生成
    samples = [
        ("hello_world.png", create_hello_world()),
        ("fibonacci.png", create_fibonacci()),
        ("variables.png", create_variables()),
        ("parallel.png", create_parallel()),
        ("calculator.png", create_calculator()),
    ]
    
    for filename, img in samples:
        filepath = os.path.join(OUTPUT_DIR, filename)
        img.save(filepath)
        print(f"Generated: {filepath}")

if __name__ == "__main__":
    main()