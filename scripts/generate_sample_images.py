#!/usr/bin/env python3
"""
Grimoireサンプル画像生成スクリプト - 魔法陣スタイル版

このスクリプトは、Grimoire言語の記号ベースプログラムの
サンプル画像を魔法陣形式で生成します。
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
    OUTER_CIRCLE_WIDTH = 5
    
# 図形のサイズ定数
class Sizes:
    CANVAS_SIZE = 600  # 正方形のキャンバス
    OUTER_RADIUS = 250  # 外周円の半径
    INNER_RADIUS = 200  # 内部要素の配置可能半径
    CIRCLE_RADIUS = 30
    SQUARE_SIZE = 35
    STAR_SIZE = 30
    TRIANGLE_SIZE = 35
    PENTAGON_SIZE = 35
    HEXAGON_SIZE = 40
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

class MagicCircleDrawer:
    """魔法陣描画を担当するクラス"""
    
    def __init__(self, size: int = Sizes.CANVAS_SIZE):
        self.size = size
        self.center = Point(size // 2, size // 2)
        self.img = Image.new('RGB', (size, size), Colors.BACKGROUND)
        self.draw = ImageDraw.Draw(self.img)
        
    def draw_outer_circle(self, double: bool = False):
        """外周円を描画（必須要素）"""
        x, y = self.center.to_tuple()
        radius = Sizes.OUTER_RADIUS
        
        if double:
            # 二重円（メインエントリ）
            self.draw.ellipse(
                [x-radius-5, y-radius-5, x+radius+5, y+radius+5], 
                outline=Colors.FOREGROUND, 
                width=Colors.OUTER_CIRCLE_WIDTH
            )
            self.draw.ellipse(
                [x-radius, y-radius, x+radius, y+radius], 
                outline=Colors.FOREGROUND, 
                width=Colors.OUTER_CIRCLE_WIDTH
            )
        else:
            # 通常の外周円
            self.draw.ellipse(
                [x-radius, y-radius, x+radius, y+radius], 
                outline=Colors.FOREGROUND, 
                width=Colors.OUTER_CIRCLE_WIDTH
            )
    
    def get_position_on_circle(self, angle: float, radius: float) -> Point:
        """円周上の位置を計算"""
        x = self.center.x + radius * math.cos(angle)
        y = self.center.y + radius * math.sin(angle)
        return Point(x, y)
    
    def draw_radial_lines(self, count: int):
        """放射状の線を描画"""
        for i in range(count):
            angle = (i * 360 / count - 90) * math.pi / 180
            start = self.get_position_on_circle(angle, 50)
            end = self.get_position_on_circle(angle, Sizes.INNER_RADIUS)
            self.draw.line(
                [start.to_tuple(), end.to_tuple()], 
                fill=Colors.FOREGROUND, 
                width=2
            )

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
        elif style == "curved":
            # 曲線的な接続（ベジエ曲線の簡易版）
            self._draw_curved_line(start, end, width)
    
    def _draw_curved_line(self, start: Point, end: Point, width: int):
        """曲線を描画"""
        # 制御点を計算
        mid_x = (start.x + end.x) / 2
        mid_y = (start.y + end.y) / 2
        
        # 簡易的な曲線を複数の線分で近似
        points = []
        for t in range(0, 11):
            t = t / 10.0
            x = (1-t)**2 * start.x + 2*(1-t)*t * mid_x + t**2 * end.x
            y = (1-t)**2 * start.y + 2*(1-t)*t * (mid_y - 30) + t**2 * end.y
            points.append((x, y))
        
        for i in range(len(points) - 1):
            self.draw.line([points[i], points[i+1]], fill=Colors.FOREGROUND, width=width)

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

def create_hello_world() -> Image.Image:
    """Hello World相当（星を表示）- 魔法陣スタイル"""
    magic = MagicCircleDrawer()
    drawer = ShapeDrawer(magic.draw)
    
    # 外周円（二重円でメインエントリを示す）
    magic.draw_outer_circle(double=True)
    
    # 中心に星を配置
    drawer.star(magic.center, Sizes.STAR_SIZE)
    
    return magic.img

def create_fibonacci() -> Image.Image:
    """フィボナッチプログラム - 魔法陣スタイル"""
    magic = MagicCircleDrawer()
    drawer = ShapeDrawer(magic.draw)
    numbers = NumericSymbols(drawer)
    operators = OperatorSymbols(magic.draw)
    
    # 外周円
    magic.draw_outer_circle()
    
    # 中心に関数円
    drawer.circle(magic.center, Sizes.CIRCLE_RADIUS)
    
    # パラメータ（上部）
    param_pos = Point(magic.center.x, magic.center.y - 80)
    drawer.square(param_pos, Sizes.SQUARE_SIZE, PatternType.DOT)
    drawer.connection(Point(magic.center.x, magic.center.y - 30), param_pos)
    
    # 条件分岐（中央やや下）
    triangle_pos = Point(magic.center.x, magic.center.y + 40)
    drawer.triangle(triangle_pos, Sizes.TRIANGLE_SIZE)
    
    # 条件記号
    operators.draw_operator(Point(triangle_pos.x - 20, triangle_pos.y), "≤")
    numbers.draw_dots(Point(triangle_pos.x + 20, triangle_pos.y), 1)
    
    # 左側：直接返す
    left_pos = Point(magic.center.x - 100, magic.center.y + 120)
    drawer.star(left_pos, 20)
    drawer.connection(Point(triangle_pos.x - 20, triangle_pos.y + 20), left_pos, "curved")
    
    # 右側：再帰計算
    right_pos1 = Point(magic.center.x + 60, magic.center.y + 100)
    right_pos2 = Point(magic.center.x + 120, magic.center.y + 100)
    
    drawer.circle(right_pos1, 20)
    operators.draw_operator(Point(right_pos1.x, right_pos1.y - 30), "-")
    numbers.draw_dots(Point(right_pos1.x, right_pos1.y + 25), 1)
    
    drawer.circle(right_pos2, 20)
    operators.draw_operator(Point(right_pos2.x, right_pos2.y - 30), "-")
    numbers.draw_dots(Point(right_pos2.x, right_pos2.y + 25), 2)
    
    # 加算と出力
    add_pos = Point(magic.center.x + 90, magic.center.y + 160)
    operators.draw_operator(add_pos, "+")
    
    output_pos = Point(magic.center.x, magic.center.y + 180)
    drawer.star(output_pos, 20)
    
    # 接続線
    drawer.connection(Point(triangle_pos.x + 20, triangle_pos.y + 20), right_pos1, "curved")
    drawer.connection(right_pos1, add_pos)
    drawer.connection(right_pos2, add_pos)
    drawer.connection(add_pos, output_pos)
    
    return magic.img

def create_variables() -> Image.Image:
    """変数の例 - 魔法陣スタイル"""
    magic = MagicCircleDrawer()
    drawer = ShapeDrawer(magic.draw)
    numbers = NumericSymbols(drawer)
    operators = OperatorSymbols(magic.draw)
    
    # 外周円
    magic.draw_outer_circle()
    
    # 中心に向かって放射状に配置
    angles = [0, 72, 144, 216, 288]  # 五芒星の頂点
    radius = 120
    
    # 各データ型を配置
    patterns = [
        PatternType.DOT,      # 整数
        PatternType.DOUBLE_DOT,  # 浮動小数点
        PatternType.LINES,    # 文字列
        PatternType.HALF_CIRCLE,  # ブール
        PatternType.STARS     # 配列
    ]
    
    for i, (angle, pattern) in enumerate(zip(angles, patterns)):
        angle_rad = (angle - 90) * math.pi / 180
        pos = magic.get_position_on_circle(angle_rad, radius)
        drawer.square(pos, Sizes.SQUARE_SIZE, pattern)
        
        # 中心から放射状の線
        drawer.connection(magic.center, pos)
        
        # 値を表示（整数型の例）
        if pattern == PatternType.DOT:
            value_pos = magic.get_position_on_circle(angle_rad, radius + 50)
            operators.draw_operator(Point(value_pos.x - 10, value_pos.y), "=")
            numbers.draw_dots(Point(value_pos.x + 10, value_pos.y), 5)
    
    # 中心に小さな円
    drawer.circle(magic.center, 20)
    
    return magic.img

def create_parallel() -> Image.Image:
    """並列処理の例 - 魔法陣スタイル"""
    magic = MagicCircleDrawer()
    drawer = ShapeDrawer(magic.draw)
    operators = OperatorSymbols(magic.draw)
    
    # 外周円
    magic.draw_outer_circle()
    
    # 中心に六角形（並列処理）
    drawer.hexagon(magic.center, Sizes.HEXAGON_SIZE)
    
    # 6つの頂点に向けて配置
    tasks = ["☆", "♪", "✉", "☀", "☾", "✓"]
    for i in range(6):
        angle = (i * 60) * math.pi / 180
        
        # タスク円を配置
        task_pos = magic.get_position_on_circle(angle, 120)
        drawer.circle(task_pos, 25)
        
        # タスクのシンボル
        symbol_pos = magic.get_position_on_circle(angle, 160)
        operators.draw_operator(symbol_pos, tasks[i])
        
        # 中心から接続
        drawer.connection(magic.center, task_pos)
    
    return magic.img

def create_calculator() -> Image.Image:
    """計算プログラム - 魔法陣スタイル"""
    magic = MagicCircleDrawer()
    drawer = ShapeDrawer(magic.draw)
    numbers = NumericSymbols(drawer)
    operators = OperatorSymbols(magic.draw)
    
    # 外周円（二重円でメインエントリ）
    magic.draw_outer_circle(double=True)
    
    # 上部に2つの変数
    var1_pos = Point(magic.center.x - 60, magic.center.y - 100)
    var2_pos = Point(magic.center.x + 60, magic.center.y - 100)
    
    drawer.square(var1_pos, Sizes.SQUARE_SIZE, PatternType.DOT)
    drawer.square(var2_pos, Sizes.SQUARE_SIZE, PatternType.DOT)
    
    # 値を表示
    operators.draw_operator(Point(var1_pos.x - 30, var1_pos.y), "=")
    numbers.draw_dots(Point(var1_pos.x - 50, var1_pos.y), 10)
    
    operators.draw_operator(Point(var2_pos.x + 30, var2_pos.y), "=")
    numbers.draw_dots(Point(var2_pos.x + 50, var2_pos.y), 10)
    numbers.draw_dots(Point(var2_pos.x + 50, var2_pos.y + 15), 10)
    
    # 中央に演算子
    operators.draw_operator(magic.center, "+")
    
    # 中央下に第二の演算子
    mul_pos = Point(magic.center.x, magic.center.y + 60)
    operators.draw_operator(mul_pos, "×")
    
    # 底部に出力
    output_pos = Point(magic.center.x, magic.center.y + 120)
    drawer.star(output_pos, Sizes.STAR_SIZE)
    
    # 接続線（曲線的に）
    drawer.connection(var1_pos, magic.center, "curved")
    drawer.connection(var2_pos, magic.center, "curved")
    drawer.connection(magic.center, mul_pos)
    drawer.connection(mul_pos, output_pos)
    
    # 装飾的な小円を追加
    for angle in [45, 135, 225, 315]:
        angle_rad = angle * math.pi / 180
        deco_pos = magic.get_position_on_circle(angle_rad, Sizes.INNER_RADIUS)
        drawer.circle(deco_pos, 10)
    
    return magic.img

def create_loop() -> Image.Image:
    """ループの例 - 魔法陣スタイル"""
    magic = MagicCircleDrawer()
    drawer = ShapeDrawer(magic.draw)
    numbers = NumericSymbols(drawer)
    operators = OperatorSymbols(magic.draw)
    
    # 外周円
    magic.draw_outer_circle()
    
    # 中心に五角形（ループ）
    drawer.pentagon(magic.center, Sizes.PENTAGON_SIZE)
    
    # ループ回数
    count_pos = Point(magic.center.x + 80, magic.center.y - 80)
    drawer.square(count_pos, Sizes.SQUARE_SIZE, PatternType.DOT)
    numbers.draw_dots(Point(count_pos.x + 40, count_pos.y), 10)
    operators.draw_operator(Point(count_pos.x - 30, count_pos.y), "←")
    drawer.connection(magic.center, count_pos, "curved")
    
    # ループ内の処理（下部に星）
    star_pos = Point(magic.center.x, magic.center.y + 80)
    drawer.star(star_pos, Sizes.STAR_SIZE)
    drawer.connection(magic.center, star_pos)
    
    # ループバック矢印
    loop_start = Point(magic.center.x - 100, magic.center.y + 120)
    loop_end = Point(magic.center.x - 100, magic.center.y - 60)
    operators.draw_operator(loop_start, "⟲")
    
    # 円形のループパスを描画
    for i in range(180, 360, 10):
        angle1 = i * math.pi / 180
        angle2 = (i + 10) * math.pi / 180
        pos1 = magic.get_position_on_circle(angle1, 150)
        pos2 = magic.get_position_on_circle(angle2, 150)
        drawer.connection(pos1, pos2)
    
    return magic.img

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
        ("loop.png", create_loop()),
    ]
    
    for filename, img in samples:
        filepath = os.path.join(OUTPUT_DIR, filename)
        img.save(filepath)
        print(f"Generated: {filepath}")

if __name__ == "__main__":
    main()