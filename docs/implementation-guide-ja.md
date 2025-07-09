# Grimoire実装ガイド

## 1. プロジェクト構造

```
grimoire/
├── src/
│   ├── compiler/
│   │   ├── lexer.rs         # 画像からトークン抽出
│   │   ├── parser.rs        # トポロジー解析とAST構築
│   │   ├── type_checker.rs  # 型推論と検証
│   │   ├── codegen.rs       # コード生成
│   │   └── optimizer.rs     # 最適化パス
│   ├── vision/
│   │   ├── preprocessor.rs  # 画像前処理
│   │   ├── shape_detector.rs # 図形認識
│   │   ├── symbol_recognizer.rs # シンボル認識
│   │   └── topology.rs      # 接続解析
│   ├── runtime/
│   │   ├── vm.rs           # バイトコードVM
│   │   ├── debugger.rs     # デバッガサポート
│   │   └── visualizer.rs   # 実行可視化
│   └── main.rs             # CLIエントリーポイント
├── tests/
├── examples/
└── docs/
```

## 2. 実装の段階

### フェーズ1: 基礎実装（MVP）
1. 基本的な図形認識（円、四角、三角）
2. 単純な接続検出
3. 基本的なAST生成
4. C言語へのトランスパイル

### フェーズ2: 中級機能
1. 全図形のサポート
2. シンボル認識
3. 型推論
4. エラーハンドリング

### フェーズ3: 高度な機能
1. 並列処理サポート
2. デバッガ統合
3. 最適化
4. LLVM統合

## 3. 技術スタック推奨

### 言語選択
- **Rust**: 安全性とパフォーマンス
- **Python**: プロトタイピングと画像処理
- **TypeScript**: Webベースのビジュアライザ

### ライブラリ
```toml
# Cargo.toml
[dependencies]
# 画像処理
image = "0.24"
imageproc = "0.23"
opencv = "0.88"

# パーサ・コンパイラ
nom = "7.1"          # パーサコンビネータ
inkwell = "0.2"      # LLVM バインディング

# その他
clap = "4.0"         # CLI
serde = "1.0"        # シリアライゼーション
tokio = "1.0"        # 非同期ランタイム
```

## 4. 画像認識実装詳細

### 前処理パイプライン
```rust
pub fn preprocess_image(image: &DynamicImage) -> Result<GrayImage> {
    let gray = image.to_luma8();
    
    // ノイズ除去
    let blurred = gaussian_blur(&gray, 1.0);
    
    // 適応的二値化
    let binary = adaptive_threshold(&blurred, 11);
    
    // モルフォロジー演算で線を強調
    let kernel = morphology_kernel(3);
    let cleaned = morphology_close(&binary, &kernel);
    
    Ok(cleaned)
}
```

### 図形検出アルゴリズム
```rust
pub fn detect_shapes(image: &GrayImage) -> Vec<Shape> {
    let contours = find_contours(image);
    let mut shapes = Vec::new();
    
    for contour in contours {
        let approx = approximate_polygon(&contour, 0.02);
        let shape_type = classify_shape(&approx);
        
        shapes.push(Shape {
            shape_type,
            contour,
            center: calculate_center(&contour),
            bounds: calculate_bounds(&contour),
        });
    }
    
    shapes
}
```

## 5. AST設計

```rust
#[derive(Debug, Clone)]
pub enum ASTNode {
    Program {
        main: Box<ASTNode>,
        functions: Vec<ASTNode>,
        globals: Vec<ASTNode>,
    },
    Function {
        name: String,
        params: Vec<Parameter>,
        body: Box<ASTNode>,
        return_type: Type,
    },
    Variable {
        name: String,
        var_type: Type,
        value: Option<Box<ASTNode>>,
        is_const: bool,
    },
    Conditional {
        condition: Box<ASTNode>,
        then_branch: Box<ASTNode>,
        else_branch: Option<Box<ASTNode>>,
    },
    Loop {
        count: LoopCount,
        body: Box<ASTNode>,
    },
    Parallel {
        tasks: Vec<ASTNode>,
    },
    // ... 他のノードタイプ
}
```

## 6. 型システム

```rust
#[derive(Debug, Clone, PartialEq)]
pub enum Type {
    Integer,
    Float,
    String,
    Boolean,
    Array(Box<Type>),
    Map(Box<Type>, Box<Type>),
    Function(Vec<Type>, Box<Type>),
    Reference(Box<Type>),
    Optional(Box<Type>),
    Unknown,
}

pub struct TypeInferencer {
    constraints: Vec<TypeConstraint>,
    substitutions: HashMap<TypeVar, Type>,
}
```

## 7. コード生成戦略

### Cバックエンド例
```rust
impl CCodeGenerator {
    pub fn generate_function(&mut self, func: &ASTNode) -> String {
        match func {
            ASTNode::Function { name, params, body, return_type } => {
                let params_str = self.generate_params(params);
                let return_str = self.type_to_c(return_type);
                let body_str = self.generate_statements(body);
                
                format!("{} {}({}) {{\n{}\n}}", 
                    return_str, name, params_str, body_str)
            }
            _ => panic!("Expected function node"),
        }
    }
}
```

## 8. デバッガ実装

### ソースマップ生成
```rust
pub struct SourceMap {
    // 図形ID -> 生成されたコードの行番号
    shape_to_line: HashMap<ShapeId, LineNumber>,
    // 行番号 -> 元の図形の位置
    line_to_shape: HashMap<LineNumber, ShapeLocation>,
}
```

### ビジュアルデバッガ
```rust
pub struct VisualDebugger {
    original_image: DynamicImage,
    shapes: Vec<Shape>,
    current_shape: Option<ShapeId>,
    breakpoints: HashSet<ShapeId>,
    watch_list: Vec<WatchExpression>,
}

impl VisualDebugger {
    pub fn highlight_current(&mut self) -> DynamicImage {
        let mut img = self.original_image.clone();
        if let Some(shape_id) = self.current_shape {
            // 現在実行中の図形を赤でハイライト
            draw_highlight(&mut img, &self.shapes[shape_id], RED);
        }
        img
    }
}
```

## 9. テスト戦略

### ユニットテスト
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_circle_detection() {
        let test_image = create_test_circle();
        let shapes = detect_shapes(&test_image);
        assert_eq!(shapes.len(), 1);
        assert_eq!(shapes[0].shape_type, ShapeType::Circle);
    }
}
```

### 統合テスト
```rust
#[test]
fn test_hello_world_compilation() {
    let image = load_test_image("hello_world.png");
    let result = compile_image(&image);
    assert!(result.is_ok());
    
    let output = execute_compiled(result.unwrap());
    assert_eq!(output, "Hello World\n");
}
```

## 10. パフォーマンス最適化

### 画像処理の並列化
```rust
use rayon::prelude::*;

pub fn parallel_shape_detection(regions: Vec<ImageRegion>) -> Vec<Shape> {
    regions.par_iter()
        .flat_map(|region| detect_shapes_in_region(region))
        .collect()
}
```

### キャッシング
```rust
pub struct CompilationCache {
    shape_cache: HashMap<ImageHash, Vec<Shape>>,
    ast_cache: HashMap<ShapeHash, ASTNode>,
}
```

---

このガイドは、Grimoireコンパイラの実装を始めるための基礎を提供します。