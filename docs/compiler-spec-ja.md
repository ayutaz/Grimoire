# Grimoireコンパイラアーキテクチャ仕様

## 目次
1. [概要](#overview)
2. [コンパイルパイプライン](#compilation-pipeline)
3. [画像認識ステージ](#image-recognition)
4. [シンボル抽出](#symbol-extraction)
5. [トポロジー解析](#topology-analysis)
6. [AST生成](#ast-generation)
7. [型推論](#type-inference)
8. [コード生成](#code-generation)
9. [最適化](#optimization)
10. [エラー処理](#error-handling)
11. [デバッグサポート](#debug-support)

## 概要 {#overview}

Grimoireコンパイラは、手描きの魔法陣を実行可能プログラムに変換します。コンピュータビジョン、記号解析、従来のコンパイル技術を組み合わせた多段階パイプラインを使用します。

### コンパイラコンポーネント
```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│画像入力     │ --> │CV処理        │ --> │シンボル抽出  │
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

### ステージ2: 図形認識
1. 輪郭検出
2. 機械学習による図形分類
3. シンボル境界抽出
4. テキスト認識（OCR）

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
            shape_type = "三角形"
        elif vertices == 4:
            shape_type = "四角形"
        elif vertices == 5:
            shape_type = "五角形"
        elif vertices == 6:
            shape_type = "六角形"
        elif vertices > 6:
            # 円判定
            area = cv2.contourArea(contour)
            perimeter = cv2.arcLength(contour, True)
            circularity = 4 * np.pi * area / (perimeter ** 2)
            if circularity > 0.8:
                shape_type = "円"
            else:
                shape_type = "星" if self.is_star(approx) else "多角形"
                
        shapes.append(Shape(shape_type, contour, approx))
    return shapes
```

## シンボル抽出 {#symbol-extraction}

### シンボル認識パイプライン
1. **分離**: 図形からシンボル領域を抽出
2. **正規化**: 標準方向にスケール・回転
3. **特徴抽出**: 
   - 幾何学的特徴（角度、曲線）
   - トポロジカル特徴（穴、交差）
   - 統計的特徴（モーメント、分布）
4. **分類**: ニューラルネットワークまたはテンプレートマッチング

### カスタムシンボル学習
```python
class SymbolRecognizer:
    def __init__(self):
        self.model = self.load_pretrained_model()
        self.custom_symbols = {}
    
    def add_custom_symbol(self, image, meaning):
        features = self.extract_features(image)
        self.custom_symbols[meaning] = features
    
    def recognize(self, symbol_image):
        # まず標準シンボルを試す
        prediction = self.model.predict(symbol_image)
        if prediction.confidence > 0.8:
            return prediction.symbol
            
        # カスタムシンボルにフォールバック
        return self.match_custom_symbol(symbol_image)
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

### 接続検出
- **直接接触**: 図形が境界を共有
- **線接続**: 図形間の明示的な線
- **包含**: 一つの図形が別の図形内部
- **近接**: 暗黙的接続を持つ近い図形

### フロー解析
```python
def analyze_flow(topology):
    # フロー方向を検出：
    # 1. 矢印の方向
    # 2. 時計回り/反時計回りパターン
    # 3. 数値注釈
    # 4. デフォルトは上から下、左から右
    
    flow_graph = FlowGraph()
    for edge in topology.edges:
        direction = infer_direction(edge)
        flow_graph.add_directed_edge(edge.source, edge.target, direction)
    
    return flow_graph
```

## AST生成 {#ast-generation}

### ASTノードタイプ
```python
class ASTNode:
    pass

class ProgramNode(ASTNode):
    def __init__(self, main_circle, functions, globals):
        self.main = main_circle
        self.functions = functions
        self.globals = globals

class CircleNode(ASTNode):  # 関数/スコープ
    def __init__(self, name, params, body):
        self.name = name
        self.params = params
        self.body = body

class SquareNode(ASTNode):  # 変数
    def __init__(self, name, type, value):
        self.name = name
        self.type = type
        self.value = value

class TriangleNode(ASTNode):  # 条件分岐
    def __init__(self, condition, true_branch, false_branch):
        self.condition = condition
        self.true_branch = true_branch
        self.false_branch = false_branch
```

### AST構築アルゴリズム
```python
def build_ast(topology, symbols):
    # 1. メインエントリーポイントを見つける
    main = find_main_circle(topology)
    
    # 2. 関数定義を構築
    functions = []
    for circle in topology.get_circles():
        if circle != main:
            func_ast = build_function_ast(circle, topology)
            functions.append(func_ast)
    
    # 3. メイン実行フローを構築
    main_ast = build_execution_ast(main, topology)
    
    return ProgramNode(main_ast, functions, globals)
```

## 型推論 {#type-inference}

### 型推論ルール
1. **図形ベース推論**: 四角の輪郭スタイルが型を示す
2. **演算子ベース推論**: 接続された演算子が型を制約
3. **フローベース推論**: 型は接続を通じて伝播
4. **注釈ベース**: 明示的な型シンボルが推論を上書き

### 型推論アルゴリズム
```python
class TypeInferencer:
    def infer_types(self, ast):
        # 制約グラフを構築
        constraints = self.collect_constraints(ast)
        
        # 単一化で制約を解決
        substitutions = self.unify(constraints)
        
        # ASTに型を適用
        typed_ast = self.apply_types(ast, substitutions)
        
        # 競合をチェック
        self.verify_types(typed_ast)
        
        return typed_ast
```

## コード生成 {#code-generation}

### バックエンドオプション

#### 1. Cバックエンド
```python
class CCodeGenerator:
    def generate(self, ast):
        self.emit("#include <stdio.h>")
        self.emit("#include <stdlib.h>")
        
        # 関数宣言を生成
        for func in ast.functions:
            self.generate_function_decl(func)
        
        # mainを生成
        self.emit("int main() {")
        self.generate_statements(ast.main.body)
        self.emit("return 0;")
        self.emit("}")
```

#### 2. LLVMバックエンド
```python
class LLVMCodeGenerator:
    def __init__(self):
        self.module = llvm.Module()
        self.builder = llvm.Builder()
    
    def generate(self, ast):
        # LLVM IRを生成
        for func in ast.functions:
            self.generate_function(func)
        
        # mainを生成
        main_func = self.module.add_function("main", ...)
        self.generate_main(ast.main)
```

#### 3. バイトコードバックエンド
```python
class BytecodeGenerator:
    def generate(self, ast):
        bytecode = []
        
        # 命令を生成
        for node in ast.walk():
            instructions = self.generate_instructions(node)
            bytecode.extend(instructions)
        
        return GrimoireBytecode(bytecode)
```

## 最適化 {#optimization}

### 視覚的最適化ヒント
- **線の太さ**: ホットパスを示す
- **色の濃度**: 最適化レベルを示唆
- **重なる図形**: インライン化を有効に

### 最適化パス
1. **デッドシェイプ除去**: 到達不可能な図形を削除
2. **シェイプ融合**: 隣接する操作を結合
3. **ループ展開**: 五角形の注釈に基づく
4. **並列検出**: 六角形パターンをスレッドプールへ

## エラー処理 {#error-handling}

### コンパイルエラー
```python
class GrimoireError:
    def __init__(self, shape, message, suggestion=None):
        self.shape = shape
        self.location = shape.bounding_box
        self.message = message
        self.suggestion = suggestion
    
    def visualize(self, image):
        # 元画像にエラーハイライトを描画
        cv2.rectangle(image, self.location, (0, 0, 255), 3)
        # エラーメッセージを追加
        cv2.putText(image, self.message, ...)
```

### エラータイプ
1. **図形認識エラー**
   - 曖昧な図形
   - 不完全な円
   - 重複の競合

2. **意味エラー**
   - 切断されたコンポーネント
   - 型の不一致
   - 未定義シンボル

3. **論理エラー**
   - 出口のない無限ループ
   - 到達不可能なコード
   - 循環依存

## デバッグサポート {#debug-support}

### デバッグ情報生成
```python
class DebugInfo:
    def __init__(self):
        self.shape_to_line = {}  # 図形を生成コード行にマップ
        self.breakpoints = []    # ブレークポイントとしてマークされた図形
        self.watches = []        # ウォッチ用にマークされた図形
    
    def generate_sourcemap(self, shapes, generated_code):
        # 視覚的コードとテキストコード間の双方向マッピングを作成
        pass
```

### ビジュアルデバッガ統合
1. **実行可視化**: 現在実行中の図形をハイライト
2. **変数検査**: 図形の近くに値を表示
3. **ステップ実行**: 図形ごとの実行
4. **タイムトラベルデバッグ**: 実行を視覚的に再生

### デバッグコンパイルモード
```bash
grimoire compile --debug circle.png
# 生成物:
# - circle.grim.debug (デバッグシンボル)
# - circle.grim.map (図形からコードへのマッピング)
# - circle.grim.trace (実行トレースフォーマット)
```

---

*この仕様はGrimoireコンパイラの技術的アーキテクチャを定義します。*