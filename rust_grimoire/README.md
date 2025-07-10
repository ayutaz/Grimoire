# Rust版 Grimoire

## セットアップ

### 1. OpenCVのインストール

**macOS:**
```bash
brew install opencv
export OPENCV_LINK_LIBS=opencv4
export OPENCV_LINK_PATHS=/opt/homebrew/lib,/usr/local/lib
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y libopencv-dev clang libclang-dev
```

**Windows:**
```bash
# chocolateyを使用
choco install opencv
# または vcpkgを使用
vcpkg install opencv4
```

### 2. 環境変数の設定

```bash
# Linux/macOS
export OPENCV_LINK_LIBS=opencv4
export OPENCV_INCLUDE_PATHS=/usr/include/opencv4

# Windows
set OPENCV_LINK_LIBS=opencv_world4
set OPENCV_INCLUDE_PATHS=C:\tools\opencv\build\include
```

### 3. ビルド

```bash
# 開発ビルド
cargo build

# リリースビルド（最適化）
cargo build --release

# 実行
cargo run --release -- run examples/images/hello_world.png
```

## クロスコンパイル

### GitHub Actionsでの自動ビルド

`.github/workflows/build.yml`を使用して、各プラットフォーム用のバイナリを自動生成します。

### ローカルでのクロスコンパイル

```bash
# ターゲットを追加
rustup target add x86_64-unknown-linux-gnu
rustup target add x86_64-apple-darwin
rustup target add x86_64-pc-windows-msvc

# 各プラットフォーム用にビルド
cargo build --release --target x86_64-unknown-linux-gnu
```

## パフォーマンス最適化

1. **LTO（Link Time Optimization）有効**
2. **単一コード生成ユニット**
3. **ストリップによるサイズ削減**
4. **SIMD最適化**

## 注意点

- OpenCVは動的リンクされるため、実行環境にもOpenCVが必要
- 静的リンクする場合は、OpenCVを静的ライブラリとしてビルドする必要がある
- Windows版は`opencv_world`を使用することが多い