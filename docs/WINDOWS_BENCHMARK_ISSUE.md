# Windows環境でのベンチマークテスト実行問題

## 問題の概要

GitHub ActionsのWindows環境で `go test -bench=. -benchmem ./...` を実行すると、「no Go files in D:\a\Grimoire\Grimoire」というエラーが発生する問題があります。

## 原因

この問題はPowerShellの引数解析に関する既知の問題（[Go issue #43179](https://github.com/golang/go/issues/43179)）に起因します。PowerShellは `-bench=.` のような、`=.` で終わるフラグを正しく解析できません。

## 影響

- 通常のテスト (`go test -v ./...`) は問題なく動作します
- ベンチマークテストのみが影響を受けます
- Linux/macOS環境では問題ありません

## 解決策

### 1. cmd.exeを使用する（推奨）

PowerShellの代わりにcmd.exeを使用します：

```yaml
- name: Run benchmarks (Windows)
  if: runner.os == 'Windows'
  shell: cmd
  run: go test -bench=. -benchmem ./...
```

### 2. PowerShellでクォートを使用

```yaml
- name: Run benchmarks (Windows)
  if: runner.os == 'Windows'
  shell: pwsh
  run: go test -bench="." -benchmem ./...
```

### 3. スペースで区切る

```yaml
- name: Run benchmarks (Windows)
  if: runner.os == 'Windows'
  shell: pwsh
  run: go test -bench . -benchmem ./...
```

### 4. パッケージごとに実行

```yaml
- name: Run benchmarks (Windows)
  if: runner.os == 'Windows'
  shell: pwsh
  run: |
    $packages = go list ./...
    foreach ($pkg in $packages) {
      go test -bench="." -benchmem $pkg
    }
```

## 実装された修正

1. **GitHub Actions**: `.github/workflows/go.yml` でWindows環境でもベンチマークを実行するように修正
2. **Makefile**: Windows環境を検出してcmd.exeを使用するように修正
3. **テストワークフロー**: `.github/workflows/go-benchmark-windows-fix.yml` で各解決策をテスト

## 参考リンク

- [golang/go#43179: cmd/go: flags ending with "=." is not correctly parsed by go tool when run via powershell](https://github.com/golang/go/issues/43179)
- [golang/go#23053: cmd/go: go test -bench ./subpackage tests current package instead](https://github.com/golang/go/issues/23053)