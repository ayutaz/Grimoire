# バブルソート - Bubble Sort

## 概要 - Overview
バブルソートアルゴリズムの実装例です。配列の要素を昇順に並べ替えます。
This example demonstrates the bubble sort algorithm, sorting array elements in ascending order.

## 魔法陣の構造 - Magic Circle Structure

### 外周円 - Outer Circle
プログラム全体を囲む大きな円

### メインエントリ - Main Entry (二重円)
プログラムの開始点を示す二重円シンボル

### 変数定義 - Variable Definition
- `array`: ソート対象の配列 [5, 2, 8, 1, 9]
- `n`: 配列の長さ
- `i`, `j`: ループカウンタ
- `temp`: 交換用の一時変数

### 制御フロー - Control Flow
1. 外側のループ（i = 0 to n-1）
2. 内側のループ（j = 0 to n-i-2）
3. 条件分岐：array[j] > array[j+1] の場合
4. 要素の交換処理

## 生成されるPythonコード - Generated Python Code

```python
def main():
    # 配列の初期化
    array = [5, 2, 8, 1, 9]
    n = len(array)
    
    # バブルソート
    for i in range(n):
        for j in range(n - i - 1):
            if array[j] > array[j + 1]:
                # 要素の交換
                temp = array[j]
                array[j] = array[j + 1]
                array[j + 1] = temp
    
    # 結果の出力
    print("Sorted array:", array)

if __name__ == "__main__":
    main()
```

## 実行結果 - Execution Result
```
Sorted array: [1, 2, 5, 8, 9]
```

## 解説 - Explanation
バブルソートは隣接する要素を比較し、必要に応じて交換することで配列をソートします。
- 時間複雑度: O(n²)
- 空間複雑度: O(1)
- 安定なソートアルゴリズム

Grimoireでは、ネストしたループは魔法陣内の同心円として表現され、条件分岐は分岐線で表現されます。

## 応用例 - Applications
- 小規模なデータセットのソート
- 教育目的でのアルゴリズム学習
- ほぼソート済みのデータに対する最適化版の実装