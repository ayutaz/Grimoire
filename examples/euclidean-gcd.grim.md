# ユークリッドの互除法 - Euclidean Algorithm (GCD)

## 概要 - Overview
2つの整数の最大公約数（GCD: Greatest Common Divisor）を求めるユークリッドの互除法の実装です。再帰版と反復版の両方を示します。
This example implements the Euclidean algorithm to find the Greatest Common Divisor (GCD) of two integers. Both recursive and iterative versions are shown.

## 魔法陣の構造 - Magic Circle Structure

### 外周円 - Outer Circle
プログラム全体を囲む大きな円

### メインエントリ - Main Entry (二重円)
プログラムの開始点を示す二重円シンボル

### 関数定義 - Function Definitions
1. **再帰版**: 自己参照を示す特別な円形パターン
2. **反復版**: ループを示す渦巻きパターン

### 変数定義 - Variable Definition
- `a`, `b`: 入力される2つの整数
- `remainder`: 余り
- `temp`: 一時変数

### 制御フロー - Control Flow
1. b が 0 になるまで繰り返し
2. a を b で割った余りを計算
3. a に b を、b に余りを代入

## 生成されるPythonコード - Generated Python Code

```python
def gcd_recursive(a, b):
    """再帰版のユークリッドの互除法"""
    if b == 0:
        return a
    return gcd_recursive(b, a % b)

def gcd_iterative(a, b):
    """反復版のユークリッドの互除法"""
    while b != 0:
        remainder = a % b
        a = b
        b = remainder
    return a

def gcd_with_steps(a, b):
    """計算過程を表示する版"""
    print(f"\nGCD({a}, {b})の計算過程:")
    print(f"{'a':>10} {'b':>10} {'余り':>10}")
    print("-" * 35)
    
    original_a, original_b = a, b
    
    while b != 0:
        remainder = a % b
        print(f"{a:>10} {b:>10} {remainder:>10}")
        a = b
        b = remainder
    
    print(f"\nGCD({original_a}, {original_b}) = {a}")
    return a

def lcm(a, b):
    """最小公倍数（LCM）を計算"""
    return abs(a * b) // gcd_iterative(a, b)

def extended_gcd(a, b):
    """拡張ユークリッドの互除法
    ax + by = gcd(a,b) となる x, y も求める"""
    if b == 0:
        return a, 1, 0
    
    gcd, x1, y1 = extended_gcd(b, a % b)
    x = y1
    y = x1 - (a // b) * y1
    
    return gcd, x, y

def main():
    print("=== ユークリッドの互除法のデモンストレーション ===")
    
    # テストケース
    test_cases = [
        (48, 18),
        (100, 35),
        (1071, 462),
        (13, 17),  # 互いに素
        (144, 0),   # 特殊ケース
        (0, 25),    # 特殊ケース
    ]
    
    print("\n1. 基本的なGCD計算:")
    print(f"{'a':>10} {'b':>10} {'再帰版':>10} {'反復版':>10}")
    print("-" * 45)
    
    for a, b in test_cases:
        recursive_result = gcd_recursive(a, b) if b != 0 or a != 0 else max(a, b)
        iterative_result = gcd_iterative(a, b) if b != 0 or a != 0 else max(a, b)
        print(f"{a:>10} {b:>10} {recursive_result:>10} {iterative_result:>10}")
    
    # 詳細な計算過程
    print("\n2. 詳細な計算過程:")
    gcd_with_steps(48, 18)
    gcd_with_steps(1071, 462)
    
    # 最小公倍数の計算
    print("\n3. 最小公倍数（LCM）の計算:")
    for a, b in test_cases[:4]:  # 0を含まないケースのみ
        gcd_val = gcd_iterative(a, b)
        lcm_val = lcm(a, b)
        print(f"GCD({a}, {b}) = {gcd_val}, LCM({a}, {b}) = {lcm_val}")
    
    # 拡張ユークリッドの互除法
    print("\n4. 拡張ユークリッドの互除法:")
    print("ax + by = gcd(a, b) となる x, y を求める")
    for a, b in [(48, 18), (100, 35)]:
        gcd_val, x, y = extended_gcd(a, b)
        print(f"\n{a}x + {b}y = {gcd_val}")
        print(f"x = {x}, y = {y}")
        print(f"検証: {a} × {x} + {b} × {y} = {a*x + b*y}")

if __name__ == "__main__":
    main()
```

## 実行結果 - Execution Result
```
=== ユークリッドの互除法のデモンストレーション ===

1. 基本的なGCD計算:
         a          b      再帰版      反復版
---------------------------------------------
        48         18         6          6
       100         35         5          5
      1071        462        21         21
        13         17         1          1
       144          0       144        144
         0         25        25         25

2. 詳細な計算過程:

GCD(48, 18)の計算過程:
         a          b       余り
-----------------------------------
        48         18         12
        18         12          6
        12          6          0

GCD(48, 18) = 6

GCD(1071, 462)の計算過程:
         a          b       余り
-----------------------------------
      1071        462        147
       462        147         21
       147         21          0

GCD(1071, 462) = 21

3. 最小公倍数（LCM）の計算:
GCD(48, 18) = 6, LCM(48, 18) = 144
GCD(100, 35) = 5, LCM(100, 35) = 700
GCD(1071, 462) = 21, LCM(1071, 462) = 23562
GCD(13, 17) = 1, LCM(13, 17) = 221

4. 拡張ユークリッドの互除法:
ax + by = gcd(a, b) となる x, y を求める

48x + 18y = 6
x = -1, y = 3
検証: 48 × -1 + 18 × 3 = 6

100x + 35y = 5
x = 2, y = -5
検証: 100 × 2 + 35 × -5 = 5
```

## 解説 - Explanation
ユークリッドの互除法の原理：
- **基本原理**: gcd(a, b) = gcd(b, a mod b)
- **終了条件**: b = 0 のとき、gcd = a
- **時間複雑度**: O(log(min(a, b)))

アルゴリズムの特徴：
1. **効率性**: 対数時間で計算可能
2. **簡潔性**: 実装が非常にシンプル
3. **汎用性**: 拡張版では線形方程式の解も求まる

Grimoireでは、再帰は自己参照する円形パターンで表現され、反復は渦巻きパターンで表現されます。条件分岐は分岐線で示されます。

## 応用例 - Applications
- 分数の約分
- 暗号理論（RSA暗号など）
- 線形合同方程式の解法
- 多項式の計算
- コンピュータグラフィックスでの座標計算