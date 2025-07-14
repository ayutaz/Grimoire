# スタックの実装 - Stack Implementation

## 概要 - Overview
基本的なデータ構造であるスタック（LIFO: Last In First Out）の実装例です。
This example implements a basic stack data structure (LIFO: Last In First Out).

## 魔法陣の構造 - Magic Circle Structure

### 外周円 - Outer Circle
スタッククラス全体を囲む大きな円

### メインエントリ - Main Entry (二重円)
デモンストレーションプログラムの開始点

### クラス定義 - Class Definition
- スタッククラスを表す特別な円形領域
- 内部にメソッドを表す小さな円が配置

### メソッド - Methods
1. `__init__`: コンストラクタ（初期化）
2. `push`: 要素の追加
3. `pop`: 要素の取り出し
4. `peek`: 最上位要素の参照
5. `is_empty`: 空判定
6. `size`: サイズ取得

### 変数定義 - Variable Definition
- `items`: スタックの内部配列
- `element`: 操作対象の要素

## 生成されるPythonコード - Generated Python Code

```python
class Stack:
    """スタック（LIFO）データ構造の実装"""
    
    def __init__(self):
        """スタックの初期化"""
        self.items = []
    
    def push(self, element):
        """要素をスタックの頂上に追加"""
        self.items.append(element)
        print(f"Pushed: {element}")
    
    def pop(self):
        """スタックの頂上から要素を取り出す"""
        if self.is_empty():
            print("Error: Stack is empty!")
            return None
        element = self.items.pop()
        print(f"Popped: {element}")
        return element
    
    def peek(self):
        """スタックの頂上の要素を参照（取り出さない）"""
        if self.is_empty():
            print("Stack is empty!")
            return None
        return self.items[-1]
    
    def is_empty(self):
        """スタックが空かどうかを判定"""
        return len(self.items) == 0
    
    def size(self):
        """スタックのサイズを取得"""
        return len(self.items)
    
    def display(self):
        """スタックの内容を表示"""
        if self.is_empty():
            print("Stack: [empty]")
        else:
            print(f"Stack: {self.items} <- top")

def demonstrate_stack():
    """スタックの動作をデモンストレーション"""
    print("=== スタックのデモンストレーション ===\n")
    
    # スタックの作成
    stack = Stack()
    
    # 初期状態
    print("初期状態:")
    stack.display()
    print(f"Is empty? {stack.is_empty()}")
    print()
    
    # 要素の追加
    print("要素を追加:")
    stack.push(10)
    stack.push(20)
    stack.push(30)
    stack.display()
    print(f"Size: {stack.size()}")
    print()
    
    # Peek操作
    print("Peek操作:")
    top = stack.peek()
    print(f"Top element: {top}")
    stack.display()
    print()
    
    # Pop操作
    print("Pop操作:")
    stack.pop()
    stack.display()
    stack.pop()
    stack.display()
    print()
    
    # さらに要素を追加
    print("さらに要素を追加:")
    stack.push(40)
    stack.push(50)
    stack.display()
    print()
    
    # 全要素を取り出す
    print("全要素を取り出す:")
    while not stack.is_empty():
        stack.pop()
    stack.display()
    
    # 空のスタックからpop（エラーケース）
    print("\n空のスタックからpop:")
    stack.pop()

def main():
    demonstrate_stack()
    
    # 実用的な例：括弧の対応チェック
    print("\n\n=== 括弧の対応チェックの例 ===")
    
    def check_parentheses(expression):
        stack = Stack()
        pairs = {'(': ')', '[': ']', '{': '}'}
        
        for char in expression:
            if char in pairs:  # 開き括弧
                stack.push(char)
            elif char in pairs.values():  # 閉じ括弧
                if stack.is_empty():
                    return False
                opening = stack.pop()
                if pairs[opening] != char:
                    return False
        
        return stack.is_empty()
    
    # テストケース
    test_cases = [
        "(a + b) * [c - d]",
        "{x * (y + z)}",
        "((a + b)",
        "a + b) * c",
        "{[()]}",
        "{[(])}"
    ]
    
    for expr in test_cases:
        result = check_parentheses(expr)
        status = "OK" if result else "NG"
        print(f"'{expr}' -> {status}")

if __name__ == "__main__":
    main()
```

## 実行結果 - Execution Result
```
=== スタックのデモンストレーション ===

初期状態:
Stack: [empty]
Is empty? True

要素を追加:
Pushed: 10
Pushed: 20
Pushed: 30
Stack: [10, 20, 30] <- top
Size: 3

Peek操作:
Top element: 30
Stack: [10, 20, 30] <- top

Pop操作:
Popped: 30
Stack: [10, 20] <- top
Popped: 20
Stack: [10] <- top

さらに要素を追加:
Pushed: 40
Pushed: 50
Stack: [10, 40, 50] <- top

全要素を取り出す:
Popped: 50
Popped: 40
Popped: 10
Stack: [empty]

空のスタックからpop:
Error: Stack is empty!


=== 括弧の対応チェックの例 ===
'(a + b) * [c - d]' -> OK
'{x * (y + z)}' -> OK
'((a + b)' -> NG
'a + b) * c' -> NG
'{[()]}' -> OK
'{[([)]' -> NG
```

## 解説 - Explanation
スタックの特徴：
- **LIFO**: 最後に入れた要素が最初に出る
- **時間複雑度**: push/pop/peek はすべてO(1)
- **空間複雑度**: O(n)（n は要素数）

Grimoireでは、クラスは特別な境界を持つ魔法陣として表現され、メソッドは内部の小円として配置されます。データフローは矢印で示され、条件分岐は分岐線で表現されます。

## 応用例 - Applications
- 式の評価（後置記法）
- 関数呼び出しスタック
- Undo/Redo機能の実装
- 深さ優先探索（DFS）
- バックトラッキングアルゴリズム