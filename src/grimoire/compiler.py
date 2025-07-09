"""Main compiler for Grimoire - Integrates all components"""

import os
from typing import Optional, Tuple
from pathlib import Path

from .image_recognition import MagicCircleDetector
from .parser import MagicCircleParser
from .interpreter import GrimoireInterpreter
from .ast_nodes import Program


class CompilationError(Exception):
    """Compilation error"""
    pass


class GrimoireCompiler:
    """Main compiler that orchestrates the compilation process"""
    
    def __init__(self):
        self.detector = MagicCircleDetector()
        self.parser = MagicCircleParser()
        self.interpreter = GrimoireInterpreter()
        self.debug_mode = False
        self.errors = []
        self.warnings = []
    
    def compile_and_run(self, image_path: str) -> str:
        """Compile and run a magic circle image"""
        try:
            # Step 1: Detect symbols
            symbols, connections = self._detect_symbols(image_path)
            
            # Step 2: Parse to AST
            ast = self._parse(symbols, connections)
            
            # Step 3: Interpret
            result = self._interpret(ast)
            
            return result
            
        except Exception as e:
            self.errors.append(str(e))
            raise CompilationError(f"Compilation failed: {e}")
    
    def compile_to_python(self, image_path: str, output_path: Optional[str] = None) -> str:
        """Compile to Python code"""
        try:
            # Step 1: Detect symbols
            symbols, connections = self._detect_symbols(image_path)
            
            # Step 2: Parse to AST
            ast = self._parse(symbols, connections)
            
            # Step 3: Generate Python code
            python_code = self._generate_python(ast)
            
            # Save if output path provided
            if output_path:
                with open(output_path, 'w', encoding='utf-8') as f:
                    f.write(python_code)
            
            return python_code
            
        except Exception as e:
            self.errors.append(str(e))
            raise CompilationError(f"Compilation failed: {e}")
    
    def debug(self, image_path: str) -> Tuple[Program, str]:
        """Debug mode - returns AST and execution trace"""
        self.debug_mode = True
        
        try:
            # Step 1: Detect symbols
            symbols, connections = self._detect_symbols(image_path)
            
            if self.debug_mode:
                print(f"Detected {len(symbols)} symbols and {len(connections)} connections")
                for symbol in symbols:
                    print(f"  - {symbol.type.value} at {symbol.position}")
            
            # Step 2: Parse to AST
            ast = self._parse(symbols, connections)
            
            if self.debug_mode:
                print("\nAST structure:")
                print(f"  - Has outer circle: {ast.has_outer_circle}")
                print(f"  - Functions: {len(ast.functions)}")
                print(f"  - Global statements: {len(ast.globals)}")
            
            # Step 3: Interpret with trace
            result = self._interpret(ast)
            
            return ast, result
            
        except Exception as e:
            self.errors.append(str(e))
            raise CompilationError(f"Compilation failed: {e}")
    
    def _detect_symbols(self, image_path: str):
        """Detect symbols from image"""
        if not os.path.exists(image_path):
            raise CompilationError(f"Image file not found: {image_path}")
        
        try:
            symbols, connections = self.detector.detect_symbols(image_path)
            
            if not symbols:
                raise CompilationError("No symbols detected in image")
            
            # Check for outer circle
            has_outer_circle = any(s.type.value == "outer_circle" for s in symbols)
            if not has_outer_circle:
                raise CompilationError("No outer circle detected. All Grimoire programs must be enclosed in a magic circle.")
            
            return symbols, connections
            
        except Exception as e:
            raise CompilationError(f"Symbol detection failed: {e}")
    
    def _parse(self, symbols, connections) -> Program:
        """Parse symbols to AST"""
        try:
            ast = self.parser.parse(symbols, connections)
            
            # Validate AST
            if not ast.has_outer_circle:
                raise CompilationError("Invalid AST: no outer circle")
            
            return ast
            
        except Exception as e:
            raise CompilationError(f"Parsing failed: {e}")
    
    def _interpret(self, ast: Program) -> str:
        """Interpret AST"""
        try:
            result = self.interpreter.interpret(ast)
            return result
            
        except Exception as e:
            raise CompilationError(f"Runtime error: {e}")
    
    def _generate_python(self, ast: Program) -> str:
        """Generate Python code from AST"""
        # Use the code generator
        from .code_generator import PythonCodeGenerator
        generator = PythonCodeGenerator()
        return generator.generate(ast)


# Convenience functions for CLI

def compile_grimoire(image_path: str, output_path: Optional[str] = None) -> str:
    """Compile a Grimoire image"""
    compiler = GrimoireCompiler()
    
    if output_path:
        # Generate executable
        python_code = compiler.compile_to_python(image_path, None)
        
        # Create executable wrapper
        if os.name == 'nt':  # Windows
            # Create batch file
            with open(output_path + '.bat', 'w') as f:
                f.write(f'@echo off\npython -c "{python_code}"\n')
            return f"Compilation complete: {output_path}.bat"
        else:
            # Create shell script
            with open(output_path, 'w') as f:
                f.write(f'#!/usr/bin/env python3\n{python_code}\n')
            os.chmod(output_path, 0o755)
            return f"Compilation complete: {output_path}"
    else:
        # Just return the result
        return compiler.compile_and_run(image_path)


def run_grimoire(image_path: str) -> str:
    """Run a Grimoire image directly"""
    compiler = GrimoireCompiler()
    return compiler.compile_and_run(image_path)


def debug_grimoire(image_path: str) -> None:
    """Debug a Grimoire image"""
    compiler = GrimoireCompiler()
    ast, result = compiler.debug(image_path)
    
    print("\n=== Execution Result ===")
    print(result)
    
    if compiler.errors:
        print("\n=== Errors ===")
        for error in compiler.errors:
            print(f"  - {error}")
    
    if compiler.warnings:
        print("\n=== Warnings ===")
        for warning in compiler.warnings:
            print(f"  - {warning}")