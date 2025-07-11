# Platform Support / プラットフォームサポート

## Supported Platforms / サポートプラットフォーム

Grimoire provides pre-built binaries for the following platforms:

Grimoireは以下のプラットフォーム向けにビルド済みバイナリを提供しています：

### Desktop Platforms / デスクトッププラットフォーム

| Platform | Architecture | Binary Name | Tested |
|----------|-------------|-------------|---------|
| macOS | x86_64 (Intel) | `grimoire-darwin-amd64` | ✅ |
| macOS | ARM64 (Apple Silicon) | `grimoire-darwin-arm64` | ✅ |
| Linux | x86_64 | `grimoire-linux-amd64` | ✅ |
| Linux | ARM64 | `grimoire-linux-arm64` | ✅ |
| Windows | x86_64 | `grimoire-windows-amd64.exe` | ✅ |

## Binary Sizes / バイナリサイズ

Optimized release builds with stripped symbols:
シンボル削除済みの最適化リリースビルド：

- **macOS Intel**: ~2.6MB
- **macOS ARM**: ~2.8MB  
- **Linux x86_64**: ~2.4MB
- **Linux ARM64**: ~2.5MB
- **Windows x86_64**: ~2.6MB

## System Requirements / システム要件

### Minimum Requirements / 最小要件
- **OS**: Windows 10+, macOS 10.15+, Linux (kernel 3.0+)
- **Memory**: 64MB RAM
- **Storage**: 10MB free space

### Runtime Dependencies / 実行時依存関係
- **None!** Grimoire is a statically linked binary with no external dependencies
- **なし！** Grimoireは静的リンクされたバイナリで、外部依存関係はありません

## Installation / インストール

### Pre-built Binaries / ビルド済みバイナリ

1. Download the appropriate binary from [Releases](https://github.com/ayutaz/Grimoire/releases)
2. Make it executable (Unix/Linux/macOS):
   ```bash
   chmod +x grimoire-*
   ```
3. Move to your PATH:
   ```bash
   sudo mv grimoire-* /usr/local/bin/grimoire
   ```

### Build from Source / ソースからビルド

Requirements:
- Go 1.21 or higher

```bash
# Clone the repository
git clone https://github.com/ayutaz/Grimoire.git
cd Grimoire

# Build for your platform
make build

# Build for all platforms
make build-all

# Build optimized release binaries
make build-release
```

## Cross Compilation / クロスコンパイル

Grimoire uses Go's built-in cross-compilation support:

```bash
# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o grimoire-linux-amd64 cmd/grimoire/main.go

# Available GOOS values: darwin, linux, windows
# Available GOARCH values: amd64, arm64, arm, 386
```

## Testing on Different Platforms / 異なるプラットフォームでのテスト

The CI/CD pipeline automatically tests on:
- Ubuntu Latest (Linux x86_64)
- macOS Latest (macOS x86_64)
- Windows Latest (Windows x86_64)

## Known Issues / 既知の問題

### Windows
- File paths with non-ASCII characters may cause issues
- Use forward slashes (/) or escaped backslashes (\\) in paths

### Linux
- Some distributions may require `libX11` for clipboard operations (not used by Grimoire)

### macOS
- First run may require security approval in System Preferences

## Reporting Platform Issues / プラットフォーム問題の報告

If you encounter platform-specific issues:

1. Check the [existing issues](https://github.com/ayutaz/Grimoire/issues)
2. Create a new issue with:
   - Platform details (OS, version, architecture)
   - Error messages
   - Steps to reproduce
   - `grimoire --version` output