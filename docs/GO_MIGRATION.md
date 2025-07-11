# Go言語移行ドキュメント

## 概要

Grimoireのコア実装をPythonからGo言語に移行しました。この移行により、大幅なパフォーマンス向上と、デプロイメントの簡素化を実現しています。

## 移行の理由

### 1. パフォーマンス問題
- PyInstallerによるバイナリ化で起動時間が14.9秒に
- OpenCVの初期化オーバーヘッドが大きい
- 画像処理の実行速度が遅い

### 2. デプロイメント問題
- 80MBという大きなバイナリサイズ
- OpenCVやNumPyなどの重い依存関係
- プラットフォーム固有の問題

## 移行結果

### パフォーマンス改善

| 指標 | Python版 | Go版 | 改善率 |
|------|----------|------|--------|
| 起動時間 | 14.9秒 | 0.06秒 | 248倍 |
| 実行時間 | 3秒 | 0.2秒 | 15倍 |
| バイナリサイズ | 80MB | 10MB | 8倍小 |

### 技術的改善
- **依存関係ゼロ**: Pure Goで実装、外部ライブラリ不要
- **クロスプラットフォーム**: 単一バイナリで全OS対応
- **並行処理**: Goのgoroutineによる効率的な並行処理

## アーキテクチャ

### モジュール構成

```
internal/
├── detector/       # 画像認識・シンボル検出
│   ├── detector.go      # メインの検出ロジック
│   ├── image_utils.go   # 画像処理ユーティリティ
│   ├── contour.go       # 輪郭検出
│   ├── shape_classifier.go # 図形分類
│   ├── pattern_detector.go # パターン認識
│   └── connection.go    # 接続線検出
├── parser/         # 構文解析
│   ├── parser.go        # パーサー実装
│   ├── ast.go          # AST定義
│   └── types.go        # 型定義
├── compiler/       # コード生成
│   └── compiler.go     # Python コード生成
└── cli/           # CLIインターフェース
    └── cli.go         # Cobraベースのコマンド
```

### 主要アルゴリズム

#### 1. 画像前処理
```go
// ガウシアンブラー → 適応的二値化 → モルフォロジー演算
preprocessImage(gray) -> binary
```

#### 2. 輪郭検出
```go
// Moore近傍探索による輪郭追跡
findContours(binary) -> []Contour
```

#### 3. 図形分類
```go
// Douglas-Peucker多角形近似 + 頂点数による分類
classifyShape(contour) -> SymbolType
```

#### 4. パターン検出
```go
// 図形内部のドット、ライン、クロスパターンの認識
detectInternalPattern(contour, binary) -> string
```

## 実装の詳細

### Pure Go画像処理

OpenCVに依存せず、Go標準ライブラリのみで実装：

- `image`パッケージによる基本的な画像操作
- カスタムフィルタの実装（ガウシアンブラー、適応的閾値処理）
- Moore近傍探索アルゴリズムによる輪郭検出
- Douglas-Peucker多角形近似アルゴリズム

### 並行処理の活用

```go
// 複数の輪郭を並行処理
for _, contour := range contours {
    go processContour(contour)
}
```

## 移行状況

### 完了した機能 ✅
- 基本的な画像処理
- 輪郭検出
- 図形分類（円、三角形、五角形、六角形、星）
- パターン認識（ドット、ライン、クロス）
- 構文解析
- Pythonコード生成
- CLIインターフェース
- Hello Worldプログラムの実行

### 進行中の機能 🚧
- 四角形の検出精度向上
- 演算子記号の認識改善
- 接続線の完全な検出

### 未実装の機能 ❌
- 複雑な演算子（比較、論理演算）
- 配列・マップ型のサポート
- エラーハンドリングの改善

## 今後の計画

1. **検出精度の向上**
   - 四角形検出アルゴリズムの改善
   - より堅牢な図形分類

2. **機能の完全性**
   - すべての演算子のサポート
   - 複雑なデータ型の実装

3. **最適化**
   - さらなるパフォーマンス改善
   - メモリ使用量の削減

## 開発者向け情報

### ビルド方法

```bash
# 開発用ビルド
go build -o grimoire cmd/grimoire/main.go

# リリース用ビルド（最適化）
go build -ldflags="-s -w" -o grimoire cmd/grimoire/main.go
```

### テスト実行

```bash
# 単体テスト
go test ./...

# カバレッジ付き
go test -cover ./...

# 統合テスト
go test -tags=integration ./...
```

### デバッグ

```bash
# デバッグモードで実行
GRIMOIRE_DEBUG=1 grimoire compile image.png
```

## 貢献ガイドライン

1. Goの標準的なコーディング規約に従う
2. 外部ライブラリの使用は最小限に
3. すべての公開関数にドキュメントコメントを記載
4. テストカバレッジ80%以上を維持