"""Tests for __main__ module"""

import pytest
import sys
from unittest.mock import patch, MagicMock
import subprocess
import os


class TestMain:
    """Test __main__ module"""
    
    @patch('grimoire.__main__.main')
    def test_main_module_execution(self, mock_main):
        """Test that __main__ calls main function"""
        mock_main.return_value = 0
        
        # Import the module (which will execute the code)
        import grimoire.__main__
        
        # Since the module is only executed when __name__ == "__main__",
        # we need to test it differently
        
    def test_main_as_module(self):
        """Test running grimoire as a module"""
        # Test that the module can be imported without errors
        import grimoire.__main__
        assert grimoire.__main__
    
    @patch('sys.exit')
    @patch('grimoire.cli.main')
    def test_main_execution_with_mock(self, mock_main, mock_exit):
        """Test __main__ execution flow"""
        mock_main.return_value = 0
        
        # Execute the __main__ module code
        exec_globals = {'__name__': '__main__'}
        with open('src/grimoire/__main__.py', 'r') as f:
            exec(f.read(), exec_globals)
        
        mock_main.assert_called_once()
        mock_exit.assert_called_once_with(0)
    
    @patch('sys.exit')
    @patch('grimoire.cli.main')
    def test_main_execution_with_error(self, mock_main, mock_exit):
        """Test __main__ execution with error"""
        mock_main.return_value = 1
        
        # Execute the __main__ module code
        exec_globals = {'__name__': '__main__'}
        with open('src/grimoire/__main__.py', 'r') as f:
            exec(f.read(), exec_globals)
        
        mock_main.assert_called_once()
        mock_exit.assert_called_once_with(1)
    
    def test_main_module_subprocess(self):
        """Test running grimoire as a module via subprocess"""
        # This test actually executes the module in a subprocess
        result = subprocess.run(
            [sys.executable, '-m', 'grimoire', '--help'],
            capture_output=True,
            text=True,
            cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
        )
        
        assert result.returncode == 0
        assert 'Grimoire - Visual Programming Language' in result.stdout
        assert 'compile' in result.stdout
        assert 'run' in result.stdout
        assert 'debug' in result.stdout
    
    def test_main_module_subprocess_compile(self):
        """Test running compile command via subprocess"""
        result = subprocess.run(
            [sys.executable, '-m', 'grimoire', 'compile', '--help'],
            capture_output=True,
            text=True,
            cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
        )
        
        assert result.returncode == 0
        assert 'Compile a magic circle image' in result.stdout
    
    def test_main_module_subprocess_run(self):
        """Test running run command via subprocess"""
        result = subprocess.run(
            [sys.executable, '-m', 'grimoire', 'run', '--help'],
            capture_output=True,
            text=True,
            cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
        )
        
        assert result.returncode == 0
        assert 'Run a magic circle image directly' in result.stdout
    
    def test_main_module_subprocess_debug(self):
        """Test running debug command via subprocess"""
        result = subprocess.run(
            [sys.executable, '-m', 'grimoire', 'debug', '--help'],
            capture_output=True,
            text=True,
            cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
        )
        
        assert result.returncode == 0
        assert 'Run in debug mode' in result.stdout
    
    def test_main_module_error_handling(self):
        """Test error handling in main module"""
        result = subprocess.run(
            [sys.executable, '-m', 'grimoire', 'compile', 'nonexistent.png'],
            capture_output=True,
            text=True,
            cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
        )
        
        assert result.returncode != 0
        assert 'Error' in result.stderr or 'does not exist' in result.stderr