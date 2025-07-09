# Grimoire実装ガイド

## 1. プロジェクト構造

```
grimoire/
├── src/
│   ├── compiler/
│   │   ├── symbol_lexer.rs   # シンボル認識とトークン化
│   │   ├── pattern_parser.rs # パターン解析とAST構築
│   │   ├── pattern_types.rs  # パターンベース型システム
│   │   ├── codegen.rs       # コード生成
│   │   └── optimizer.rs     # 最適化パス
│   ├── vision/
│   │   ├── preprocessor.rs  # 画像前処理
│   │   ├── shape_detector.rs # 基本図形認識
│   │   ├── pattern_recognizer.rs # パターン認識
│   │   ├── symbol_matcher.rs # シンボルマッチング
│   │   └── flow_analyzer.rs  # フロー線解析
│   ├── symbols/
│   │   ├── basic_symbols.rs  # 基本シンボル定義
│   │   ├── operators.rs      # 演算子シンボル
│   │   ├── numeric_dots.rs   # 数値表現（ドット）
│   │   └── special_glyphs.rs # 特殊記号
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
1. 基本図形認識（○円、□四角、△三角）
2. 基本シンボル認識（+、-、×、÷、=）
3. ドット数値システム（•、••、•••）
4. 単純なフロー線検出（─、→）
5. 基本的なAST生成

### フェーズ2: 中級機能
1. 全図形サポート（◎二重円、⬟五角形、⬢六角形、⯃八角形）
2. パターンベース型認識（•整数、••浮動小数点、≡文字列、◐ブール）
3. 複雑な数値表現（⦿十、⊡百、⊙千）
4. 特殊グリフ認識（☆出力、★入力、⚡エラー）
5. 条件付きフロー（┈、╌、～）

### フェーズ3: 高度な機能
1. 並列処理記号（⬢六角形）のサポート
2. デバッグマーカー（👁、🐛、📍、🔍）
3. メタ情報記号（▲、●、■、♦）
4. 双方向フロー（⟷）と同期フロー（═）
5. 最適化とLLVM統合

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

## 4. シンボル認識実装詳細

### 前処理パイプライン
```rust
pub fn preprocess_image(image: &DynamicImage) -> Result<ProcessedImage> {
    let gray = image.to_luma8();
    
    // ノイズ除去
    let blurred = gaussian_blur(&gray, 1.0);
    
    // 適応的二値化
    let binary = adaptive_threshold(&blurred, 11);
    
    // シンボル認識のための特徴抽出
    let features = extract_symbol_features(&binary);
    
    Ok(ProcessedImage {
        binary,
        features,
        original: image.clone(),
    })
}
```

### シンボル認識エンジン
```rust
pub struct SymbolRecognizer {
    shape_templates: HashMap<ShapeType, Template>,
    operator_templates: HashMap<OperatorType, Template>,
    special_glyphs: HashMap<GlyphType, Template>,
}

impl SymbolRecognizer {
    pub fn recognize_symbol(&self, region: &ImageRegion) -> Symbol {
        // 基本図形チェック
        if let Some(shape) = self.match_shape(region) {
            return self.analyze_shape_pattern(shape, region);
        }
        
        // 演算子チェック
        if let Some(op) = self.match_operator(region) {
            return Symbol::Operator(op);
        }
        
        // 特殊記号チェック
        if let Some(glyph) = self.match_glyph(region) {
            return Symbol::SpecialGlyph(glyph);
        }
        
        Symbol::Unknown
    }
    
    fn analyze_shape_pattern(&self, shape: ShapeType, region: &ImageRegion) -> Symbol {
        // 図形内部のパターンを解析
        let inner_pattern = extract_inner_pattern(region);
        
        match shape {
            ShapeType::Square => {
                // 四角内のパターンで型を判定
                match analyze_dot_pattern(&inner_pattern) {
                    DotPattern::Single => Symbol::Variable(DataType::Integer),
                    DotPattern::Double => Symbol::Variable(DataType::Float),
                    DotPattern::TripleLines => Symbol::Variable(DataType::String),
                    DotPattern::HalfCircle => Symbol::Variable(DataType::Boolean),
                    _ => Symbol::Variable(DataType::Unknown),
                }
            },
            ShapeType::Circle => {
                if is_double_circle(region) {
                    Symbol::MainEntry
                } else {
                    Symbol::Function
                }
            },
            _ => Symbol::Shape(shape),
        }
    }
}
```

### 数値ドット認識システム
```rust
pub struct NumericDotRecognizer {
    dot_templates: Vec<DotTemplate>,
}

impl NumericDotRecognizer {
    pub fn recognize_number(&self, region: &ImageRegion) -> Option<i32> {
        let dots = self.detect_dots(region);
        let pattern = self.analyze_dot_arrangement(&dots);
        
        match pattern {
            DotArrangement::Single => Some(1),
            DotArrangement::Double => Some(2),
            DotArrangement::Triple => Some(3),
            DotArrangement::Superscript(n) => Some(n),
            DotArrangement::CircledDot => Some(10),
            DotArrangement::CircledDots(n) => Some(10 + n),
            DotArrangement::DoubleCircled => Some(100),
            DotArrangement::TripleCircled => Some(1000),
            DotArrangement::Compound(base, modifier) => {
                Some(self.calculate_compound_value(base, modifier))
            },
            _ => None,
        }
    }
    
    fn detect_dots(&self, region: &ImageRegion) -> Vec<DotLocation> {
        // ハフ変換で円形の点を検出
        let circles = hough_circle_transform(region);
        circles.into_iter()
            .filter(|c| c.radius < DOT_MAX_RADIUS)
            .map(|c| DotLocation { x: c.x, y: c.y })
            .collect()
    }
}
```

### パターンベース型システム
```rust
pub enum PatternType {
    SingleDot,      // • 整数
    DoubleDot,      // •• 浮動小数点
    TripleLines,    // ≡ 文字列
    HalfCircle,     // ◐ ブール
    StarPattern,    // ※ 配列
    GridPattern,    // ⊞ マップ
    EmptySet,       // ∅ null/void
}

pub struct PatternTypeMatcher {
    patterns: HashMap<PatternType, PatternTemplate>,
}

impl PatternTypeMatcher {
    pub fn match_type(&self, inner_region: &ImageRegion) -> DataType {
        // 各パターンテンプレートとマッチング
        for (pattern_type, template) in &self.patterns {
            if self.matches_template(inner_region, template) {
                return pattern_type_to_data_type(pattern_type);
            }
        }
        DataType::Unknown
    }
    
    fn matches_template(&self, region: &ImageRegion, template: &PatternTemplate) -> bool {
        // テンプレートマッチングアルゴリズム
        let score = template_match_score(region, template);
        score > MATCH_THRESHOLD
    }
}
```

```rust
#[derive(Debug, Clone)]
pub enum ASTNode {
    Program {
        entry: Symbol,  // ◎ メインエントリ
        functions: Vec<ASTNode>,
        globals: Vec<ASTNode>,
    },
    Function {
        symbol: Symbol,  // ○ 関数シンボル
        inputs: Vec<FlowConnection>,  // → 入力接続
        outputs: Vec<FlowConnection>, // ← 出力接続
        body: Box<ASTNode>,
    },
    Variable {
        symbol: Symbol,  // □ with pattern
        pattern_type: PatternType,
        value: Option<NumericValue>,
    },
    Conditional {
        symbol: Symbol,  // △ 判定シンボル
        condition_op: OperatorSymbol,
        true_flow: Option<Box<ASTNode>>,  // ┈ 真のフロー
        false_flow: Option<Box<ASTNode>>, // ╌ 偽のフロー
    },
    Loop {
        symbol: Symbol,  // ⬟ ループシンボル
        count: NumericValue,
        body: Box<ASTNode>,
        loop_back: FlowConnection, // ⟲
    },
    Parallel {
        symbol: Symbol,  // ⬢ 並列シンボル
        tasks: Vec<ASTNode>,
    },
    Output {
        symbol: Symbol,  // ☆ 出力シンボル
    },
    Input {
        symbol: Symbol,  // ★ 入力シンボル
    },
}

#[derive(Debug, Clone)]
pub struct NumericValue {
    dots: Vec<DotPattern>,
    value: i32,
}

#[derive(Debug, Clone)]
pub struct FlowConnection {
    line_type: LineType,  // ─, ┈, ╌, ～, ═, ⟷
    direction: FlowDirection, // →, ←, ↔, ⊸, ⟲
}
```

## 6. シンボルベース演算子実装

```rust
#[derive(Debug, Clone, PartialEq)]
pub enum OperatorSymbol {
    // 算術演算子
    Plus,        // +
    Minus,       // -
    Multiply,    // ×
    Divide,      // ÷
    
    // 比較演算子
    Equals,      // =
    NotEquals,   // ≠
    LessThan,    // <
    GreaterThan, // >
    LessOrEqual, // ≤
    GreaterOrEqual, // ≥
    
    // 論理演算子
    And,         // ∧
    Or,          // ∨
    Not,         // ¬
    Xor,         // ⊕
}

pub struct OperatorRecognizer {
    templates: HashMap<OperatorSymbol, SymbolTemplate>,
}

impl OperatorRecognizer {
    pub fn recognize(&self, region: &ImageRegion) -> Option<OperatorSymbol> {
        // OCR不要、純粋なパターンマッチング
        for (op, template) in &self.templates {
            if self.match_symbol(region, template) {
                return Some(op.clone());
            }
        }
        None
    }
    
    fn match_symbol(&self, region: &ImageRegion, template: &SymbolTemplate) -> bool {
        // 形状ベースのマッチング
        let features = extract_shape_features(region);
        template.matches(&features)
    }
}
```

## 7. フロー制御実装

```rust
#[derive(Debug, Clone)]
pub enum LineType {
    Normal,      // ─ 通常のフロー
    Conditional, // ┈ 条件付きフロー
    Alternative, // ╌ 代替フロー
    Exception,   // ～ 例外フロー
    Synchronous, // ═ 同期フロー
    Bidirectional, // ⟷ 双方向フロー
}

pub struct FlowAnalyzer {
    line_detector: LineDetector,
    connection_mapper: ConnectionMapper,
}

impl FlowAnalyzer {
    pub fn analyze_flows(&self, image: &ProcessedImage) -> FlowGraph {
        // 線分検出
        let lines = self.line_detector.detect_lines(image);
        
        // 線種分類
        let classified_lines = lines.into_iter()
            .map(|line| self.classify_line_type(&line))
            .collect();
        
        // 接続マッピング
        let connections = self.connection_mapper.map_connections(
            &classified_lines,
            &image.symbols
        );
        
        FlowGraph::new(connections)
    }
    
    fn classify_line_type(&self, line: &Line) -> ClassifiedLine {
        match line.pattern {
            LinePattern::Solid => ClassifiedLine::new(line, LineType::Normal),
            LinePattern::Dashed => ClassifiedLine::new(line, LineType::Conditional),
            LinePattern::Dotted => ClassifiedLine::new(line, LineType::Alternative),
            LinePattern::Wavy => ClassifiedLine::new(line, LineType::Exception),
            LinePattern::Double => ClassifiedLine::new(line, LineType::Synchronous),
            LinePattern::Arrows => ClassifiedLine::new(line, LineType::Bidirectional),
        }
    }
}
```

## 8. コード生成戦略

### シンボルベースコード生成
```rust
impl SymbolCodeGenerator {
    pub fn generate(&mut self, ast: &ASTNode) -> String {
        match ast {
            ASTNode::Program { entry, functions, globals } => {
                self.generate_program(entry, functions, globals)
            },
            ASTNode::Output { symbol } => {
                // ☆ シンボルは標準出力へ
                "printf(\"*\\n\");".to_string()
            },
            ASTNode::Variable { symbol, pattern_type, value } => {
                self.generate_variable(pattern_type, value)
            },
            ASTNode::Loop { symbol, count, body, loop_back } => {
                self.generate_loop(count, body)
            },
            _ => self.generate_node(ast),
        }
    }
    
    fn generate_numeric_literal(&self, value: &NumericValue) -> String {
        format!("{}", value.value)
    }
}
```

## 9. デバッガ実装

### シンボルベースデバッガ
```rust
pub struct SymbolDebugger {
    // シンボルID -> 生成されたコードの行番号
    symbol_to_line: HashMap<SymbolId, LineNumber>,
    // 行番号 -> 元のシンボル位置
    line_to_symbol: HashMap<LineNumber, SymbolLocation>,
    // デバッグマーカー
    debug_markers: Vec<DebugMarker>,
}

#[derive(Debug)]
pub enum DebugMarker {
    WatchPoint { symbol: Symbol, location: Point },      // 👁
    BreakPoint { symbol: Symbol, location: Point },      // 🐛
    Assertion { symbol: Symbol, condition: Expression },  // 📍
    TracePoint { symbol: Symbol, location: Point },      // 🔍
}

impl SymbolDebugger {
    pub fn detect_debug_markers(&mut self, image: &ProcessedImage) {
        // デバッグシンボルの検出
        for region in &image.regions {
            if let Some(marker) = self.recognize_debug_marker(region) {
                self.debug_markers.push(marker);
            }
        }
    }
    
    pub fn visualize_execution(&self, image: &DynamicImage, current_symbol: SymbolId) -> DynamicImage {
        let mut vis_image = image.clone();
        
        // 現在実行中のシンボルをハイライト
        if let Some(location) = self.get_symbol_location(current_symbol) {
            draw_glow_effect(&mut vis_image, location, EXECUTION_COLOR);
        }
        
        // フロー線のアニメーション
        self.animate_flow_lines(&mut vis_image, current_symbol);
        
        vis_image
    }
}
```

## 10. テスト戦略

### シンボル認識テスト
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_basic_symbol_recognition() {
        let test_image = create_test_symbol("circle");
        let symbol = recognize_symbol(&test_image);
        assert_eq!(symbol, Symbol::Function);
    }
    
    #[test]
    fn test_double_circle_recognition() {
        let test_image = create_test_symbol("double_circle");
        let symbol = recognize_symbol(&test_image);
        assert_eq!(symbol, Symbol::MainEntry);
    }
    
    #[test]
    fn test_numeric_dot_recognition() {
        // 3つのドット = 3
        let test_image = create_dots_pattern(3);
        let value = recognize_numeric_value(&test_image);
        assert_eq!(value, Some(3));
        
        // 囲みドット = 10
        let test_image = create_circled_dot();
        let value = recognize_numeric_value(&test_image);
        assert_eq!(value, Some(10));
    }
    
    #[test]
    fn test_operator_recognition() {
        let operators = vec![
            ("+", OperatorSymbol::Plus),
            ("×", OperatorSymbol::Multiply),
            ("≤", OperatorSymbol::LessOrEqual),
            ("∧", OperatorSymbol::And),
        ];
        
        for (symbol_str, expected) in operators {
            let test_image = create_operator_image(symbol_str);
            let op = recognize_operator(&test_image);
            assert_eq!(op, Some(expected));
        }
    }
}
```

### 統合テスト
```rust
#[test]
fn test_star_output_compilation() {
    // ◎ → ☆ (メインエントリから星出力)
    let image = load_test_image("star_output.grim");
    let result = compile_image(&image);
    assert!(result.is_ok());
    
    let output = execute_compiled(result.unwrap());
    assert_eq!(output, "*\n");
}

#[test]
fn test_loop_compilation() {
    // ⬟ ← □⦿ (10回ループ)
    let image = load_test_image("loop_ten.grim");
    let result = compile_image(&image);
    assert!(result.is_ok());
    
    let generated_code = result.unwrap();
    assert!(generated_code.contains("for"));
    assert!(generated_code.contains("10"));
}
```

## 11. パフォーマンス最適化

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

## 12. 完全なコンパイラパイプライン

### メインコンパイラ実装
```rust
pub struct GrimoireCompiler {
    preprocessor: ImagePreprocessor,
    symbol_recognizer: SymbolRecognizer,
    pattern_matcher: PatternTypeMatcher,
    numeric_recognizer: NumericDotRecognizer,
    operator_recognizer: OperatorRecognizer,
    flow_analyzer: FlowAnalyzer,
    parser: SymbolParser,
    code_generator: SymbolCodeGenerator,
}

impl GrimoireCompiler {
    pub fn compile(&self, image_path: &Path) -> Result<String> {
        // 1. 画像読み込みと前処理
        let image = image::open(image_path)?;
        let processed = self.preprocessor.process(&image)?;
        
        // 2. シンボル認識フェーズ
        let symbols = self.recognize_all_symbols(&processed)?;
        
        // 3. フロー解析
        let flow_graph = self.flow_analyzer.analyze_flows(&processed)?;
        
        // 4. AST構築
        let ast = self.parser.parse(symbols, flow_graph)?;
        
        // 5. コード生成
        let code = self.code_generator.generate(&ast)?;
        
        Ok(code)
    }
    
    fn recognize_all_symbols(&self, image: &ProcessedImage) -> Result<Vec<RecognizedSymbol>> {
        let mut symbols = Vec::new();
        
        // 並列処理で各領域のシンボルを認識
        let regions = segment_image_regions(image);
        let recognized: Vec<_> = regions.par_iter()
            .map(|region| self.recognize_region(region))
            .collect();
        
        for symbol in recognized {
            symbols.push(symbol?);
        }
        
        Ok(symbols)
    }
    
    fn recognize_region(&self, region: &ImageRegion) -> Result<RecognizedSymbol> {
        // 基本図形チェック
        if let Some(shape) = self.symbol_recognizer.recognize_symbol(region) {
            // 図形内部のパターンチェック
            if let Symbol::Variable(_) = shape {
                let pattern_type = self.pattern_matcher.match_type(region);
                let value = self.numeric_recognizer.recognize_number(region);
                return Ok(RecognizedSymbol::new(shape, Some(pattern_type), value));
            }
            return Ok(RecognizedSymbol::new(shape, None, None));
        }
        
        // 演算子チェック
        if let Some(op) = self.operator_recognizer.recognize(region) {
            return Ok(RecognizedSymbol::Operator(op));
        }
        
        Err(CompileError::UnrecognizedSymbol)
    }
}
```

### 実行例
```rust
fn main() -> Result<()> {
    let args = Args::parse();
    let compiler = GrimoireCompiler::new();
    
    match args.command {
        Command::Compile { input, output } => {
            let code = compiler.compile(&input)?;
            std::fs::write(output, code)?;
            println!("コンパイル完了!");
        },
        Command::Run { input } => {
            let code = compiler.compile(&input)?;
            let temp_file = compile_to_executable(&code)?;
            execute_program(&temp_file)?;
        },
        Command::Debug { input } => {
            let debugger = SymbolDebugger::new();
            debugger.debug_visual_program(&input)?;
        }
    }
    
    Ok(())
}
```

---

このガイドは、純粋にシンボルベースのGrimoireコンパイラを実装するための完全な指針を提供します。テキストを一切使用せず、記号の組み合わせのみでプログラミングを実現する革新的なアプローチです。