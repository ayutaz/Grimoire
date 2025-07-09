"""Grimoire Mock Compiler - Mock compiler implementation (symbol-based)"""

import os
import sys
import platform
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
    
    "parallel.png": """Starting parallel execution...
Task 1: ☆
Task 2: ♪
Task 3: ✉
All tasks complete! ✓""",
    
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
    """Mock compile process"""
    filename = os.path.basename(filepath)
    
    # Output debug information
    print(f"Compiling: {filepath}", file=sys.stderr)
    print(f"Detecting symbols...", file=sys.stderr)
    print(f"  ◎ (main entry) detected", file=sys.stderr)
    print(f"  ☆ (output) detected", file=sys.stderr)
    print(f"Building AST...", file=sys.stderr)
    print(f"Generating code...", file=sys.stderr)
    print(f"Compilation complete!", file=sys.stderr)
    print("", file=sys.stderr)
    
    # Determine output based on filename
    if filename in OUTPUT_MAP:
        output = OUTPUT_MAP[filename]
    else:
        output = f"Unknown program: {filename}"
    
    # If output file is specified
    if output_file:
        # Generate platform-specific executable file
        if platform.system() == 'Windows':
            # Windows: Generate batch file
            if not output_file.endswith('.bat'):
                output_file = output_file + '.bat'
            with open(output_file, 'w', encoding='utf-8') as f:
                f.write("@echo off\n")
                f.write("python -c \"print(r'''{}''')\"\n".format(output))
        else:
            # Unix-like: Generate shell script
            with open(output_file, 'w', encoding='utf-8') as f:
                f.write("#!/usr/bin/env python3\n")
                f.write("# Grimoire generated code\n")
                f.write(f'print("""{output}""")\n')
            # Grant execute permission only on Unix-like systems
            try:
                os.chmod(output_file, 0o755)
            except OSError:
                pass  # Ignore permission errors
        
        return f"Compilation complete: {output_file}"
    
    return output


def run_grimoire(filepath):
    """Run image directly"""
    filename = os.path.basename(filepath)
    
    print(f"Running: {filepath}", file=sys.stderr)
    print("", file=sys.stderr)
    
    if filename in OUTPUT_MAP:
        return OUTPUT_MAP[filename]
    else:
        return f"Unknown program: {filename}"


def debug_grimoire(filepath):
    """Run in debug mode"""
    filename = os.path.basename(filepath)
    
    print(f"Debug mode: {filepath}", file=sys.stderr)
    print("Opening visual debugger...", file=sys.stderr)
    print("", file=sys.stderr)
    print("=== Execution trace ===", file=sys.stderr)
    print("Current symbol: ◎ (main)", file=sys.stderr)
    print("Next: ☆ (output)", file=sys.stderr)
    print("", file=sys.stderr)
    
    if filename in OUTPUT_MAP:
        output = OUTPUT_MAP[filename]
        print("=== Output ===", file=sys.stderr)
        print(output)
    else:
        print(f"Unknown program: {filename}", file=sys.stderr)