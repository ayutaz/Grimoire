# Grimoire Web Demo E2E Tests

このディレクトリには、Grimoire WebデモのE2E（End-to-End）テストが含まれています。

## セットアップ

```bash
# 依存関係のインストール
npm install
```

## テストの実行

```bash
# WASMをビルドしてテストを実行
./run-tests.sh

# または個別に実行
npm test

# ヘッドレスブラウザでテストを実行
npm run test:headed

# デバッグモードでテストを実行
npm run test:debug
```

## CI/CD

GitHub Actionsで自動的に実行されます。以下の条件でトリガーされます：

- `main`ブランチへのプッシュ
- プルリクエスト
- Web関連のファイルが変更された場合

## テストの内容

### web-demo.spec.js
- ページの基本的な動作
- サンプル画像の処理
- ファイルアップロード
- タブ切り替え
- エラーハンドリング

### wasm-integration.spec.js
- WASM関数の直接呼び出し
- シンボル検出の確認
- エラーハンドリング
- デバッグ情報の確認
- すべてのサンプル画像の処理

## トラブルシューティング

### テストが失敗する場合

1. WASMが正しくビルドされているか確認
   ```bash
   make web-build
   ```

2. ローカルサーバーが起動しているか確認
   ```bash
   npx http-server ../static -p 8080
   ```

3. ブラウザのコンソールでエラーを確認
   ```bash
   npm run test:headed
   ```