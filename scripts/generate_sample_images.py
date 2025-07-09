#!/usr/bin/env python3
# 注: このスクリプトを実行する前に、プロジェクトルートで以下を実行:
# uv venv
# uv pip install -e .
"""
Grimoireサンプル画像生成スクリプト
"""

from PIL import Image, ImageDraw, ImageFont
import os

# 出力ディレクトリ
OUTPUT_DIR = "examples/images"

def create_canvas(width=600, height=400):
    """白いキャンバスを作成"""
    img = Image.new('RGB', (width, height), 'white')
    draw = ImageDraw.Draw(img)
    return img, draw

def draw_circle(draw, center, radius, double=False, label=None):
    """円を描画"""
    x, y = center
    if double:
        # 二重円
        draw.ellipse([x-radius-3, y-radius-3, x+radius+3, y+radius+3], outline='black', width=2)
        draw.ellipse([x-radius, y-radius, x+radius, y+radius], outline='black', width=2)
    else:
        draw.ellipse([x-radius, y-radius, x+radius, y+radius], outline='black', width=2)
    
    if label:
        # ラベルを描画（可能な限りシンプルに）
        try:
            font = ImageFont.load_default()
            bbox = draw.textbbox((0, 0), label, font=font)
            text_width = bbox[2] - bbox[0]
            text_height = bbox[3] - bbox[1]
            draw.text((x - text_width//2, y - text_height//2), label, fill='black', font=font)
        except:
            # フォントが利用できない場合は中心に点を描く
            draw.ellipse([x-2, y-2, x+2, y+2], fill='black')

def draw_square(draw, center, size, label=None):
    """四角を描画"""
    x, y = center
    half = size // 2
    draw.rectangle([x-half, y-half, x+half, y+half], outline='black', width=2)
    
    if label:
        try:
            font = ImageFont.load_default()
            bbox = draw.textbbox((0, 0), label, font=font)
            text_width = bbox[2] - bbox[0]
            text_height = bbox[3] - bbox[1]
            draw.text((x - text_width//2, y - text_height//2), label, fill='black', font=font)
        except:
            pass

def draw_star(draw, center, size, points=5):
    """星を描画"""
    x, y = center
    # 簡略化された星（十字で代用）
    draw.line([x-size, y, x+size, y], fill='black', width=2)
    draw.line([x, y-size, x, y+size], fill='black', width=2)
    draw.line([x-size//2, y-size//2, x+size//2, y+size//2], fill='black', width=2)
    draw.line([x-size//2, y+size//2, x+size//2, y-size//2], fill='black', width=2)

def draw_connection(draw, start, end):
    """接続線を描画"""
    draw.line([start, end], fill='black', width=2)

def create_hello_world():
    """Hello Worldプログラム"""
    img, draw = create_canvas(400, 500)
    
    # メイン円（二重円）
    draw_circle(draw, (200, 100), 40, double=True)
    
    # 接続線
    draw_connection(draw, (200, 140), (200, 200))
    
    # 出力星
    draw_star(draw, (200, 200), 30)
    
    # 接続線
    draw_connection(draw, (200, 230), (200, 300))
    
    # テキスト
    draw.text((130, 290), '"Hello World"', fill='black')
    
    return img

def create_fibonacci():
    """フィボナッチプログラム"""
    img, draw = create_canvas(600, 700)
    
    # 関数定義円
    draw_circle(draw, (300, 100), 40, label="fib")
    
    # パラメータ
    draw_square(draw, (400, 100), 40, label="#n")
    draw_connection(draw, (340, 100), (380, 100))
    
    # 条件分岐（三角形の代わりに菱形で代用）
    points = [(300, 180), (350, 230), (300, 280), (250, 230)]
    draw.polygon(points, outline='black', width=2)
    draw.text((280, 220), "n<=1", fill='black')
    
    # 接続
    draw_connection(draw, (300, 140), (300, 180))
    
    # true分岐
    draw_connection(draw, (250, 230), (150, 300))
    draw_star(draw, (150, 300), 25)
    draw.text((140, 330), "n", fill='black')
    
    # false分岐（再帰）
    draw_connection(draw, (350, 230), (450, 300))
    draw_square(draw, (450, 300), 40, label="a")
    draw_square(draw, (450, 380), 40, label="b")
    draw_square(draw, (450, 460), 40, label="sum")
    
    # メインエントリ
    draw_circle(draw, (300, 600), 40, double=True, label="main")
    
    return img

def create_variables():
    """変数の例"""
    img, draw = create_canvas(500, 600)
    
    # メイン円
    draw_circle(draw, (250, 80), 40, double=True)
    
    y_pos = 180
    
    # 整数変数
    draw_connection(draw, (250, 120), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, label="#age")
    draw.text((320, y_pos - 10), "= 25", fill='black')
    
    # 浮動小数点変数
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, label="~price")
    draw.text((320, y_pos - 10), "= 99.99", fill='black')
    
    # 文字列変数
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, label="$name")
    draw.text((320, y_pos - 10), '= "Grimoire"', fill='black')
    
    # ブール変数
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    draw_square(draw, (250, y_pos), 40, label="?valid")
    draw.text((320, y_pos - 10), "= true", fill='black')
    
    # 配列
    y_pos += 80
    draw_connection(draw, (250, y_pos - 60), (250, y_pos - 20))
    # 連結された四角で配列を表現
    draw_square(draw, (200, y_pos), 30)
    draw_square(draw, (250, y_pos), 30)
    draw_square(draw, (300, y_pos), 30)
    draw.text((350, y_pos - 10), "[1,2,3]", fill='black')
    
    return img

def create_parallel():
    """並列処理の例"""
    img, draw = create_canvas(600, 500)
    
    # メイン円
    draw_circle(draw, (300, 80), 40, double=True, label="main")
    
    # 六角形（菱形で代用）
    hex_points = [(300, 150), (350, 180), (350, 220), (300, 250), (250, 220), (250, 180)]
    draw.polygon(hex_points, outline='black', width=2)
    
    # 接続
    draw_connection(draw, (300, 120), (300, 150))
    
    # 並列タスク
    draw_connection(draw, (250, 220), (150, 300))
    draw_circle(draw, (150, 300), 30, label="task1")
    
    draw_connection(draw, (300, 250), (300, 300))
    draw_circle(draw, (300, 300), 30, label="task2")
    
    draw_connection(draw, (350, 220), (450, 300))
    draw_circle(draw, (450, 300), 30, label="task3")
    
    # 結合
    draw_connection(draw, (150, 330), (250, 380))
    draw_connection(draw, (300, 330), (300, 380))
    draw_connection(draw, (450, 330), (350, 380))
    
    # 下の六角形
    hex_points2 = [(300, 380), (350, 410), (350, 450), (300, 480), (250, 450), (250, 410)]
    draw.polygon(hex_points2, outline='black', width=2)
    
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
    ]
    
    for filename, img in samples:
        filepath = os.path.join(OUTPUT_DIR, filename)
        img.save(filepath)
        print(f"Generated: {filepath}")
    
    # 簡単な計算プログラムも追加
    img, draw = create_canvas(400, 400)
    draw_circle(draw, (200, 80), 40, double=True)
    draw_connection(draw, (200, 120), (200, 180))
    draw_square(draw, (150, 180), 30, label="#a")
    draw.text((100, 210), "= 10", fill='black')
    draw_square(draw, (250, 180), 30, label="#b")
    draw.text((300, 210), "= 20", fill='black')
    draw_connection(draw, (200, 210), (200, 280))
    draw_star(draw, (200, 280), 30)
    draw.text((120, 320), "a + b, a × b", fill='black')
    
    img.save(os.path.join(OUTPUT_DIR, "calculator.png"))
    print(f"Generated: {os.path.join(OUTPUT_DIR, 'calculator.png')}")

if __name__ == "__main__":
    main()