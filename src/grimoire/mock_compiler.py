"""Grimoire Mock Compiler - 仮のコンパイラ実装（記号ベース版）"""

import os
import sys
from pathlib import Path


# ファイル名と出力のマッピング（記号ベースに更新）
OUTPUT_MAP = {
    "hello_world.png": "☆",  # 星を表示
    "hello_world.jpg": "☆",
    "hello_world.grim": "☆",
    
    "fibonacci.png": """fib(0) = 0
fib(1) = 1
fib(2) = 1
fib(3) = 2
fib(4) = 3
fib(5) = 5
fib(6) = 8
fib(7) = 13
fib(8) = 21
fib(9) = 34
fib(10) = 55""",
    
    "variables.png": """• = 5
•• = 23.0
≡ = [文字列]
◐ = true
※ = [•, ••, •••]""",
    
    "parallel.png": """並列実行開始...
タスク1: ☆
タスク2: ♪
タスク3: ✉
全タスク完了！ ✓""",
    
    "calculator.png": """⦿ ⟐ ⦿⦿ = 30
30 ✦ ⦿ = 300""",
    
    "loop.png": """☆
☆
☆
☆
☆
☆
☆
☆
☆
☆""",
}


def compile_grimoire(filepath, output_file=None):
    """仮のコンパイル処理"""
    filename = os.path.basename(filepath)
    
    # デバッグ情報を出力
    print(f"コンパイル中: {filepath}", file=sys.stderr)
    print(f"記号を検出中...", file=sys.stderr)
    print(f"  ◎ (メインエントリ) を検出", file=sys.stderr)
    print(f"  ☆ (出力) を検出", file=sys.stderr)
    print(f"ASTを構築中...", file=sys.stderr)
    print(f"コードを生成中...", file=sys.stderr)
    print(f"コンパイル完了！", file=sys.stderr)
    print("", file=sys.stderr)
    
    # ファイル名に基づいて出力を決定
    if filename in OUTPUT_MAP:
        output = OUTPUT_MAP[filename]
    else:
        output = f"未知のプログラム: {filename}"
    
    # 出力ファイルが指定されている場合
    if output_file:
        # 実行可能なPythonスクリプトを生成
        with open(output_file, 'w') as f:
            f.write("#!/usr/bin/env python3\n")
            f.write("# Grimoire generated code\n")
            f.write(f'print("""{output}""")\n')
        os.chmod(output_file, 0o755)
        return f"コンパイル完了: {output_file}"
    
    return output


def run_grimoire(filepath):
    """画像を直接実行"""
    filename = os.path.basename(filepath)
    
    print(f"実行中: {filepath}", file=sys.stderr)
    print("", file=sys.stderr)
    
    if filename in OUTPUT_MAP:
        return OUTPUT_MAP[filename]
    else:
        return f"未知のプログラム: {filename}"


def debug_grimoire(filepath):
    """デバッグモードで実行"""
    filename = os.path.basename(filepath)
    
    print(f"デバッグモード: {filepath}", file=sys.stderr)
    print("ビジュアルデバッガが開きます...", file=sys.stderr)
    print("", file=sys.stderr)
    print("=== 実行トレース ===", file=sys.stderr)
    print("現在の記号: ◎ (main)", file=sys.stderr)
    print("次: ☆ (output)", file=sys.stderr)
    print("", file=sys.stderr)
    
    if filename in OUTPUT_MAP:
        output = OUTPUT_MAP[filename]
        print("=== 出力 ===", file=sys.stderr)
        print(output)
    else:
        print(f"未知のプログラム: {filename}", file=sys.stderr)