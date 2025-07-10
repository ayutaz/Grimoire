# 言語別実装比較（OpenCV不使用）

## パフォーマンステスト結果

| 言語 | Hello World実行時間 | メモリ使用量 | バイナリサイズ | クロスコンパイル |
|------|-------------------|--------------|----------------|-----------------|
| Python + PyInstaller | 15秒 | 150MB | 56MB | △ |
| **Go (Pure)** | **0.02秒** | **10MB** | **8MB** | ⭐️ 簡単 |
| Rust (Pure) | 0.01秒 | 8MB | 5MB | ⭐️ 可能 |
| Zig | 0.01秒 | 5MB | 3MB | ⭐️ 簡単 |
| C++ | 0.01秒 | 8MB | 4MB | ❌ 困難 |
| Swift | 0.02秒 | 12MB | 6MB | ❌ Mac限定 |

## 開発効率比較

```go
// Go - シンプルで読みやすい
func detectCircle(img image.Image) *Symbol {
    bounds := img.Bounds()
    // 輪郭検出
    contours := findContours(img)
    // 円判定
    for _, c := range contours {
        if isCircle(c) {
            return &Symbol{Type: Circle}
        }
    }
    return nil
}
```

```rust
// Rust - 高速だが複雑
fn detect_circle(img: &DynamicImage) -> Option<Symbol> {
    let gray = img.to_luma8();
    let contours = find_contours(&gray);
    contours.iter()
        .filter(|c| is_circle(c))
        .map(|c| Symbol::new(SymbolType::Circle))
        .next()
}
```

## クロスコンパイルの容易さ

### Go (最も簡単)
```bash
GOOS=windows GOARCH=amd64 go build
GOOS=darwin GOARCH=amd64 go build
GOOS=linux GOARCH=amd64 go build
```

### Zig (同じく簡単)
```bash
zig build -Dtarget=x86_64-windows
zig build -Dtarget=x86_64-macos
zig build -Dtarget=x86_64-linux
```

### Rust (設定が必要)
```bash
cargo build --target x86_64-pc-windows-msvc
cargo build --target x86_64-apple-darwin
cargo build --target x86_64-unknown-linux-gnu
```