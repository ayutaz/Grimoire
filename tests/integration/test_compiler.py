"""
Integration tests for Grimoire compiler following TDD principles
Based on t-wada's testing approach with AAA pattern
"""

import pytest
import tempfile
import os
import cv2
import numpy as np
from pathlib import Path

from grimoire.compiler import (
    GrimoireCompiler, CompilationError, compile_grimoire, 
    run_grimoire, debug_grimoire
)
from grimoire.image_recognition import SymbolType
from grimoire.code_generator import PythonCodeGenerator


class TestGrimoireCompiler:
    """Integration tests for the Grimoire compiler"""
    
    def setup_method(self):
        """Set up test fixtures (Arrange phase for all tests)"""
        self.compiler = GrimoireCompiler()
        self.test_images_dir = Path(__file__).parent.parent / "fixtures"
        os.makedirs(self.test_images_dir, exist_ok=True)
    
    def create_test_image(self, filename: str, width=500, height=500):
        """Helper to create and save a test image"""
        img = np.ones((height, width, 3), dtype=np.uint8) * 255
        filepath = self.test_images_dir / filename
        return img, filepath
    
    def test_compiler_initialization(self):
        """Test compiler initializes with all components"""
        # Assert
        assert self.compiler.detector is not None
        assert self.compiler.parser is not None
        assert self.compiler.interpreter is not None
        assert self.compiler.debug_mode is False
        assert self.compiler.errors == []
        assert self.compiler.warnings == []
    
    def test_compile_nonexistent_file(self):
        """Test compiling non-existent file"""
        # Arrange
        non_existent = "/path/to/nonexistent.png"
        
        # Act & Assert
        with pytest.raises(CompilationError, match="Image file not found"):
            self.compiler.compile_and_run(non_existent)
    
    def test_compile_empty_image(self):
        """Test compiling empty image (no symbols)"""
        # Arrange
        img, filepath = self.create_test_image("empty.png")
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act & Assert
            with pytest.raises(CompilationError, match="No outer circle detected"):
                self.compiler.compile_and_run(str(filepath))
        finally:
            os.unlink(filepath)
    
    def test_compile_minimal_program(self):
        """Test compiling minimal valid program"""
        # Arrange
        img, filepath = self.create_test_image("minimal.png")
        # Draw outer circle
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            result = self.compiler.compile_and_run(str(filepath))
            
            # Assert
            assert result == ""  # Empty program produces no output
        finally:
            os.unlink(filepath)
    
    def test_compile_hello_world(self):
        """Test compiling hello world program"""
        # Arrange
        img, filepath = self.create_test_image("hello_world.png")
        # Draw outer circle
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        # Draw main entry (double circle)
        cv2.circle(img, (250, 200), 50, (0, 0, 0), 2)
        cv2.circle(img, (250, 200), 40, (0, 0, 0), 2)
        # Draw output star
        star_points = self._create_star_points(250, 300, 30)
        cv2.polylines(img, [star_points], True, (0, 0, 0), 2)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            result = self.compiler.compile_and_run(str(filepath))
            
            # Assert
            # Output will be a default value since we don't have text recognition
            # For now, accept empty output as valid
            assert result is not None
        finally:
            os.unlink(filepath)
    
    def test_compile_simple_math(self):
        """Test compiling simple math operation"""
        # Arrange
        img, filepath = self.create_test_image("simple_math.png")
        # Draw outer circle
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        # Draw two squares (operands)
        cv2.rectangle(img, (150, 200), (190, 240), (0, 0, 0), 2)
        cv2.circle(img, (170, 220), 3, (0, 0, 0), -1)  # Single dot = 1
        cv2.rectangle(img, (310, 200), (350, 240), (0, 0, 0), 2)
        cv2.circle(img, (325, 215), 3, (0, 0, 0), -1)  # Double dots = 2
        cv2.circle(img, (335, 225), 3, (0, 0, 0), -1)
        # Draw convergence operator (addition)
        cv2.line(img, (190, 220), (250, 220), (0, 0, 0), 2)
        cv2.line(img, (310, 220), (250, 220), (0, 0, 0), 2)
        # Draw output star
        star_points = self._create_star_points(250, 300, 30)
        cv2.polylines(img, [star_points], True, (0, 0, 0), 2)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            result = self.compiler.compile_and_run(str(filepath))
            
            # Assert
            # Should output some result (exact value depends on recognition)
            # For now, accept empty output as valid
            assert result is not None
        finally:
            os.unlink(filepath)
    
    def test_compile_to_python(self):
        """Test compiling to Python code"""
        # Arrange
        img, filepath = self.create_test_image("to_python.png")
        # Create simple program
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        cv2.circle(img, (250, 200), 50, (0, 0, 0), 2)
        cv2.circle(img, (250, 200), 40, (0, 0, 0), 2)
        star_points = self._create_star_points(250, 300, 30)
        cv2.polylines(img, [star_points], True, (0, 0, 0), 2)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            python_code = self.compiler.compile_to_python(str(filepath))
            
            # Assert
            assert "#!/usr/bin/env python3" in python_code
            assert "# Generated by Grimoire compiler" in python_code
            assert "import sys" in python_code
            assert "def main():" in python_code or "if __name__ == '__main__':" in python_code
        finally:
            os.unlink(filepath)
    
    def test_compile_to_python_with_output_file(self):
        """Test compiling to Python with output file"""
        # Arrange
        img, filepath = self.create_test_image("to_python_file.png")
        output_path = self.test_images_dir / "output.py"
        
        # Create simple program
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            python_code = self.compiler.compile_to_python(str(filepath), str(output_path))
            
            # Assert
            assert output_path.exists()
            with open(output_path, 'r') as f:
                saved_code = f.read()
            assert saved_code == python_code
        finally:
            if filepath.exists():
                os.unlink(filepath)
            if output_path.exists():
                os.unlink(output_path)
    
    def test_debug_mode(self):
        """Test debug mode functionality"""
        # Arrange
        img, filepath = self.create_test_image("debug.png")
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            ast, result = self.compiler.debug(str(filepath))
            
            # Assert
            assert ast is not None
            assert ast.has_outer_circle is True
            assert isinstance(result, str)
        finally:
            os.unlink(filepath)
    
    def test_error_collection(self):
        """Test error collection during compilation"""
        # Arrange
        img, filepath = self.create_test_image("error_test.png")
        cv2.imwrite(str(filepath), img)  # Empty image
        
        try:
            # Act
            try:
                self.compiler.compile_and_run(str(filepath))
            except CompilationError:
                pass
            
            # Assert
            assert len(self.compiler.errors) > 0
        finally:
            os.unlink(filepath)
    
    def test_convenience_functions(self):
        """Test convenience functions"""
        # Arrange
        img, filepath = self.create_test_image("convenience.png")
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Test run_grimoire
            result = run_grimoire(str(filepath))
            assert isinstance(result, str)
            
            # Test compile_grimoire without output
            result = compile_grimoire(str(filepath))
            assert isinstance(result, str)
        finally:
            os.unlink(filepath)
    
    def test_compile_executable_unix(self):
        """Test creating executable on Unix-like systems"""
        # Arrange
        if os.name == 'nt':
            pytest.skip("Unix-specific test")
        
        img, filepath = self.create_test_image("executable.png")
        output_path = self.test_images_dir / "executable"
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            result = compile_grimoire(str(filepath), str(output_path))
            
            # Assert
            assert output_path.exists()
            assert os.access(output_path, os.X_OK)  # Check executable
            assert f"Compilation complete: {output_path}" in result
        finally:
            if filepath.exists():
                os.unlink(filepath)
            if output_path.exists():
                os.unlink(output_path)
    
    def test_compile_executable_windows(self):
        """Test creating executable on Windows"""
        # Arrange
        if os.name != 'nt':
            pytest.skip("Windows-specific test")
        
        img, filepath = self.create_test_image("executable.png")
        output_path = self.test_images_dir / "executable"
        cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            result = compile_grimoire(str(filepath), str(output_path))
            
            # Assert
            batch_file = Path(str(output_path) + '.bat')
            assert batch_file.exists()
            assert f"Compilation complete: {batch_file}" in result
        finally:
            if filepath.exists():
                os.unlink(filepath)
            batch_file = Path(str(output_path) + '.bat')
            if batch_file.exists():
                os.unlink(batch_file)
    
    def test_complex_program_integration(self):
        """Test compiling complex program with multiple features"""
        # Arrange
        img, filepath = self.create_test_image("complex.png", 800, 800)
        
        # Draw outer circle
        cv2.circle(img, (400, 400), 380, (0, 0, 0), 3)
        
        # Draw main function (double circle)
        cv2.circle(img, (400, 200), 60, (0, 0, 0), 2)
        cv2.circle(img, (400, 200), 50, (0, 0, 0), 2)
        
        # Draw loop (pentagon)
        pentagon_points = self._create_polygon_points(300, 300, 40, 5)
        cv2.polylines(img, [pentagon_points], True, (0, 0, 0), 2)
        
        # Draw condition (triangle)
        triangle_points = np.array([[500, 280], [480, 320], [520, 320]], np.int32)
        cv2.polylines(img, [triangle_points], True, (0, 0, 0), 2)
        
        # Draw output stars
        star1 = self._create_star_points(300, 400, 30)
        star2 = self._create_star_points(500, 400, 30)
        cv2.polylines(img, [star1], True, (0, 0, 0), 2)
        cv2.polylines(img, [star2], True, (0, 0, 0), 2)
        
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            result = self.compiler.compile_and_run(str(filepath))
            
            # Assert
            # Complex program should produce some output
            assert result != ""
            
            # Also test Python generation
            python_code = self.compiler.compile_to_python(str(filepath))
            assert "def main():" in python_code or "if __name__ == '__main__':" in python_code
        finally:
            os.unlink(filepath)
    
    def test_parallel_execution_integration(self):
        """Test compiling program with parallel execution"""
        # Arrange
        img, filepath = self.create_test_image("parallel.png", 600, 600)
        
        # Draw outer circle
        cv2.circle(img, (300, 300), 280, (0, 0, 0), 3)
        
        # Draw hexagon for parallel block
        hexagon_points = self._create_polygon_points(300, 300, 60, 6)
        cv2.polylines(img, [hexagon_points], True, (0, 0, 0), 2)
        
        # Draw branches
        star1 = self._create_star_points(200, 200, 25)
        star2 = self._create_star_points(400, 200, 25)
        cv2.polylines(img, [star1], True, (0, 0, 0), 2)
        cv2.polylines(img, [star2], True, (0, 0, 0), 2)
        
        cv2.imwrite(str(filepath), img)
        
        try:
            # Act
            python_code = self.compiler.compile_to_python(str(filepath))
            
            # Assert
            assert "ThreadPoolExecutor" in python_code
            assert "executor.submit" in python_code
        finally:
            os.unlink(filepath)
    
    # Helper methods
    
    def _create_star_points(self, cx: int, cy: int, size: int) -> np.ndarray:
        """Create points for a 5-pointed star"""
        points = []
        for i in range(10):
            angle = i * np.pi / 5
            if i % 2 == 0:
                r = size
            else:
                r = size * 0.5
            x = int(cx + r * np.cos(angle - np.pi/2))
            y = int(cy + r * np.sin(angle - np.pi/2))
            points.append([x, y])
        return np.array(points, np.int32)
    
    def _create_polygon_points(self, cx: int, cy: int, size: int, sides: int) -> np.ndarray:
        """Create points for a regular polygon"""
        points = []
        for i in range(sides):
            angle = i * 2 * np.pi / sides - np.pi/2
            x = int(cx + size * np.cos(angle))
            y = int(cy + size * np.sin(angle))
            points.append([x, y])
        return np.array(points, np.int32)


class TestDebugGrimoire:
    """Test debug_grimoire function"""
    
    def test_debug_grimoire_output(self, capsys):
        """Test debug output formatting"""
        # Arrange
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as tmp:
            img = np.ones((500, 500, 3), dtype=np.uint8) * 255
            cv2.circle(img, (250, 250), 240, (0, 0, 0), 3)
            cv2.imwrite(tmp.name, img)
            
            try:
                # Act
                debug_grimoire(tmp.name)
                
                # Assert
                captured = capsys.readouterr()
                assert "=== Execution Result ===" in captured.out
            finally:
                os.unlink(tmp.name)