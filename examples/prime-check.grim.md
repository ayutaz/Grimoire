# 素数判定 - Prime Number Check

## 概要 - Overview
与えられた数が素数かどうかを判定するプログラムです。エラトステネスの篩の簡易版を実装しています。
This program checks whether a given number is prime. It implements a simple version of the Sieve of Eratosthenes.

## 魔法陣の構造 - Magic Circle Structure

### 外周円 - Outer Circle
プログラム全体を囲む大きな円

### メインエントリ - Main Entry (二重円)
プログラムの開始点を示す二重円シンボル

### 変数定義 - Variable Definition
- `number`: 判定対象の数（例: 17）
- `is_prime`: 素数判定フラグ（True/False）
- `i`: ループカウンタ
- `sqrt_n`: numberの平方根（最適化のため）

### 制御フロー - Control Flow
1. 特殊ケースの処理（number <= 1）
2. 2の場合の処理
3. 偶数の場合の処理
4. 3以上の奇数での除算チェック（平方根まで）

## 生成されるPythonコード - Generated Python Code

```python
import math

def is_prime_number(number):
    # 1以下の数は素数ではない
    if number <= 1:
        return False
    
    # 2は素数
    if number == 2:
        return True
    
    # 偶数は素数ではない（2を除く）
    if number % 2 == 0:
        return False
    
    # 3以上の奇数で割り切れるかチェック
    sqrt_n = int(math.sqrt(number))
    for i in range(3, sqrt_n + 1, 2):
        if number % i == 0:
            return False
    
    return True

def main():
    # テストケース
    test_numbers = [2, 3, 4, 17, 20, 29, 100, 101]
    
    for num in test_numbers:
        if is_prime_number(num):
            print(f"{num} は素数です")
        else:
            print(f"{num} は素数ではありません")

if __name__ == "__main__":
    main()
```

## 実行結果 - Execution Result
```
2 は素数です
3 は素数です
4 は素数ではありません
17 は素数です
20 は素数ではありません
29 は素数です
100 は素数ではありません
101 は素数です
```

## 解説 - Explanation
素数判定の最適化ポイント：
1. **偶数の除外**: 2以外の偶数は素数ではない
2. **平方根までのチェック**: nの約数は√n以下に必ず存在する
3. **奇数のみチェック**: 偶数は既に除外済みなので、3以上の奇数のみで除算

時間複雑度: O(√n)

Grimoireでは、条件分岐は魔法陣内の分岐パスとして表現され、早期リターンは特別な終了シンボルで表現されます。

## 応用例 - Applications
- 暗号化アルゴリズムの基礎
- 数学的な問題解決
- より高度な素数生成アルゴリズムへの拡張
- エラトステネスの篩の完全実装