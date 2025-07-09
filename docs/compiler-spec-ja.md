# Grimoireコンパイラアーキテクチャ仕様

## 目次
1. [概要](#overview)
2. [コンパイルパイプライン](#compilation-pipeline)
3. [画像認識ステージ](#image-recognition)
4. [シンボル認識](#symbol-recognition)
5. [トポロジー解析](#topology-analysis)
6. [AST生成](#ast-generation)
7. [型推論](#type-inference)
8. [コード生成](#code-generation)
9. [最適化](#optimization)
10. [エラー処理](#error-handling)
11. [デバッグサポート](#debug-support)

## 概要 {#overview}

Grimoireコンパイラは、純粋にシンボルベースの魔法陣を実行可能プログラムに変換します。コンピュータビジョン、パターン認識、従来のコンパイル技術を組み合わせた多段階パイプラインを使用します。テキストは一切使用せず、図形の形状、内部パターン、配置のみで全ての意味を表現します。

### コンパイラコンポーネント
```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│画像入力     │ --> │CV処理        │ --> │シンボル認識  │
└─────────────┘     └──────────────┘     └──────────────┘
                           |
                           v
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│コード生成   │ <-- │型推論        │ <-- │ASTビルダー   │
└─────────────┘     └──────────────┘     └──────────────┘
```

## コンパイルパイプライン {#compilation-pipeline}

### ステージ1: 画像前処理
1. 画像読み込み（PNG/JPG/SVG）
2. ノイズ除去
3. コントラスト強調
4. 二値化処理
5. エッジ検出

### ステージ2: シンボル認識
1. 輪郭検出と図形分類
2. 内部パターン認識
3. ドット配列解析（数値用）
4. 演算子シンボルマッチング

### ステージ3: 意味解析
1. トポロジーグラフ構築
2. シンボル関係マッピング
3. フロー方向検出
4. スコープ階層構築

### ステージ4: コード生成
1. AST構築
2. 型推論
3. 最適化パス
4. ターゲットコード出力

## 画像認識ステージ {#image-recognition}

### 前処理パイプライン
```python
def preprocess_image(image):
    # 1. グレースケール変換
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    
    # 2. ガウシアンブラーでノイズ除去
    blurred = cv2.GaussianBlur(gray, (5, 5), 0)
    
    # 3. 適応的二値化で照明変化に対応
    binary = cv2.adaptiveThreshold(blurred, 255, 
                                  cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                                  cv2.THRESH_BINARY, 11, 2)
    
    # 4. モルフォロジー演算
    kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (3, 3))
    cleaned = cv2.morphologyEx(binary, cv2.MORPH_CLOSE, kernel)
    
    return cleaned
```

### 図形検出アルゴリズム
```python
def detect_shapes(contours):
    shapes = []
    for contour in contours:
        # 多角形近似
        epsilon = 0.02 * cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, epsilon, True)
        
        # 頂点数で分類
        vertices = len(approx)
        if vertices == 3:
            shape_type = "triangle"  # 条件分岐
        elif vertices == 4:
            shape_type = "square"    # 変数/データ
        elif vertices == 5:
            shape_type = "pentagon"  # ループ
        elif vertices == 6:
            shape_type = "hexagon"   # 並列処理
        elif vertices > 6:
            # 円判定
            area = cv2.contourArea(contour)
            perimeter = cv2.arcLength(contour, True)
            circularity = 4 * np.pi * area / (perimeter ** 2)
            if circularity > 0.8:
                shape_type = "circle"  # 関数/スコープ
            else:
                shape_type = "star" if self.is_star(approx) else "polygon"
                
        shapes.append(Shape(shape_type, contour, approx))
    return shapes
```

## シンボル認識 {#symbol-recognition}

### 純粋シンボル認識パイプライン

#### 1. 基本図形の意味
- **円**: 関数定義、スコープ境界
- **四角形**: 変数、データ格納
- **三角形**: 条件分岐、フロー制御
- **五角形**: ループ構造
- **六角形**: 並列処理、非同期操作
- **星形**: 外部参照、インポート

#### 2. 内部パターン認識
```python
class InternalPatternRecognizer:
    def recognize_pattern(self, shape_image):
        # 図形内部のパターンを分析
        internal_region = self.extract_internal_region(shape_image)
        
        # パターンタイプを判定
        if self.has_dots(internal_region):
            return self.parse_dot_pattern(internal_region)
        elif self.has_geometric_pattern(internal_region):
            return self.parse_geometric_pattern(internal_region)
        elif self.has_line_pattern(internal_region):
            return self.parse_line_pattern(internal_region)
        else:
            return "empty"  # 空のシンボル
```

#### 3. ドット配列による数値表現
```python
def parse_dot_pattern(dot_region):
    """3x3グリッドのドット配列を数値に変換"""
    grid = self.extract_3x3_grid(dot_region)
    
    # 各セルのドット有無をビットとして解釈
    value = 0
    for row in range(3):
        for col in range(3):
            if grid[row][col]:  # ドットが存在
                bit_position = row * 3 + col
                value |= (1 << bit_position)
    
    return {"type": "number", "value": value}
```

#### 4. 演算子シンボル認識
```python
class OperatorRecognizer:
    def __init__(self):
        self.operators = {
            "plus": self.create_plus_pattern(),      # +形状
            "minus": self.create_minus_pattern(),    # -形状
            "multiply": self.create_x_pattern(),     # ×形状
            "divide": self.create_slash_pattern(),   # /形状
            "equal": self.create_equal_pattern(),    # =形状
            "greater": self.create_gt_pattern(),     # >形状
            "less": self.create_lt_pattern(),        # <形状
            "and": self.create_and_pattern(),        # ∧形状
            "or": self.create_or_pattern(),          # ∨形状
            "not": self.create_not_pattern()         # ¬形状
        }
    
    def recognize_operator(self, symbol_region):
        for op_name, pattern in self.operators.items():
            if self.match_pattern(symbol_region, pattern):
                return {"type": "operator", "operation": op_name}
        return None
```

#### 5. データ型パターン
```python
def recognize_type_pattern(shape):
    """図形の輪郭スタイルから型を判定"""
    contour_style = self.analyze_contour_style(shape)
    
    if contour_style == "solid":
        return "integer"
    elif contour_style == "dashed":
        return "float"
    elif contour_style == "dotted":
        return "string"
    elif contour_style == "double":
        return "boolean"
    elif contour_style == "wavy":
        return "array"
    else:
        return "any"
```

## トポロジー解析 {#topology-analysis}

### グラフ構築
```python
class TopologyGraph:
    def __init__(self):
        self.nodes = {}  # shape_id -> Shape
        self.edges = {}  # (shape1_id, shape2_id) -> Connection
    
    def build_from_shapes(self, shapes):
        # 1. ノード作成
        for shape in shapes:
            self.nodes[shape.id] = shape
        
        # 2. 接続検出
        for shape1 in shapes:
            for shape2 in shapes:
                if shape1.id != shape2.id:
                    connection = self.detect_connection(shape1, shape2)
                    if connection:
                        self.edges[(shape1.id, shape2.id)] = connection
```

### 接続タイプとフロー方向
```python
def detect_connection_type(shape1, shape2):
    """接続線のパターンから接続タイプを判定"""
    line = self.extract_connecting_line(shape1, shape2)
    
    if not line:
        return None
    
    # 線のスタイルを分析
    if self.is_arrow_line(line):
        return {"type": "directional", "direction": self.get_arrow_direction(line)}
    elif self.is_dashed_line(line):
        return {"type": "conditional", "condition": self.extract_condition_symbol(line)}
    elif self.is_double_line(line):
        return {"type": "bidirectional"}
    else:
        return {"type": "simple"}
```

## AST生成 {#ast-generation}

### シンボルベースASTノード
```python
class ASTNode:
    def __init__(self, symbol_shape):
        self.shape = symbol_shape
        self.internal_pattern = symbol_shape.internal_pattern

class CircleNode(ASTNode):  # 関数/スコープ
    def __init__(self, shape, inner_symbols, connections):
        super().__init__(shape)
        self.inner_symbols = inner_symbols  # 内部のシンボル群
        self.connections = connections      # 接続情報

class SquareNode(ASTNode):  # 変数/データ
    def __init__(self, shape, dot_pattern=None):
        super().__init__(shape)
        self.value = self.parse_dots(dot_pattern) if dot_pattern else None
        self.type = self.infer_type_from_outline(shape)

class TriangleNode(ASTNode):  # 条件分岐
    def __init__(self, shape, condition_symbol, branches):
        super().__init__(shape)
        self.condition = condition_symbol
        self.true_branch = branches.get('true')
        self.false_branch = branches.get('false')

class PentagonNode(ASTNode):  # ループ
    def __init__(self, shape, loop_body):
        super().__init__(shape)
        self.body = loop_body
        self.iteration_pattern = self.extract_iteration_pattern(shape)

class HexagonNode(ASTNode):  # 並列処理
    def __init__(self, shape, parallel_tasks):
        super().__init__(shape)
        self.tasks = parallel_tasks
```

### シンボルからASTへの変換
```python
def symbol_to_ast(symbol, topology):
    """純粋シンボルをASTノードに変換"""
    shape_type = symbol.shape_type
    
    if shape_type == "circle":
        inner_symbols = topology.get_contained_symbols(symbol)
        connections = topology.get_connections(symbol)
        return CircleNode(symbol, inner_symbols, connections)
    
    elif shape_type == "square":
        dot_pattern = symbol.internal_pattern
        return SquareNode(symbol, dot_pattern)
    
    elif shape_type == "triangle":
        condition = extract_condition_symbol(symbol)
        branches = topology.get_branches(symbol)
        return TriangleNode(symbol, condition, branches)
    
    elif shape_type == "pentagon":
        body = topology.get_loop_body(symbol)
        return PentagonNode(symbol, body)
    
    elif shape_type == "hexagon":
        tasks = topology.get_parallel_tasks(symbol)
        return HexagonNode(symbol, tasks)
```

## 型推論 {#type-inference}

### シンボルベース型推論
```python
class SymbolTypeInferencer:
    def infer_type(self, symbol):
        # 1. 輪郭スタイルから基本型を推論
        outline_type = self.infer_from_outline(symbol.outline_style)
        
        # 2. 内部パターンから詳細型を推論
        if symbol.has_wavy_interior():
            return ArrayType(outline_type)
        elif symbol.has_grid_pattern():
            return MatrixType(outline_type)
        elif symbol.has_nested_shape():
            return ObjectType(self.analyze_nested_structure(symbol))
        
        return outline_type
    
    def propagate_types(self, ast, topology):
        """接続を通じて型を伝播"""
        type_constraints = []
        
        # 接続された図形間で型制約を収集
        for connection in topology.connections:
            source_type = self.infer_type(connection.source)
            target_type = self.infer_type(connection.target)
            
            # 演算子による型制約
            if connection.has_operator():
                constraint = self.operator_constraint(
                    connection.operator,
                    source_type,
                    target_type
                )
                type_constraints.append(constraint)
        
        # 制約を解決
        return self.solve_constraints(type_constraints)
```

## コード生成 {#code-generation}

### シンボルから実行可能コードへの変換
```python
class SymbolCodeGenerator:
    def generate(self, ast):
        """ASTから実行可能コードを生成"""
        code = []
        
        # メインサークルを見つける
        main_circle = self.find_main_circle(ast)
        
        # 各シンボルをコードに変換
        for node in ast.traverse():
            if isinstance(node, CircleNode):
                code.append(self.generate_function(node))
            elif isinstance(node, SquareNode):
                code.append(self.generate_variable(node))
            elif isinstance(node, TriangleNode):
                code.append(self.generate_conditional(node))
            elif isinstance(node, PentagonNode):
                code.append(self.generate_loop(node))
            elif isinstance(node, HexagonNode):
                code.append(self.generate_parallel(node))
        
        return '\n'.join(code)
    
    def generate_variable(self, square_node):
        """ドットパターンから値を生成"""
        if square_node.value is not None:
            # ドット配列を数値に変換
            value = self.dots_to_value(square_node.value)
            return f"var_{square_node.id} = {value}"
        else:
            return f"var_{square_node.id} = null"
```

### 演算子シンボルの変換
```python
def translate_operator_symbol(operator_shape):
    """演算子シンボルを実際の演算に変換"""
    operator_map = {
        "plus_shape": "+",
        "minus_shape": "-",
        "x_shape": "*",
        "slash_shape": "/",
        "equal_shape": "==",
        "greater_shape": ">",
        "less_shape": "<",
        "and_shape": "&&",
        "or_shape": "||",
        "not_shape": "!"
    }
    
    pattern = recognize_operator_pattern(operator_shape)
    return operator_map.get(pattern, "unknown_op")
```

## 最適化 {#optimization}

### 視覚的最適化ヒント
- **線の太さ**: ホットパスを示す
- **図形の塗りつぶし密度**: 最適化レベルを示唆
- **重なる図形**: インライン化を有効に
- **図形のグループ化**: ベクトル化可能な操作

## エラー処理 {#error-handling}

### シンボル認識エラー
```python
class SymbolError:
    def __init__(self, shape, error_type):
        self.shape = shape
        self.error_type = error_type
        self.location = shape.bounding_box
    
    def get_error_message(self):
        errors = {
            "ambiguous_shape": "曖昧な図形：頂点数が不明確",
            "incomplete_circle": "不完全な円：閉じていない",
            "invalid_dots": "無効なドットパターン：3x3グリッドに収まらない",
            "unknown_operator": "認識できない演算子シンボル",
            "conflicting_connections": "矛盾する接続：複数の出力",
            "isolated_symbol": "孤立したシンボル：接続なし"
        }
        return errors.get(self.error_type, "不明なエラー")
```

## デバッグサポート {#debug-support}

### ビジュアルデバッグ
```python
class VisualDebugger:
    def __init__(self, original_image):
        self.image = original_image.copy()
        self.execution_state = {}
    
    def highlight_current_symbol(self, symbol):
        """現在実行中のシンボルをハイライト"""
        cv2.drawContours(self.image, [symbol.contour], -1, (0, 255, 0), 3)
    
    def show_symbol_value(self, symbol, value):
        """シンボルの現在値を表示（ドットパターンで）"""
        dot_pattern = self.value_to_dots(value)
        self.draw_dot_pattern_near_symbol(symbol, dot_pattern)
    
    def trace_execution_path(self, path):
        """実行パスを矢印で可視化"""
        for i in range(len(path) - 1):
            start = path[i].center
            end = path[i + 1].center
            cv2.arrowedLine(self.image, start, end, (255, 0, 0), 2)
```

---

*この仕様はGrimoireコンパイラの純粋シンボルベースアーキテクチャを定義します。テキストは一切使用せず、図形とパターンのみで全ての計算を表現します。*