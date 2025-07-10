"""Enhanced error classes for Grimoire with position information"""

from typing import Optional, Tuple, Dict, Any
import json


class GrimoireError(Exception):
    """Base error class for all Grimoire errors"""
    
    def __init__(self, message: str, error_code: Optional[str] = None, 
                 position: Optional[Tuple[int, int]] = None,
                 symbol_type: Optional[str] = None,
                 context: Optional[Dict[str, Any]] = None):
        super().__init__(message)
        self.message = message
        self.error_code = error_code or "UNKNOWN_ERROR"
        self.position = position  # (x, y) coordinates
        self.symbol_type = symbol_type
        self.context = context or {}
        
    def __str__(self):
        """Format error message with position information"""
        parts = [self.message]
        
        if self.position:
            x, y = self.position
            parts.append(f"位置: ({x}, {y})")
            
        if self.symbol_type:
            parts.append(f"シンボル: {self.symbol_type}")
            
        if self.error_code != "UNKNOWN_ERROR":
            parts.append(f"エラーコード: {self.error_code}")
            
        return " | ".join(parts)
    
    def to_dict(self):
        """Convert error to dictionary for JSON serialization"""
        return {
            "error_type": self.__class__.__name__,
            "message": self.message,
            "error_code": self.error_code,
            "position": self.position,
            "symbol_type": self.symbol_type,
            "context": self.context
        }


class CompilationError(GrimoireError):
    """Compilation error with enhanced information"""
    
    def __init__(self, message: str, **kwargs):
        super().__init__(message, **kwargs)
        if not self.error_code.startswith("COMP_"):
            self.error_code = f"COMP_{self.error_code}"


class ParseError(GrimoireError):
    """Parse error with enhanced information"""
    
    def __init__(self, message: str, **kwargs):
        super().__init__(message, **kwargs)
        if not self.error_code.startswith("PARSE_"):
            self.error_code = f"PARSE_{self.error_code}"


class InterpreterError(GrimoireError):
    """Interpreter error with enhanced information"""
    
    def __init__(self, message: str, **kwargs):
        super().__init__(message, **kwargs)
        if not self.error_code.startswith("INTERP_"):
            self.error_code = f"INTERP_{self.error_code}"


class ImageRecognitionError(GrimoireError):
    """Image recognition error with enhanced information"""
    
    def __init__(self, message: str, **kwargs):
        super().__init__(message, **kwargs)
        if not self.error_code.startswith("IMG_"):
            self.error_code = f"IMG_{self.error_code}"


# Error codes
ERROR_CODES = {
    # Compilation errors
    "COMP_NO_MAGIC_CIRCLE": "魔法陣（外周円）が見つかりません",
    "COMP_NO_SYMBOLS": "シンボルが検出されませんでした",
    "COMP_INVALID_STRUCTURE": "魔法陣の構造が不正です",
    "COMP_PARSE_FAILED": "構文解析に失敗しました",
    "COMP_CODEGEN_FAILED": "コード生成に失敗しました",
    
    # Parse errors
    "PARSE_INVALID_OPERATOR": "不正な演算子です",
    "PARSE_INVALID_LITERAL": "不正なリテラル値です",
    "PARSE_INVALID_CONTROL": "不正な制御構造です",
    "PARSE_MISSING_CONDITION": "条件式が見つかりません",
    "PARSE_MISSING_BODY": "本体が見つかりません",
    "PARSE_CIRCULAR_REFERENCE": "循環参照が検出されました",
    
    # Interpreter errors
    "INTERP_UNDEFINED_VAR": "未定義の変数です",
    "INTERP_UNDEFINED_FUNC": "未定義の関数です",
    "INTERP_TYPE_ERROR": "型エラーです",
    "INTERP_DIVISION_BY_ZERO": "ゼロ除算エラーです",
    "INTERP_STACK_OVERFLOW": "スタックオーバーフローです",
    "INTERP_INDEX_ERROR": "インデックスエラーです",
    "INTERP_KEY_ERROR": "キーエラーです",
    
    # Image recognition errors
    "IMG_LOAD_FAILED": "画像の読み込みに失敗しました",
    "IMG_INVALID_FORMAT": "画像フォーマットが不正です",
    "IMG_NO_CIRCLE": "円が検出されませんでした",
    "IMG_AMBIGUOUS_SYMBOL": "シンボルの形状が曖昧です",
}


def get_error_message(error_code: str) -> str:
    """Get detailed error message for error code"""
    return ERROR_CODES.get(error_code, "不明なエラーです")


def format_error_with_suggestions(error: GrimoireError) -> str:
    """Format error with helpful suggestions"""
    lines = [
        f"\n🔴 エラーが発生しました: {error.__class__.__name__}",
        f"📍 {str(error)}",
    ]
    
    # Add detailed error message
    detailed_msg = get_error_message(error.error_code)
    if detailed_msg:
        lines.append(f"📝 詳細: {detailed_msg}")
    
    # Add suggestions based on error code
    suggestions = get_suggestions(error.error_code)
    if suggestions:
        lines.append("\n💡 解決方法:")
        for suggestion in suggestions:
            lines.append(f"   • {suggestion}")
    
    return "\n".join(lines)


def get_suggestions(error_code: str) -> list:
    """Get suggestions for fixing specific errors"""
    suggestions_map = {
        "COMP_NO_MAGIC_CIRCLE": [
            "画像に外周円を描いてください",
            "円が閉じていることを確認してください",
            "円の線が十分に太いことを確認してください"
        ],
        "COMP_NO_SYMBOLS": [
            "魔法陣の中にシンボルを配置してください",
            "シンボルの線が明確であることを確認してください",
            "背景と十分なコントラストがあることを確認してください"
        ],
        "PARSE_INVALID_OPERATOR": [
            "サポートされている演算子: +, -, *, /",
            "演算子シンボルの形状を確認してください"
        ],
        "INTERP_UNDEFINED_VAR": [
            "変数を使用する前に定義してください",
            "変数名のスペルを確認してください"
        ],
        "INTERP_DIVISION_BY_ZERO": [
            "除数が0でないことを確認してください",
            "条件分岐で0除算を回避してください"
        ],
        "IMG_LOAD_FAILED": [
            "ファイルパスが正しいことを確認してください",
            "画像ファイルが存在することを確認してください",
            "サポートされている形式: PNG, JPG, JPEG"
        ],
    }
    
    # Get specific suggestions or general ones based on error type
    if error_code in suggestions_map:
        return suggestions_map[error_code]
    elif error_code.startswith("COMP_"):
        return ["魔法陣の構造を確認してください", "画像が鮮明であることを確認してください"]
    elif error_code.startswith("PARSE_"):
        return ["シンボルの配置と接続を確認してください", "制御構造の形式を確認してください"]
    elif error_code.startswith("INTERP_"):
        return ["プログラムのロジックを確認してください", "デバッグモードで実行してみてください"]
    elif error_code.startswith("IMG_"):
        return ["画像の品質を確認してください", "別の画像形式で試してください"]
    
    return []