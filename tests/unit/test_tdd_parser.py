"""
TDD approach for parser module
Following t-wada's testing principles
"""

import pytest
from grimoire.parser import MagicCircleParser, ParseError
from grimoire.image_recognition import Symbol, SymbolType
from grimoire.ast_nodes import (
    Program, FunctionDef, OutputStatement, Assignment,
    BinaryOp, OperatorType, Literal, DataType, IfStatement,
    WhileLoop
)


class TestParserBasicRequirements:
    """パーサーの基本要件テスト"""
    
    @pytest.fixture
    def parser(self):
        return MagicCircleParser()
    
    def test_外円なしプログラムはパースエラー(self, parser):
        # Arrange
        symbols = [
            Symbol(
                type=SymbolType.DOUBLE_CIRCLE,
                position=(250, 250),
                size=30,
                confidence=0.9,
                properties={}
            )
        ]
        
        # Act & Assert
        with pytest.raises(ParseError, match="外周円が見つかりません"):
            parser.parse(symbols, [])
    
    def test_最小限のプログラムがパースできる(self, parser):
        # Arrange
        symbols = [
            Symbol(
                type=SymbolType.OUTER_CIRCLE,
                position=(250, 250),
                size=200,
                confidence=1.0,
                properties={}
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        assert isinstance(ast, Program)
        assert ast.has_outer_circle is True
        assert ast.main_entry is None  # メインエントリなし
        assert len(ast.functions) == 0
        assert len(ast.globals) == 0
    
    def test_メインエントリが正しく認識される(self, parser):
        # Arrange
        symbols = [
            Symbol(
                type=SymbolType.OUTER_CIRCLE,
                position=(250, 250),
                size=200,
                confidence=1.0,
                properties={}
            ),
            Symbol(
                type=SymbolType.DOUBLE_CIRCLE,
                position=(250, 150),
                size=30,
                confidence=0.9,
                properties={}
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        assert ast.main_entry is not None
        assert isinstance(ast.main_entry, FunctionDef)
        assert ast.main_entry.is_main is True


class TestExpressionParsing:
    """式のパーステスト"""
    
    @pytest.fixture
    def parser(self):
        return MagicCircleParser()
    
    @pytest.fixture
    def base_symbols(self):
        return [
            Symbol(
                type=SymbolType.OUTER_CIRCLE,
                position=(300, 300),
                size=280,
                confidence=1.0,
                properties={}
            ),
            Symbol(
                type=SymbolType.DOUBLE_CIRCLE,
                position=(300, 100),
                size=30,
                confidence=0.9,
                properties={}
            )
        ]
    
    def test_数値リテラルがパースできる(self, parser, base_symbols):
        # Arrange
        symbols = base_symbols + [
            Symbol(
                type=SymbolType.SQUARE,
                position=(200, 200),
                size=40,
                confidence=0.9,
                properties={"pattern": "dot"}  # 1
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        # パーサーの実装により詳細なアサーションが必要
        assert ast is not None
    
    def test_加算演算がパースできる(self, parser, base_symbols):
        # Arrange
        symbols = base_symbols + [
            Symbol(
                type=SymbolType.SQUARE,
                position=(200, 200),
                size=40,
                confidence=0.9,
                properties={"pattern": "dot"}  # 1
            ),
            Symbol(
                type=SymbolType.SQUARE,
                position=(400, 200),
                size=40,
                confidence=0.9,
                properties={"pattern": "double_dot"}  # 2
            ),
            Symbol(
                type=SymbolType.CONVERGENCE,  # 加算
                position=(300, 200),
                size=20,
                confidence=0.8,
                properties={}
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        assert ast is not None
        # 実装により詳細なアサーションが必要


class TestControlFlowParsing:
    """制御フローのパーステスト"""
    
    @pytest.fixture
    def parser(self):
        return MagicCircleParser()
    
    @pytest.fixture
    def base_symbols(self):
        return [
            Symbol(
                type=SymbolType.OUTER_CIRCLE,
                position=(300, 300),
                size=280,
                confidence=1.0,
                properties={}
            ),
            Symbol(
                type=SymbolType.DOUBLE_CIRCLE,
                position=(300, 100),
                size=30,
                confidence=0.9,
                properties={}
            )
        ]
    
    def test_条件分岐がパースできる(self, parser, base_symbols):
        # Arrange
        symbols = base_symbols + [
            Symbol(
                type=SymbolType.TRIANGLE,  # if文
                position=(300, 200),
                size=50,
                confidence=0.8,
                properties={}
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        assert ast is not None
    
    def test_ループがパースできる(self, parser, base_symbols):
        # Arrange
        symbols = base_symbols + [
            Symbol(
                type=SymbolType.PENTAGON,  # ループ
                position=(300, 200),
                size=60,
                confidence=0.8,
                properties={}
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        assert ast is not None


class TestConnectionInference:
    """接続推論のテスト"""
    
    @pytest.fixture
    def parser(self):
        return MagicCircleParser()
    
    def test_接続なしでも空間配置から推論される(self, parser):
        # Arrange
        symbols = [
            Symbol(
                type=SymbolType.OUTER_CIRCLE,
                position=(300, 300),
                size=280,
                confidence=1.0,
                properties={}
            ),
            Symbol(
                type=SymbolType.DOUBLE_CIRCLE,
                position=(300, 100),
                size=30,
                confidence=0.9,
                properties={}
            ),
            Symbol(
                type=SymbolType.STAR,  # 出力
                position=(300, 200),
                size=40,
                confidence=0.9,
                properties={}
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        assert ast.main_entry is not None
        # 接続推論により、スターがメイン関数の子になるはず
    
    def test_演算子の両側のオペランドが推論される(self, parser):
        # Arrange
        symbols = [
            Symbol(
                type=SymbolType.OUTER_CIRCLE,
                position=(300, 300),
                size=280,
                confidence=1.0,
                properties={}
            ),
            Symbol(
                type=SymbolType.CONVERGENCE,
                position=(300, 200),
                size=20,
                confidence=0.8,
                properties={}
            ),
            Symbol(
                type=SymbolType.SQUARE,
                position=(200, 200),  # 左側
                size=40,
                confidence=0.9,
                properties={"pattern": "dot"}
            ),
            Symbol(
                type=SymbolType.SQUARE,
                position=(400, 200),  # 右側
                size=40,
                confidence=0.9,
                properties={"pattern": "double_dot"}
            )
        ]
        
        # Act
        ast = parser.parse(symbols, [])
        
        # Assert
        assert ast is not None


class TestErrorCases:
    """エラーケースのテスト"""
    
    @pytest.fixture
    def parser(self):
        return MagicCircleParser()
    
    def test_空のシンボルリストはエラー(self, parser):
        # Arrange
        symbols = []
        
        # Act & Assert
        with pytest.raises(ParseError):
            parser.parse(symbols, [])
    
    def test_重複する関数定義はエラー(self, parser):
        # これは実装により詳細化が必要
        pass


class TestPropertyBasedTesting:
    """プロパティベーステスト"""
    
    @pytest.fixture
    def parser(self):
        return MagicCircleParser()
    
    def test_任意の有効なシンボル配置でクラッシュしない(self, parser):
        # hypothesis を使った場合
        # @given(valid_symbols())
        # def test_no_crash(symbols):
        #     try:
        #         parser.parse(symbols, [])
        #     except ParseError:
        #         pass  # ParseErrorは許容
        pass