#!/usr/bin/env python3
# 注: このスクリプトを実行する前に、プロジェクトルートで以下を実行:
# uv venv
# uv pip install -e .
"""
Grimoireサンプル画像生成スクリプト - 記号ベース版
"""

from PIL import Image, ImageDraw, ImageFont
import os
import math

# 出力ディレクトリ
OUTPUT_DIR = "examples/images"

def create_canvas(width=600, height=400):
    """白いキャンバスを作成"""
    img = Image.new('RGB', (width, height), 'white')
    draw = ImageDraw.Draw(img)
    return img, draw

def draw_circle(draw, center, radius, double=False, filled=False):
    """円を描画"""
    x, y = center
    if double:
        # 二重円
        draw.ellipse([x-radius-3, y-radius-3, x+radius+3, y+radius+3], outline='black', width=3)
        draw.ellipse([x-radius, y-radius, x+radius, y+radius], outline='black', width=3)
    else:
        if filled:
            draw.ellipse([x-radius, y-radius, x+radius, y+radius], fill='black')
        else:
            draw.ellipse([x-radius, y-radius, x+radius, y+radius], outline='black', width=3)

def draw_square(draw, center, size, pattern=None):
    """四角を描画（パターン付き）"""
    x, y = center
    half = size // 2
    draw.rectangle([x-half, y-half, x+half, y+half], outline='black', width=3)
    
    # 内部パターンを描画
    if pattern == "dot":  # 整数型
        draw.ellipse([x-3, y-3, x+3, y+3], fill='black')
    elif pattern == "double_dot":  # 浮動小数点型
        draw.ellipse([x-8, y-3, x-2, y+3], fill='black')
        draw.ellipse([x+2, y-3, x+8, y+3], fill='black')
    elif pattern == "lines":  # 文字列型
        draw.line([x-half+10, y-5, x+half-10, y-5], fill='black', width=2)
        draw.line([x-half+10, y, x+half-10, y], fill='black', width=2)
        draw.line([x-half+10, y+5, x+half-10, y+5], fill='black', width=2)
    elif pattern == "half_circle":  # ブール型
        draw.pieslice([x-10, y-10, x+10, y+10], 270, 90, fill='black')
    elif pattern == "stars":  # 配列型
        positions = [(-10, -10), (0, -10), (10, -10), (-10, 0), (10, 0), (-10, 10), (0, 10), (10, 10)]
        for px, py in positions[:5]:  # 5つの点で星形を表現
            draw.ellipse([x+px-2, y+py-2, x+px+2, y+py+2], fill='black')

def draw_star(draw, center, size):
    """5点星を描画"""
    x, y = center
    # 5点星の座標を計算
    points = []
    for i in range(10):
        angle = (i * 36 - 90) * math.pi / 180
        if i % 2 == 0:
            r = size
        else:
            r = size * 0.4
        px = x + r * math.cos(angle)
        py = y + r * math.sin(angle)
        points.append((px, py))
    draw.polygon(points, outline='black', fill='white', width=3)

def draw_triangle(draw, center, size):
    """三角形を描画"""
    x, y = center
    h = size * 0.866  # sqrt(3)/2
    points = [
        (x, y - size * 0.577),  # 上頂点
        (x - size/2, y + h/2),  # 左下
        (x + size/2, y + h/2)   # 右下
    ]
    draw.polygon(points, outline='black', fill='white', width=3)

def draw_pentagon(draw, center, size):
    """五角形を描画"""
    x, y = center
    points = []
    for i in range(5):
        angle = (i * 72 - 90) * math.pi / 180
        px = x + size * math.cos(angle)
        py = y + size * math.sin(angle)
        points.append((px, py))
    draw.polygon(points, outline='black', fill='white', width=3)

def draw_hexagon(draw, center, size):
    """六角形を描画"""
    x, y = center
    points = []
    for i in range(6):
        angle = (i * 60) * math.pi / 180
        px = x + size * math.cos(angle)
        py = y + size * math.sin(angle)
        points.append((px, py))
    draw.polygon(points, outline='black', fill='white', width=3)

def draw_connection(draw, start, end, style="solid", width=3):
    """接続線を描画"""
    if style == "solid":
        draw.line([start, end], fill='black', width=width)
    elif style == "dashed":
        # 破線を描画
        x1, y1 = start
        x2, y2 = end
        length = ((x2-x1)**2 + (y2-y1)**2)**0.5
        dash_len = 10
        gap_len = 5
        dashes = int(length / (dash_len + gap_len))
        for i in range(dashes):
            t1 = i * (dash_len + gap_len) / length
            t2 = min((i * (dash_len + gap_len) + dash_len) / length, 1)
            px1 = x1 + t1 * (x2 - x1)
            py1 = y1 + t1 * (y2 - y1)
            px2 = x1 + t2 * (x2 - x1)
            py2 = y1 + t2 * (y2 - y1)
            draw.line([(px1, py1), (px2, py2)], fill='black', width=width)

def draw_dots(draw, center, count):
    """点の数を描画（数値表現）"""
    x, y = center
    if count == 1:
        draw.ellipse([x-3, y-3, x+3, y+3], fill='black')
    elif count == 2:
        draw.ellipse([x-8, y-3, x-2, y+3], fill='black')
        draw.ellipse([x+2, y-3, x+8, y+3], fill='black')
    elif count == 3:
        draw.ellipse([x-10, y-3, x-4, y+3], fill='black')
        draw.ellipse([x-3, y-3, x+3, y+3], fill='black')
        draw.ellipse([x+4, y-3, x+10, y+3], fill='black')
    elif count == 10:  # 囲み点
        draw.ellipse([x-8, y-8, x+8, y+8], outline='black', width=2)
        draw.ellipse([x-3, y-3, x+3, y+3], fill='black')

def create_hello_world():
    """Hello World相当（星を表示）"""
    img, draw = create_canvas(400, 300)
    
    # メイン円（二重円）
    draw_circle(draw, (200, 100), 40, double=True)
    
    # 接続線
    draw_connection(draw, (200, 140), (200, 200))
    
    # 出力星
    draw_star(draw, (200, 200), 35)
    
    return img

def create_fibonacci():
    """フィボナッチプログラム（記号版）"""
    img, draw = create_canvas(600, 700)
    
    # 関数定義円
    draw_circle(draw, (300, 100), 40)
    
    # パラメータ（整数型）
    draw_square(draw, (400, 100), 40, pattern="dot")
    draw_connection(draw, (340, 100), (380, 100))
    
    # 条件分岐
    draw_triangle(draw, (300, 200), 40)
    
    # 接続
    draw_connection(draw, (300, 140), (300, 170))
    
    # 条件記号
    draw.text((270, 195), "≤", fill='black', font=ImageFont.load_default())
    draw_dots(draw, (300, 200), 1)
    
    # true分岐
    draw_connection(draw, (270, 220), (200, 300))
    draw_star(draw, (200, 300), 25)
    
    # false分岐（再帰）
    draw_connection(draw, (330, 220), (400, 300))
    
    # 再帰呼び出し
    draw_circle(draw, (350, 350), 25)
    draw.text((340, 345), "-", fill='black', font=ImageFont.load_default())
    draw_dots(draw, (360, 350), 1)
    
    draw_circle(draw, (450, 350), 25)
    draw.text((440, 345), "-", fill='black', font=ImageFont.load_default())
    draw_dots(draw, (460, 350), 2)
    
    # 加算
    draw_connection(draw, (350, 375), (400, 420))
    draw_connection(draw, (450, 375), (400, 420))
    draw.text((395, 415), "+", fill='black', font=ImageFont.load_default())
    
    # 出力
    draw_connection(draw, (400, 440), (400, 480))
    draw_star(draw, (400, 500), 25)
    
    # メインエントリ
    draw_circle(draw, (300, 600), 40, double=True)
    
    return img

def create_variables():
    """変数の例（記号版）"""
    img, draw = create_canvas(500, 600)
    
    # メイン円
    draw_circle(draw, (250, 80), 40, double=True)
    
    y_pos = 180
    
    # 整数変数
    draw_connection(draw, (250, 120), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, pattern="dot")
    draw.text((300, y_pos - 10), "=", fill='black', font=ImageFont.load_default())
    draw_dots(draw, (340, y_pos), 5)  # 5を表現
    
    # 浮動小数点変数
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, pattern="double_dot")
    
    # 文字列変数
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, pattern="lines")
    
    # ブール変数（true）
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, pattern="half_circle")
    
    # 配列
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, pattern="stars")
    
    return img

def create_parallel():
    """並列処理の例（記号版）"""
    img, draw = create_canvas(600, 500)
    
    # メイン円
    draw_circle(draw, (300, 80), 40, double=True)
    
    # 六角形（並列処理）
    draw_hexagon(draw, (300, 180), 50)
    
    # 接続
    draw_connection(draw, (300, 120), (300, 130))
    
    # 並列タスク
    draw_connection(draw, (250, 200), (150, 280))
    draw_circle(draw, (150, 280), 30)
    draw_connection(draw, (150, 310), (150, 340))
    draw_star(draw, (150, 340), 20)
    
    draw_connection(draw, (300, 230), (300, 280))
    draw_circle(draw, (300, 280), 30)
    draw_connection(draw, (300, 310), (300, 340))
    draw.text((290, 330), "♪", fill='black', font=ImageFont.load_default())
    
    draw_connection(draw, (350, 200), (450, 280))
    draw_circle(draw, (450, 280), 30)
    draw_connection(draw, (450, 310), (450, 340))
    draw.text((440, 330), "✉", fill='black', font=ImageFont.load_default())
    
    # 結合
    draw_connection(draw, (150, 360), (250, 400))
    draw_connection(draw, (300, 360), (300, 400))
    draw_connection(draw, (450, 360), (350, 400))
    
    # 下の六角形（同期）
    draw_hexagon(draw, (300, 420), 50)
    
    # 完了マーク
    draw_connection(draw, (300, 470), (300, 500))
    draw.text((290, 490), "✓", fill='black', font=ImageFont.load_default())
    
    return img

def create_calculator():
    """計算プログラム（記号版）"""
    img, draw = create_canvas(400, 400)
    
    # メイン円
    draw_circle(draw, (200, 80), 40, double=True)
    
    # 変数定義
    draw_connection(draw, (200, 120), (200, 180))
    
    # 変数a（整数型、値10）
    draw_square(draw, (150, 180), 35, pattern="dot")
    draw.text((130, 210), "=", fill='black', font=ImageFont.load_default())
    draw_dots(draw, (150, 220), 10)
    
    # 変数b（整数型、値20） 
    draw_square(draw, (250, 180), 35, pattern="dot")
    draw.text((230, 210), "=", fill='black', font=ImageFont.load_default())
    draw_dots(draw, (250, 220), 10)
    draw_dots(draw, (250, 235), 10)
    
    # 加算
    draw_connection(draw, (150, 215), (200, 280))
    draw_connection(draw, (250, 215), (200, 280))
    draw.text((195, 270), "+", fill='black', font=ImageFont.load_default())
    
    # 乗算
    draw_connection(draw, (200, 290), (200, 320))
    draw.text((195, 310), "×", fill='black', font=ImageFont.load_default())
    
    # 出力
    draw_connection(draw, (200, 330), (200, 360))
    draw_star(draw, (200, 360), 25)
    
    return img

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