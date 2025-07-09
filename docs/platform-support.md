# プラットフォーム対応

Grimoireは、Windows、macOS、Linuxの主要なプラットフォームで動作します。

## 対応プラットフォーム

| プラットフォーム | アーキテクチャ | 状態 | 備考 |
|-----------------|---------------|------|------|
| Windows 10/11 | x64 | ✅ 対応 | .exeバイナリ提供 |
| Windows 10/11 | ARM64 | ⚠️ 未テスト | Pythonで実行可能 |
| macOS 12+ | x64 (Intel) | ✅ 対応 | ユニバーサルバイナリ |
| macOS 12+ | ARM64 (M1/M2) | ✅ 対応 | ネイティブ対応 |
| Ubuntu 20.04+ | x64 | ✅ 対応 | AppImage検討中 |
| Ubuntu 20.04+ | ARM64 | ⚠️ 未テスト | Pythonで実行可能 |

## インストール方法

### バイナリ版（推奨）

1. [リリースページ](https://github.com/ayutaz/Grimoire/releases)から最新版をダウンロード
2. プラットフォームに応じて以下を実行：

#### Windows
```powershell
# ダウンロードしたexeファイルを実行
grimoire.exe run magic_circle.png
```

#### macOS
```bash
# 実行権限を付与
chmod +x grimoire

# 初回実行時はセキュリティ許可が必要
./grimoire run magic_circle.png
```

#### Linux
```bash
# 実行権限を付与
chmod +x grimoire

# 実行
./grimoire run magic_circle.png
```

### Python版

すべてのプラットフォームで利用可能：

```bash
# pip経由
pip install grimoire-lang

# または開発版
git clone https://github.com/ayutaz/Grimoire.git
cd Grimoire
pip install -e .
```

## プラットフォーム別の注意事項

### Windows

- **文字エンコーディング**: UTF-8を使用。コマンドプロンプトで文字化けする場合は `chcp 65001` を実行
- **パス区切り文字**: 内部で自動変換されるため、Unix形式のパスも使用可能
- **実行ファイル生成**: `.bat`ファイルが生成される

### macOS

- **セキュリティ**: 初回実行時に「開発元が未確認」の警告が出る場合：
  - システム環境設定 → セキュリティとプライバシー → 「このまま開く」をクリック
  - または `xattr -d com.apple.quarantine grimoire` を実行
- **Rosetta 2**: Intel版バイナリはM1/M2 Macでも動作

### Linux

- **依存関係**: OpenCVが必要な場合、システムパッケージのインストールが必要：
  ```bash
  # Ubuntu/Debian
  sudo apt-get install python3-opencv
  
  # Fedora
  sudo dnf install python3-opencv
  ```

## ビルド方法

### 開発環境のセットアップ

```bash
# リポジトリをクローン
git clone https://github.com/ayutaz/Grimoire.git
cd Grimoire

# 開発用依存関係をインストール
pip install -r requirements-dev.txt
```

### バイナリのビルド

```bash
# 自動ビルドスクリプト
python scripts/build.py

# または手動でPyInstaller
pyinstaller grimoire.spec
```

### GitHub Actionsでの自動ビルド

タグをプッシュすると自動的に全プラットフォーム用のバイナリがビルドされます：

```bash
git tag v1.0.0
git push origin v1.0.0
```

## トラブルシューティング

### Windows

**問題**: `grimoire.exe`が動作しない
- **解決策**: Windows Defenderやアンチウイルスソフトが誤検知している可能性。除外リストに追加

**問題**: 日本語が文字化けする
- **解決策**: `chcp 65001`を実行してUTF-8モードに切り替え

### macOS

**問題**: 「開発元が未確認」エラー
- **解決策**: 
  ```bash
  xattr -d com.apple.quarantine grimoire
  # または右クリック → 開く
  ```

**問題**: `dyld: Library not loaded`エラー
- **解決策**: システムライブラリの更新
  ```bash
  brew update && brew upgrade
  ```

### Linux

**問題**: `GLIBC_X.XX not found`エラー
- **解決策**: より新しいディストリビューションでビルドされたバイナリ。Pythonから実行を推奨

**問題**: OpenCVのインポートエラー
- **解決策**: 
  ```bash
  sudo apt-get install libopencv-dev python3-opencv
  ```

## パフォーマンス最適化

### プラットフォーム別の最適化

- **Windows**: Windows Defenderのリアルタイムスキャンから除外
- **macOS**: Metal Performance Shadersを活用（将来対応予定）
- **Linux**: GPU処理にCUDAを使用可能（opencv-python-headless推奨）

### 画像処理の高速化

大きな魔法陣を処理する場合：

1. 画像サイズを適切に調整（推奨: 600x600ピクセル）
2. グレースケール変換で処理を高速化
3. NumPyの最適化版を使用（`pip install numpy[mkl]`）

## 今後の対応予定

- [ ] Windows ARM64ネイティブ対応
- [ ] Linux AppImage形式での配布
- [ ] macOS用.appバンドル
- [ ] WebAssembly版（ブラウザ実行）
- [ ] モバイル対応（iOS/Android）