"""Main compiler for Grimoire - Integrates all components"""

import os
from typing import Optional, Tuple, List
from pathlib import Path

from .image_recognition import MagicCircleDetector
from .parser import MagicCircleParser
from .interpreter import GrimoireInterpreter
from .ast_nodes import Program
from .errors import CompilationError, format_error_with_suggestions
from .ast_visualizer import ASTVisualizer, create_execution_trace


class GrimoireCompiler:
    """Main compiler that orchestrates the compilation process"""
    
    def __init__(self):
        self.detector = MagicCircleDetector()
        self.parser = MagicCircleParser()
        self.interpreter = GrimoireInterpreter()
        self.debug_mode = False
        self.errors: List[CompilationError] = []
        self.warnings = []
        self.current_image_path = None
        self.ast_visualizer = ASTVisualizer()
        self.execution_tracer = None
    
    def compile_and_run(self, image_path: str) -> str:
        """Compile and run a magic circle image"""
        self.current_image_path = image_path
        try:
            # Step 1: Detect symbols
            symbols, connections = self._detect_symbols(image_path)
            
            # Step 2: Parse to AST
            ast = self._parse(symbols, connections)
            
            # Step 3: Interpret
            result = self._interpret(ast)
            
            return result
            
        except CompilationError as e:
            self.errors.append(e)
            raise
        except Exception as e:
            error = CompilationError(
                f"予期しないエラーが発生しました: {str(e)}",
                error_code="COMP_UNEXPECTED"
            )
            self.errors.append(error)
            raise error
    
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
            
        except CompilationError as e:
            self.errors.append(e)
            raise
        except Exception as e:
            error = CompilationError(
                f"コンパイル中に予期しないエラー: {str(e)}",
                error_code="COMP_UNEXPECTED"
            )
            self.errors.append(error)
            raise error
    
    def debug(self, image_path: str) -> Tuple[Program, str]:
        """Debug mode - returns AST and execution trace"""
        self.debug_mode = True
        self.current_image_path = image_path
        
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
            self.execution_tracer = create_execution_trace()
            self.interpreter.set_tracer(self.execution_tracer)
            result = self._interpret(ast)
            
            return ast, result
            
        except CompilationError as e:
            self.errors.append(e)
            raise
        except Exception as e:
            error = CompilationError(
                f"デバッグ中に予期しないエラー: {str(e)}",
                error_code="COMP_DEBUG_ERROR"
            )
            self.errors.append(error)
            raise error
    
    def _detect_symbols(self, image_path: str):
        """Detect symbols from image"""
        if not os.path.exists(image_path):
            raise CompilationError(
                f"画像ファイルが見つかりません: {image_path}",
                error_code="IMG_LOAD_FAILED"
            )
        
        try:
            symbols, connections = self.detector.detect_symbols(image_path)
            
            if not symbols:
                raise CompilationError(
                    "画像からシンボルが検出されませんでした",
                    error_code="COMP_NO_SYMBOLS"
                )
            
            # Check for outer circle
            has_outer_circle = any(s.type.value == "outer_circle" for s in symbols)
            if not has_outer_circle:
                raise CompilationError(
                    "外周円が検出されませんでした。Grimoireプログラムは魔法陣（外周円）で囲まれている必要があります",
                    error_code="COMP_NO_MAGIC_CIRCLE"
                )
            
            return symbols, connections
            
        except CompilationError:
            raise
        except Exception as e:
            raise CompilationError(
                f"シンボル検出中にエラー: {str(e)}",
                error_code="IMG_DETECTION_FAILED"
            )
    
    def _parse(self, symbols, connections) -> Program:
        """Parse symbols to AST"""
        try:
            ast = self.parser.parse(symbols, connections)
            
            # Validate AST
            if not ast.has_outer_circle:
                raise CompilationError(
                    "ASTが不正です: 外周円がありません",
                    error_code="PARSE_INVALID_STRUCTURE"
                )
            
            return ast
            
        except CompilationError:
            raise
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
    
    try:
        ast, result = compiler.debug(image_path)
        
        # AST可視化
        print("\n=== AST構造 ===")
        print(compiler.ast_visualizer.visualize(ast))
        
        # 実行トレース
        if compiler.execution_tracer:
            print("\n=== 実行トレース ===")
            print(compiler.execution_tracer.format_trace())
        
        # 実行結果
        print("\n=== 実行結果 ===")
        print(result)
        
    except CompilationError as e:
        print(format_error_with_suggestions(e))
        
    # エラーと警告
    if compiler.errors:
        print("\n=== エラー ===")
        for error in compiler.errors:
            if isinstance(error, CompilationError):
                print(format_error_with_suggestions(error))
            else:
                print(f"  - {error}")
    
    if compiler.warnings:
        print("\n=== 警告 ===")
        for warning in compiler.warnings:
            print(f"  - {warning}")