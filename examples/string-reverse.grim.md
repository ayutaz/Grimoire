# 文字列反転 - String Reverse

## 概要 - Overview
文字列を逆順に並び替えるプログラムです。複数の実装方法を示しています。
This program reverses a string. Multiple implementation methods are demonstrated.

## 魔法陣の構造 - Magic Circle Structure

### 外周円 - Outer Circle
プログラム全体を囲む大きな円

### メインエントリ - Main Entry (二重円)
プログラムの開始点を示す二重円シンボル

### 変数定義 - Variable Definition
- `input_string`: 入力文字列
- `reversed_string`: 反転後の文字列
- `length`: 文字列の長さ
- `i`: ループカウンタ
- `char_array`: 文字配列（方法2用）

### 実装方法 - Implementation Methods
1. **スライス法**: Pythonのスライス記法を使用
2. **ループ法**: 文字を後ろから1つずつ結合
3. **配列交換法**: 文字配列の両端から交換

## 生成されるPythonコード - Generated Python Code

```python
def reverse_string_slice(s):
    """スライスを使った文字列反転"""
    return s[::-1]

def reverse_string_loop(s):
    """ループを使った文字列反転"""
    reversed_string = ""
    for char in s:
        reversed_string = char + reversed_string
    return reversed_string

def reverse_string_swap(s):
    """配列の要素交換による文字列反転"""
    char_array = list(s)
    left = 0
    right = len(char_array) - 1
    
    while left < right:
        # 左右の文字を交換
        char_array[left], char_array[right] = char_array[right], char_array[left]
        left += 1
        right -= 1
    
    return ''.join(char_array)

def main():
    # テストケース
    test_strings = [
        "Hello, World!",
        "Grimoire",
        "魔法陣プログラミング",
        "12345",
        "A",
        ""
    ]
    
    print("=== 文字列反転のデモンストレーション ===\n")
    
    for original in test_strings:
        print(f"元の文字列: '{original}'")
        print(f"  スライス法: '{reverse_string_slice(original)}'")
        print(f"  ループ法: '{reverse_string_loop(original)}'")
        print(f"  交換法: '{reverse_string_swap(original)}'")
        print()

if __name__ == "__main__":
    main()
```

## 実行結果 - Execution Result
```
=== 文字列反転のデモンストレーション ===

元の文字列: 'Hello, World!'
  スライス法: '!dlroW ,olleH'
  ループ法: '!dlroW ,olleH'
  交換法: '!dlroW ,olleH'

元の文字列: 'Grimoire'
  スライス法: 'eriomirG'
  ループ法: 'eriomirG'
  交換法: 'eriomirG'

元の文字列: '魔法陣プログラミング'
  スライス法: 'グンミラグロプ陣法魔'
  ループ法: 'グンミラグロプ陣法魔'
  交換法: 'グンミラグロプ陣法魔'

元の文字列: '12345'
  スライス法: '54321'
  ループ法: '54321'
  交換法: '54321'

元の文字列: 'A'
  スライス法: 'A'
  ループ法: 'A'
  交換法: 'A'

元の文字列: ''
  スライス法: ''
  ループ法: ''
  交換法: ''
```

## 解説 - Explanation
各実装方法の特徴：

1. **スライス法**
   - 最も簡潔でPythonic
   - 時間複雑度: O(n)
   - 空間複雑度: O(n)

2. **ループ法**
   - 文字列の不変性を利用
   - 理解しやすい実装
   - 時間複雑度: O(n)
   - 空間複雑度: O(n)

3. **交換法**
   - インプレース風の実装
   - 配列操作の基本を示す
   - 時間複雑度: O(n)
   - 空間複雑度: O(n)

Grimoireでは、異なるアルゴリズムは魔法陣内の異なるパスとして表現され、それぞれが独立した処理フローを持ちます。

## 応用例 - Applications
- パリンドローム（回文）の判定
- 文字列処理アルゴリズムの基礎
- データ変換処理
- 暗号化/復号化の前処理