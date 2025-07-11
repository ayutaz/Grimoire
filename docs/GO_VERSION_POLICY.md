# Go Version Policy / Goバージョンポリシー

## Overview / 概要

This document outlines the Go version support policy for the Grimoire project.

このドキュメントは、GrimoireプロジェクトのGoバージョンサポートポリシーについて説明します。

## Version Requirements / バージョン要件

### Minimum Supported Version / 最低サポートバージョン
- **Go 1.21** (Released August 2023)

### Recommended Development Version / 推奨開発バージョン
- **Go 1.23** (Released August 2024)

### CI/CD Test Matrix / CI/CDテストマトリックス
- Go 1.21 (最低保証)
- Go 1.22 (互換性確認)
- Go 1.23 (推奨/最新)

## Rationale / 採用理由

### Why Go 1.21 as Minimum? / なぜGo 1.21を最低版とするか？

1. **Standard Library Enhancements / 標準ライブラリの強化**
   - `slices` package - スライス操作の標準化
   - `maps` package - マップ操作の標準化
   - `slog` package - 構造化ログの標準サポート

2. **Performance Improvements / パフォーマンス向上**
   - Improved generics performance / ジェネリクスの性能改善
   - Better memory management / メモリ管理の改善

3. **Ecosystem Compatibility / エコシステムの互換性**
   - Widely supported in enterprise environments / 企業環境で広くサポート
   - Available on major cloud platforms / 主要クラウドプラットフォームで利用可能

### Why Go 1.23 for Development? / なぜGo 1.23を開発に使うか？

1. **Latest Features / 最新機能**
   - Latest language improvements / 最新の言語改善
   - Security updates / セキュリティアップデート
   - Performance optimizations / パフォーマンス最適化

2. **Developer Experience / 開発者体験**
   - Better tooling support / より良いツーリングサポート
   - Improved error messages / 改善されたエラーメッセージ

## Migration Strategy / 移行戦略

### Upgrading Minimum Version / 最低バージョンの更新

We will update the minimum version when:
- The current minimum version reaches end of support
- A critical feature becomes available in a newer version
- Major dependencies require a newer version

最低バージョンは以下の場合に更新します：
- 現在の最低バージョンがサポート終了になった場合
- 新しいバージョンで重要な機能が利用可能になった場合
- 主要な依存関係がより新しいバージョンを必要とする場合

### Version Support Timeline / バージョンサポートタイムライン

- **Go 1.21**: Minimum support until Go 1.25 is released (estimated February 2025)
- **Go 1.22**: Added to test matrix when it becomes widely adopted
- **Go 1.23**: Current development version
- **Go 1.24**: Will be evaluated when released (estimated February 2025)

## Development Guidelines / 開発ガイドライン

1. **Feature Usage / 機能の使用**
   - Only use features available in Go 1.21
   - Document any version-specific code
   - Go 1.21で利用可能な機能のみを使用
   - バージョン固有のコードは文書化

2. **Testing / テスト**
   - All code must pass tests on Go 1.21-1.23
   - Primary development on Go 1.23
   - すべてのコードはGo 1.21-1.23でテストに合格する必要
   - 主な開発はGo 1.23で実施

3. **Dependencies / 依存関係**
   - Ensure all dependencies support Go 1.21
   - Regularly update dependencies
   - すべての依存関係がGo 1.21をサポートすることを確認
   - 定期的に依存関係を更新

## Checking Your Version / バージョンの確認

```bash
# Check your Go version / Goバージョンを確認
go version

# Install specific version (using Go version manager) / 特定バージョンをインストール
# Example with 'g' version manager:
g install 1.23.0
g use 1.23.0
```

## References / 参考資料

- [Go Release Policy](https://go.dev/doc/devel/release)
- [Go Version History](https://go.dev/doc/devel/release)
- [Go Compatibility Promise](https://go.dev/doc/go1compat)