# Hello World Example

## 魔法陣の説明

最も基本的なGrimoireプログラムです。画面に星(☆)を表示します。
すべてのプログラムは外周円を持つ魔法陣として記述される必要があります。

## 構造

```
     ╱─────╲
   ╱    ◎   ╲    ← 外周円（必須）
  │     │     │
  │     ☆     │   ← 星を表示
   ╲         ╱
     ╲─────╱
```

## 実行結果

画面に星記号が表示されます。

## 解説

1. 外周円: 魔法陣の境界（必須要素）- すべてのプログラムはこの円の内部に封じられる
2. `◎` (二重円): プログラムのエントリーポイント（メインエントリ）
3. `│` (接続線): エネルギーの流れを示す
4. `☆` (5点星): 出力/表示操作 - 魔法の具現化

## 描き方のポイント

- **必ず外周円から描き始める** - これが魔法陣の基本
- 外周円は完全に閉じていること（エネルギーが漏れないように）
- 二重円は魔法陣の中心に配置
- 接続線は垂直にまっすぐ引く
- 星は5つの頂点を持つように描く