"""Tests for CLI module"""

import pytest
from click.testing import CliRunner
from pathlib import Path
import tempfile
from unittest.mock import patch, MagicMock
from grimoire.cli import cli, compile, run, debug, main


class TestCLI:
    """Test CLI commands"""
    
    def setup_method(self):
        """Set up test fixtures"""
        self.runner = CliRunner()
        
    def test_cli_help(self):
        """Test CLI help command"""
        result = self.runner.invoke(cli, ['--help'])
        assert result.exit_code == 0
        assert 'Grimoire - Visual Programming Language' in result.output
        assert 'compile' in result.output
        assert 'run' in result.output
        assert 'debug' in result.output
    
    def test_compile_help(self):
        """Test compile command help"""
        result = self.runner.invoke(cli, ['compile', '--help'])
        assert result.exit_code == 0
        assert 'Compile a magic circle image' in result.output
        assert '--output' in result.output
    
    def test_run_help(self):
        """Test run command help"""
        result = self.runner.invoke(cli, ['run', '--help'])
        assert result.exit_code == 0
        assert 'Run a magic circle image directly' in result.output
    
    def test_debug_help(self):
        """Test debug command help"""
        result = self.runner.invoke(cli, ['debug', '--help'])
        assert result.exit_code == 0
        assert 'Run in debug mode' in result.output
    
    @patch('grimoire.cli.compile_grimoire')
    def test_compile_success(self, mock_compile):
        """Test successful compilation"""
        mock_compile.return_value = "print('Hello, World!')"
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            result = self.runner.invoke(cli, ['compile', tmp.name])
            assert result.exit_code == 0
            assert "print('Hello, World!')" in result.output
            mock_compile.assert_called_once_with(tmp.name, None)
    
    @patch('grimoire.cli.compile_grimoire')
    def test_compile_with_output(self, mock_compile):
        """Test compilation with output file"""
        mock_compile.return_value = "print('Hello, World!')"
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            result = self.runner.invoke(cli, ['compile', tmp.name, '-o', 'output.py'])
            assert result.exit_code == 0
            assert "Compilation complete: output.py" in result.output
            mock_compile.assert_called_once_with(tmp.name, 'output.py')
    
    @patch('grimoire.cli.compile_grimoire')
    def test_compile_error(self, mock_compile):
        """Test compilation error handling"""
        mock_compile.side_effect = Exception("Compilation failed")
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            result = self.runner.invoke(cli, ['compile', tmp.name])
            assert result.exit_code == 1
            assert "🔴 エラー: Compilation failed" in result.output
    
    @patch('grimoire.cli.run_grimoire')
    def test_run_success(self, mock_run):
        """Test successful run"""
        mock_run.return_value = "42"
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            result = self.runner.invoke(cli, ['run', tmp.name])
            assert result.exit_code == 0
            assert "42" in result.output
            mock_run.assert_called_once_with(tmp.name)
    
    @patch('grimoire.cli.run_grimoire')
    def test_run_error(self, mock_run):
        """Test run error handling"""
        mock_run.side_effect = Exception("Runtime error")
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            result = self.runner.invoke(cli, ['run', tmp.name])
            assert result.exit_code == 1
            assert "🔴 エラー: Runtime error" in result.output
    
    @patch('grimoire.cli.debug_grimoire')
    def test_debug_success(self, mock_debug):
        """Test successful debug"""
        mock_debug.return_value = None
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            result = self.runner.invoke(cli, ['debug', tmp.name])
            assert result.exit_code == 0
            mock_debug.assert_called_once_with(tmp.name)
    
    @patch('grimoire.cli.debug_grimoire')
    def test_debug_error(self, mock_debug):
        """Test debug error handling"""
        mock_debug.side_effect = Exception("Debug error")
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            result = self.runner.invoke(cli, ['debug', tmp.name])
            assert result.exit_code == 1
            assert "🔴 エラー: Debug error" in result.output
    
    def test_file_not_exists(self):
        """Test with non-existent file"""
        result = self.runner.invoke(cli, ['compile', 'nonexistent.png'])
        assert result.exit_code == 2
        assert "does not exist" in result.output or "No such file" in result.output
    
    def test_no_arguments(self):
        """Test CLI with no arguments"""
        result = self.runner.invoke(cli, [])
        assert result.exit_code == 0
        assert 'Grimoire - Visual Programming Language' in result.output
    
    @patch('grimoire.cli.cli')
    def test_main_function(self, mock_cli):
        """Test main entry point"""
        main()
        mock_cli.assert_called_once()


class TestCLIIntegration:
    """Integration tests for CLI"""
    
    def setup_method(self):
        """Set up test fixtures"""
        self.runner = CliRunner()
    
    @patch('grimoire.compiler.GrimoireCompiler')
    def test_compile_integration(self, mock_compiler_class):
        """Test compile command integration"""
        mock_compiler = MagicMock()
        mock_compiler.compile_and_run.return_value = "print('Hello')"
        mock_compiler_class.return_value = mock_compiler
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            # Create a dummy image file
            tmp.write(b'\x89PNG\r\n\x1a\n')
            tmp.flush()
            
            result = self.runner.invoke(cli, ['compile', tmp.name])
            assert result.exit_code == 0
            # Check that result was returned
            assert result.exit_code == 0
    
    @patch('grimoire.compiler.GrimoireCompiler')
    def test_run_integration(self, mock_compiler_class):
        """Test run command integration"""
        mock_compiler = MagicMock()
        mock_compiler.compile.return_value = "print('42')"
        mock_compiler_class.return_value = mock_compiler
        
        with tempfile.NamedTemporaryFile(suffix='.png') as tmp:
            # Create a dummy image file
            tmp.write(b'\x89PNG\r\n\x1a\n')
            tmp.flush()
            
            result = self.runner.invoke(cli, ['run', tmp.name])
            assert result.exit_code == 0
            # The actual output depends on execution