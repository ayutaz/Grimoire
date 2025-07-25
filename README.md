# Grimoire - 純粋記号プログラミング言語

<p align="center">
  <img src="docs/images/grimoire-logo.png" alt="Grimoire Logo" width="600">
</p>

> 「すべてのプログラムは魔法陣である」

Grimoireは、文字を一切使わず、純粋に記号と図形のみでプログラムを表現する実験的なビジュアルプログラミング言語です。手描きの魔法陣を画像認識によってコンパイルし、実行可能なプログラムを作成します。

**[🎮 Webデモを試す](https://ayutaz.github.io/Grimoire/)**

![Build Status](https://github.com/ayutaz/Grimoire/actions/workflows/go.yml/badge.svg)
[![Test Coverage](https://codecov.io/gh/ayutaz/Grimoire/branch/main/graph/badge.svg)](https://codecov.io/gh/ayutaz/Grimoire)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

## 🌟 特徴

- **魔法陣パラダイム**: すべてのプログラムは外周円を持つ魔法陣として表現
- **完全記号化**: 文字を使わない純粋なビジュアル表現
- **画像認識**: 手描きまたはデジタル画像を実行可能プログラムにコンパイル
- **直感的な記号体系**: 図形とパターンで全てを表現
- **空間的プログラミング**: 図形の配置と接続でフローを制御
- **斜め接続線**: 45°と135°の斜め接続線をサポートし、より複雑な魔法陣を構築可能
- **決定的な認識**: OpenCVの古典的コンピュータビジョン技術により、同じ入力に対して常に同じ結果を保証

## 📐 基本記号

### 構造要素
- `◎` - メインエントリポイント
- `○` - 関数/スコープ
- `□` - 変数/データ
- `△` - 条件分岐
- `⬟` - ループ
- `⬢` - 並列処理
- `☆` - 出力/表示

### データ型（図形内パターン）
- `•` - 整数型
- `••` - 浮動小数点型
- `≡` - 文字列型
- `◐` - ブール型
- `※` - 配列型
- `⊞` - マップ型

### 数値表現
- `•` = 1, `••` = 2, `•••` = 3
- `⦿` = 10, `⊡` = 100, `⊙` = 1000

### 演算記号（魔法陣的表現）
- `⟐` - 結合/加算（エネルギーの収束）
- `⟑` - 分離/減算（エネルギーの分岐）
- `✦` - 増幅/乗算（4点星による力の増幅）
- `⟠` - 分割/除算（8分割円による分配）

## 📚 ドキュメント

- [言語仕様](docs/language-spec-ja.md)
- [コンパイラアーキテクチャ](docs/compiler-spec-ja.md)
- [チュートリアル](docs/tutorial-ja.md)
- [実装ガイド](docs/implementation-guide-ja.md)
- [サンプル集](docs/examples-ja.md)
- [テスト戦略](docs/TEST_STRATEGY.md)

### サンプルプログラム

基本的なサンプル:
- [Hello World](examples/hello-world.grim.md) - 基本的な出力
- [電卓](examples/calculator.grim.md) - 四則演算
- [フィボナッチ](examples/fibonacci.grim.md) - 再帰的な数列
- [ループ](examples/loop.grim.md) - 繰り返し処理
- [並列処理](examples/parallel.grim.md) - 並列実行
- [変数操作](examples/variables.grim.md) - 変数の使い方

アルゴリズムとデータ構造:
- [バブルソート](examples/bubble-sort.grim.md) - ソートアルゴリズム
- [素数判定](examples/prime-check.grim.md) - 数学的アルゴリズム
- [文字列反転](examples/string-reverse.grim.md) - 文字列処理
- [スタック実装](examples/stack-implementation.grim.md) - データ構造
- [最大公約数](examples/euclidean-gcd.grim.md) - ユークリッドの互除法

## 🚀 インストール

### バイナリ配布（推奨）

[Releases](https://github.com/ayutaz/Grimoire/releases)から、お使いのプラットフォーム用のバイナリをダウンロードしてください。

### ソースからのインストール

```bash
# Go 1.21以上が必要
go install github.com/ayutaz/grimoire/cmd/grimoire@latest

# または、リポジトリをクローンしてビルド
git clone https://github.com/ayutaz/Grimoire.git
cd Grimoire
make build

# または直接ビルド
go build -o grimoire cmd/grimoire/main.go
```

## 🎨 使い方

### 1. 魔法陣を描く

紙に手描き、またはデジタルツールで魔法陣を作成します。必ず外周円で囲んでください。

### サポートされる画像形式

Grimoireは以下の画像形式をサポートしています：
- **PNG** (.png) - 推奨形式、最高品質
- **JPEG** (.jpg, .jpeg) - 写真形式、圧縮により品質が低下する可能性
- **GIF** (.gif) - アニメーションは無視され、最初のフレームのみ使用
- **WebP** (.webp) - モダンな形式、高圧縮率

### 2. 画像をコンパイル・実行

⚠️ **重要**: Windows環境では、ファイル名に日本語や特殊文字を使用しないでください。英数字とアンダースコア、ハイフンのみを使用してください。

```bash
# ✅ 良い例
grimoire run hello_world.png
grimoire run my-magic-circle.png

# ❌ 避けるべき例
grimoire run 魔法陣.png
grimoire run こんにちは.png
```

```bash
# 直接実行
grimoire run magic_circle.png

# Pythonコードに変換
grimoire compile magic_circle.png -o output.py

# デバッグモード
grimoire debug magic_circle.png

# 英語モードで実行（デフォルトは日本語）
grimoire run magic_circle.png --lang en
# または環境変数で設定
export GRIMOIRE_LANG=en
grimoire run magic_circle.png
```

## 📝 プログラム例

### Hello World (シンプルな出力)

```
      ╭─────────╮
    ╱             ╲
   │       ◎       │   <- メインエントリ
   │       |       │
   │       ☆       │   <- 出力
    ╲             ╱
      ╰─────────╯
```

### 加算 (1 + 2)

```
      ╭─────────────╮
    ╱                 ╲
   │         ◎         │
   │      ╱  |  ╲      │
   │    □    □    □    │  <- □(•) + □(••) 
   │    •    ⟐   ••    │  <- 1 + 2
   │         |         │
   │         ☆         │  <- 出力: 3
    ╲                 ╱
      ╰─────────────╯
```

### ループ (3回繰り返し)

```
      ╭────────────────╮
    ╱                    ╲
   │          ◎          │
   │          |          │
   │      □(•••)         │  <- カウンタ: 3
   │          |          │
   │      ╱───⬟───╲      │  <- ループ
   │     │    ☆    │     │  <- ループ内で出力
   │      ╲───────╱      │
    ╲                    ╱
      ╰────────────────╯
```

## 🧪 開発

### テストの実行

```bash
# 全テストを実行
go test ./...

# カバレッジ付き
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# ベンチマークの実行
go test -bench=. ./...

# WebデモのE2Eテスト
make web-test
```

### コードフォーマット

```bash
# Go標準のフォーマット
go fmt ./...

# 静的解析
go vet ./...
```

### ビルド

```bash
# Makefileを使用（推奨）
make build        # 現在のプラットフォーム向けビルド
make build-all    # 全プラットフォーム向けビルド
make test         # テスト実行
make clean        # ビルド成果物のクリーン

# または直接ビルド
go build -o grimoire cmd/grimoire/main.go
```

## 🔧 技術詳細

### Go言語による実装

GrimoireはGo言語で実装されており、以下の特徴があります：

**パフォーマンス特性:**
- **起動時間**: 約0.06秒（高速な起動）
- **実行時間**: 約0.2秒（効率的な画像処理）  
- **バイナリサイズ**: 約10MB（コンパクトな実行ファイル）
- **依存関係**: 外部ライブラリ不要（Pure Go実装）

### 画像認識アプローチ

Go版では外部ライブラリに依存しない、Pure Goでの画像処理を実装：

- **輪郭検出**: Moore近傍探索による輪郭追跡
- **図形認識**: Douglas-Peuckerアルゴリズムによる多角形近似
- **前処理**: Gaussianブラー、適応的二値化、モルフォロジー演算
- **パターン認識**: 図形内部のドット、ライン、クロスパターンの検出

機械学習を使用しないことで、以下の利点があります：
- 決定的な結果（同じ入力→同じ出力）
- 高速な処理（GPUやモデル不要）
- 軽量な実装（数MB程度）
- 完全な説明可能性

## 📋 既知の問題

### 現在の制限事項
- 四角形の検出精度が不安定（円として誤検出される場合がある）
- 斜めの接続線の検出が未実装
- 複雑な演算子記号の認識が不完全
- 複雑なプログラムの解析には制限があります
- ループ内に実行内容がない場合、エラーになります

## 🤝 貢献

プルリクエストを歓迎します！以下のガイドラインに従ってください：

1. テストファースト（TDD）で開発
2. すべてのテストがパスすることを確認
3. カバレッジ80%以上を維持
4. コミットメッセージは日本語可

## 📄 ライセンス

このプロジェクトはApache License 2.0のもとで公開されています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。

## 🏰 哲学

Grimoireは、プログラミングを視覚的で直感的なものにすることを目指しています。文字という抽象的な記号体系から離れ、より原始的で普遍的な図形言語によって、プログラミングの本質を探求します。

すべてのプログラムは魔法陣です。そして、すべての魔法陣は意図と構造を持った芸術作品です。