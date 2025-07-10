# Grimoire Go移行計画

## 概要
PythonベースのGrimoireプロジェクトをGoに完全移行する計画書です。
この移行により、実行速度を750倍高速化（15秒→0.02秒）し、クロスプラットフォーム配布を簡素化します。

## 移行の目的
1. **パフォーマンス向上**: バイナリ実行時間を15秒から0.02秒に短縮
2. **配布の簡素化**: 単一バイナリで全プラットフォーム対応
3. **依存関係の削減**: OpenCV不要のPure Go実装
4. **開発効率の向上**: シンプルなビルドとデプロイ

## フェーズ1: プロジェクト構造の設計（1-2日）

### 新しいディレクトリ構造
```
grimoire/
├── cmd/
│   └── grimoire/
│       └── main.go          # メインエントリポイント
├── internal/
│   ├── detector/           # 画像認識
│   │   ├── detector.go
│   │   ├── shapes.go
│   │   └── detector_test.go
│   ├── parser/             # AST構築
│   │   ├── parser.go
│   │   ├── ast.go
│   │   └── parser_test.go
│   ├── compiler/           # コード生成
│   │   ├── compiler.go
│   │   ├── python.go
│   │   └── compiler_test.go
│   └── cli/               # CLIインターフェース
│       └── cli.go
├── pkg/
│   └── grimoire/          # 公開API
│       └── grimoire.go
├── examples/              # 既存のサンプル画像
├── docs/                  # ドキュメント
├── scripts/              # ビルドスクリプト
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## フェーズ2: 基本セットアップ（0.5日）

### 1. Go モジュールの初期化
```bash
go mod init github.com/ayutaz/grimoire
```

### 2. 必要なパッケージ
```go
// 画像処理
github.com/disintegration/imaging
golang.org/x/image/draw

// CLI
github.com/spf13/cobra
github.com/spf13/viper

// テスト
github.com/stretchr/testify

// ユーティリティ
github.com/pkg/errors
```

## フェーズ3: コア機能の実装（3-5日）

### 1. 画像認識モジュール (detector)
- [ ] 基本的な画像読み込み
- [ ] 二値化処理
- [ ] 輪郭検出アルゴリズム
- [ ] 形状分類（円、四角、三角、星など）
- [ ] シンボル間の接続検出
- [ ] 内部パターン認識（ドット、線など）

### 2. パーサーモジュール (parser)
- [ ] AST定義（Go構造体）
- [ ] シンボルからASTへの変換
- [ ] 文法規則の実装
- [ ] エラーハンドリング

### 3. コンパイラモジュール (compiler)
- [ ] Python コード生成
- [ ] インデント管理
- [ ] 将来的な他言語対応の準備

### 4. CLIインターフェース
- [ ] runコマンド
- [ ] compileコマンド
- [ ] debugコマンド
- [ ] ヘルプとバージョン情報

## フェーズ4: テストの実装（2日）

### 1. ユニットテスト
- [ ] 各モジュールの単体テスト
- [ ] テストカバレッジ80%以上

### 2. 統合テスト
- [ ] エンドツーエンドテスト
- [ ] 既存のサンプル画像でのテスト

### 3. ベンチマーク
- [ ] パフォーマンステスト
- [ ] メモリ使用量測定

## フェーズ5: ビルドとCI/CD（1日）

### 1. Makefile
```makefile
.PHONY: build test clean

build:
	go build -o grimoire cmd/grimoire/main.go

build-all:
	GOOS=darwin GOARCH=amd64 go build -o dist/grimoire-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o dist/grimoire-darwin-arm64
	GOOS=linux GOARCH=amd64 go build -o dist/grimoire-linux-amd64
	GOOS=windows GOARCH=amd64 go build -o dist/grimoire-windows-amd64.exe

test:
	go test -v ./...

benchmark:
	go test -bench=. ./...
```

### 2. GitHub Actions更新
- [ ] Go用のワークフロー作成
- [ ] クロスプラットフォームビルド
- [ ] 自動リリース

## フェーズ6: ドキュメントとクリーンアップ（1日）

### 1. ドキュメント更新
- [ ] README.mdの更新
- [ ] インストール手順
- [ ] API ドキュメント

### 2. 移行完了
- [ ] Pythonコードの削除
- [ ] 不要な依存関係の削除
- [ ] プルリクエストの作成

## 成功指標

1. **パフォーマンス**
   - Hello World実行: < 0.1秒
   - メモリ使用量: < 20MB
   - バイナリサイズ: < 10MB

2. **機能**
   - 全サンプル画像が正しく動作
   - エラーハンドリングが適切

3. **品質**
   - テストカバレッジ: > 80%
   - golintエラー: 0
   - go vetエラー: 0

## リスクと対策

1. **画像処理の精度**
   - リスク: OpenCVと比べて精度が落ちる可能性
   - 対策: 十分なテストと調整

2. **開発期間**
   - リスク: 予定より長くかかる可能性
   - 対策: MVPを優先し、段階的に機能追加

## タイムライン

- **週1**: プロジェクト構造とコア機能
- **週2**: テスト、ビルド、ドキュメント
- **合計**: 約2週間

## 次のステップ

1. このプランのレビューと承認
2. Go プロジェクトの初期化
3. 最初のコミット