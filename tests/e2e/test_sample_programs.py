"""
End-to-end tests for complete Grimoire programs
Following t-wada's TDD approach
"""

import pytest
import cv2
import numpy as np
from pathlib import Path
import subprocess
import sys
from grimoire.compiler import GrimoireCompiler


class TestSamplePrograms:
    """実際のプログラム例のテスト"""
    
    @pytest.fixture
    def compiler(self):
        return GrimoireCompiler()
    
    @pytest.fixture
    def samples_dir(self, tmp_path):
        samples = tmp_path / "samples"
        samples.mkdir()
        return samples
    
    @pytest.mark.e2e
    def test_fizzbuzz_program(self, compiler, samples_dir):
        """FizzBuzzプログラムのテスト"""
        # Arrange
        img = self._create_fizzbuzz_program()
        img_path = samples_dir / "fizzbuzz.png"
        cv2.imwrite(str(img_path), img)
        
        # Act
        python_code = compiler.compile_to_python(str(img_path))
        output_path = samples_dir / "fizzbuzz.py"
        with open(output_path, "w") as f:
            f.write(python_code)
        
        # Assert - 生成されたコードが実行可能
        result = subprocess.run(
            [sys.executable, str(output_path)],
            capture_output=True,
            text=True,
            timeout=5
        )
        assert result.returncode == 0
    
    @pytest.mark.e2e
    def test_fibonacci_program(self, compiler, samples_dir):
        """フィボナッチ数列プログラムのテスト"""
        # Arrange
        img = self._create_fibonacci_program()
        img_path = samples_dir / "fibonacci.png"
        cv2.imwrite(str(img_path), img)
        
        # Act
        result = compiler.compile_and_run(str(img_path))
        
        # Assert
        assert result is not None
    
    def _create_fizzbuzz_program(self):
        """FizzBuzzプログラムの魔法陣を作成"""
        img = np.ones((800, 800, 3), dtype=np.uint8) * 255
        
        # 外円
        cv2.circle(img, (400, 400), 380, (0, 0, 0), 3)
        
        # メインエントリ
        cv2.circle(img, (400, 100), 30, (0, 0, 0), 2)
        cv2.circle(img, (400, 100), 25, (0, 0, 0), 2)
        
        # ループ（1から15まで）
        pts = []
        for i in range(5):
            angle = i * 2 * np.pi / 5 - np.pi / 2
            x = int(400 + 80 * np.cos(angle))
            y = int(300 + 80 * np.sin(angle))
            pts.append([x, y])
        cv2.polylines(img, [np.array(pts)], True, (0, 0, 0), 2)
        
        # 条件分岐（3の倍数）
        cv2.polylines(img, [np.array([[350, 250], [300, 330], [400, 330]], np.int32)], True, (0, 0, 0), 2)
        
        # 条件分岐（5の倍数）
        cv2.polylines(img, [np.array([[450, 250], [400, 330], [500, 330]], np.int32)], True, (0, 0, 0), 2)
        
        # 出力
        self._draw_star(img, 400, 450)
        
        return img
    
    def _create_fibonacci_program(self):
        """フィボナッチ数列プログラムの魔法陣を作成"""
        img = np.ones((800, 800, 3), dtype=np.uint8) * 255
        
        # 外円
        cv2.circle(img, (400, 400), 380, (0, 0, 0), 3)
        
        # メインエントリ
        cv2.circle(img, (400, 100), 30, (0, 0, 0), 2)
        cv2.circle(img, (400, 100), 25, (0, 0, 0), 2)
        
        # 変数（前の2つの数）
        cv2.rectangle(img, (250, 200), (310, 260), (0, 0, 0), 2)
        cv2.rectangle(img, (490, 200), (550, 260), (0, 0, 0), 2)
        
        # ループ
        pts = []
        for i in range(5):
            angle = i * 2 * np.pi / 5 - np.pi / 2
            x = int(400 + 100 * np.cos(angle))
            y = int(350 + 100 * np.sin(angle))
            pts.append([x, y])
        cv2.polylines(img, [np.array(pts)], True, (0, 0, 0), 2)
        
        # 加算
        cv2.line(img, (310, 230), (370, 350), (0, 0, 0), 2)
        cv2.line(img, (490, 230), (430, 350), (0, 0, 0), 2)
        cv2.circle(img, (400, 350), 5, (0, 0, 0), -1)
        
        # 出力
        self._draw_star(img, 400, 500)
        
        return img
    
    def _draw_star(self, img, cx, cy, size=40):
        """星を描画"""
        pts = []
        for i in range(10):
            angle = i * np.pi / 5
            r = size if i % 2 == 0 else size * 0.5
            x = int(cx + r * np.cos(angle - np.pi/2))
            y = int(cy + r * np.sin(angle - np.pi/2))
            pts.append([x, y])
        cv2.polylines(img, [np.array(pts)], True, (0, 0, 0), 2)


class TestPerformance:
    """パフォーマンステスト"""
    
    @pytest.fixture
    def compiler(self):
        return GrimoireCompiler()
    
    @pytest.mark.slow
    def test_large_program_compilation_time(self, compiler, tmp_path):
        """大きなプログラムのコンパイル時間"""
        # Arrange
        img = self._create_large_program()
        img_path = tmp_path / "large.png"
        cv2.imwrite(str(img_path), img)
        
        # Act
        import time
        start = time.time()
        result = compiler.compile_and_run(str(img_path))
        end = time.time()
        
        # Assert
        assert end - start < 5.0  # 5秒以内
        assert result is not None
    
    def _create_large_program(self):
        """多数のシンボルを含む大きなプログラム"""
        img = np.ones((1000, 1000, 3), dtype=np.uint8) * 255
        
        # 外円
        cv2.circle(img, (500, 500), 480, (0, 0, 0), 3)
        
        # メインエントリ
        cv2.circle(img, (500, 100), 30, (0, 0, 0), 2)
        cv2.circle(img, (500, 100), 25, (0, 0, 0), 2)
        
        # 多数のシンボルを配置
        for i in range(10):
            for j in range(10):
                x = 200 + i * 60
                y = 200 + j * 60
                if (x - 500)**2 + (y - 500)**2 < 400**2:  # 外円内
                    cv2.rectangle(img, (x-20, y-20), (x+20, y+20), (0, 0, 0), 2)
        
        return img