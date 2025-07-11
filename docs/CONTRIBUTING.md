# 貢献ガイドライン

## Git設定

このリポジトリでは、統一された作者情報を使用しています。

### ローカル設定

リポジトリをクローンした後、以下のコマンドを実行してください：

```bash
git config --local user.name "ayutaz"
git config --local user.email "ka1357amnbpdr@gmail.com"
```

または、リポジトリのルートディレクトリで以下のコマンドを実行：

```bash
# 設定の確認
git config --local --list | grep user

# 設定が異なる場合は以下を実行
git config --local user.name "ayutaz"
git config --local user.email "ka1357amnbpdr@gmail.com"
```

### コミットの作成

すべてのコミットは以下の作者情報で作成されます：
- 名前: `ayutaz`
- メールアドレス: `ka1357amnbpdr@gmail.com`

### プルリクエスト

1. feature/ブランチを作成
2. 変更をコミット
3. プルリクエストを作成

## コーディング規約

### Go言語
- `go fmt`でフォーマット
- `go vet`でリント
- テストカバレッジ80%以上を維持

### Python
- `ruff`でフォーマットとリント
- 型ヒントを使用
- docstringを記載

## テスト

```bash
# Go
go test ./...

# Python
uv run pytest
```