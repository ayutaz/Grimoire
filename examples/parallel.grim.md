# 並列処理の例

## 基本的な並列実行

```
◎ main
└─ ⬢ parallel_tasks
   ├─ ○ download("file1.txt")
   ├─ ○ download("file2.txt")
   └─ ○ download("file3.txt")
   │
   └─⬢─→ ☆ "All downloads complete"
```

## MapReduce パターン

```
◎ main
├─ □ []data = [1,2,3,4,5,6,7,8,9,10]
│
├─ ⬢ map_phase
│  ├─ λ○ (x) → x * x
│  └─ []data
│  │
│  └─⬢─→ □ []squared
│
└─ ⬢ reduce_phase
   ├─ λ○ (a,b) → a + b
   └─ []squared
   │
   └─⬢─→ ☆ "Sum of squares: " + result
```

## 同期付き並列処理

```
◎ main
├─ □ #shared_counter = 0
├─ ═══ mutex ═══  ← 同期プリミティブ
│
└─ ⬢ workers(10)
   └─ ⬟ (100)  ← 各ワーカーが100回ループ
      └─ ╔═══╗
         ║lock║
         ╚═╤═╝
           │
      □ #shared_counter++
           │
         ╔═╧═╗
         ║unlock║
         ╚═══╝
```

## Producer-Consumer パターン

```
○ producer
└─ ⬟ (∞)
   ├─ □ item = generate()
   └─ ➤ queue.push(item)  ← キューに追加

○ consumer
└─ ⬟ (∞)
   ├─ ➤ item = queue.pop()  ← キューから取得
   └─ process(item)

◎ main
└─ ⬢ system
   ├─ ○ producer × 2  ← 2つのプロデューサー
   ├─ ═══ queue ═══   ← 共有キュー
   └─ ○ consumer × 4  ← 4つのコンシューマー
```

## 非同期待機

```
◎ main
├─ ⬢ async_calls
│  ├─ ○ fetch_user_data()
│  ├─ ○ fetch_posts()
│  └─ ○ fetch_comments()
│  │
│  └─⬢─→ await all
│         │
└─────────→ ☆ render_page(user, posts, comments)
```

## 解説

### 並列実行シンボル
- `⬢`: 六角形は並列実行を開始
- 上部の六角形: タスクを分散
- 下部の六角形: 結果を収集（バリア同期）

### 同期プリミティブ
- `═══`: 太い二重線はミューテックスやロック
- `➤`: キューやチャネル操作

### パターン
1. **Fire and Forget**: 結果を待たない並列実行
2. **Fork-Join**: 分散して結果を収集
3. **Pipeline**: ステージ間でデータを流す
4. **Worker Pool**: 固定数のワーカーでタスク処理

## 実行時の挙動

コンパイラは六角形を認識すると：
1. スレッドプールまたはgoroutineを生成
2. 各分岐を並列実行
3. 必要に応じて同期を挿入
4. 結果を収集して続行