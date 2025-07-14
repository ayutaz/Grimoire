# Grimoireプログラム例集

## 基本的なプログラム

### 1. 算術演算
```
◎ main
├─ □ #a = 10
├─ □ #b = 20
├─ □ #sum = a + b
├─ □ #product = a × b
└─ ☆ "合計: " + sum + ", 積: " + product
```

### 2. 文字列操作
```
◎ main
├─ □ $firstName = "太郎"
├─ □ $lastName = "山田"
├─ □ $fullName = lastName + " " + firstName
└─ ☆ fullName
```

### 3. 配列操作
```
◎ main
├─ □ []numbers = [1, 2, 3, 4, 5]
├─ □ #sum = 0
├─ ⬟ (i in numbers)
│  └─ □ sum = sum + i
└─ ☆ "配列の合計: " + sum
```

## 制御構造

### 4. ネストした条件分岐
```
◎ main
├─ □ #score = 85
└─ △ (score >= 90)
   ├─true→ ☆ "優"
   └─false→ △ (score >= 80)
            ├─true→ ☆ "良"
            └─false→ △ (score >= 70)
                     ├─true→ ☆ "可"
                     └─false→ ☆ "不可"
```

### 5. While風ループ
```
◎ main
├─ □ #count = 0
└─ ⬟ (∞)
   ├─ △ (count >= 10)
   │  └─true→ ● break
   ├─ ☆ count
   └─ □ count = count + 1
```

## 関数とスコープ

### 6. 高階関数
```
○ map([]array, λfn)
├─ □ []result = []
├─ ⬟ (item in array)
│  └─ □ result.push(fn(item))
└─ ☆ result

◎ main
├─ □ []nums = [1, 2, 3, 4, 5]
├─ λ○ double = (x) → x * 2
├─ □ []doubled = map(nums, double)
└─ ☆ doubled  // [2, 4, 6, 8, 10]
```

### 7. クロージャ
```
○ makeCounter()
├─ □ #count = 0
└─ λ○ () → {
      □ count = count + 1
      ☆ count
   }

◎ main
├─ □ counter1 = makeCounter()
├─ □ counter2 = makeCounter()
├─ ☆ counter1()  // 1
├─ ☆ counter1()  // 2
├─ ☆ counter2()  // 1
└─ ☆ counter1()  // 3
```

## エラー処理

### 8. 複数のcatch
```
◎ main
└─ ○ (try)
   ├─ ~~~~ dangerousOperation()
   ├─ ○! (catch FileError e)
   │  └─ ☆ "ファイルエラー: " + e.message
   ├─ ○! (catch NetworkError e)
   │  └─ ☆ "ネットワークエラー: " + e.message
   └─ ○! (catch)
      └─ ☆ "不明なエラー"
```

## 並列処理の応用

### 9. パイプライン処理
```
◎ main
├─ □ []rawData = loadData()
│
├─ ⬢ stage1: filter
│  └─ λ○ (x) → x > 0
│     └─ []rawData
│        │
├─────────⬢ stage2: transform
│         └─ λ○ (x) → x * x
│            └─ []filtered
│               │
└─────────────⬢ stage3: aggregate
              └─ Σ sum
                 └─ []transformed
                    │
                    ☆ result
```

### 10. アクターモデル
```
○ actor({}mailbox)
└─ ⬟ (∞)
   └─ △ (mailbox.hasMessage())
      └─true→ process(mailbox.receive())

◎ main
├─ {}mailbox1 = createMailbox()
├─ {}mailbox2 = createMailbox()
│
└─ ⬢ actors
   ├─ ○ actor(mailbox1)
   ├─ ○ actor(mailbox2)
   └─ ○ coordinator(mailbox1, mailbox2)
```

## 実用的なプログラム

### 11. 簡易Webサーバー
```
○ handleRequest($request)
├─ △ (request.path == "/")
│  └─ ☆ "<h1>Welcome to Grimoire!</h1>"
├─ △ (request.path == "/api")
│  └─ ☆ {"status": "ok", "magic": true}
└─ ☆ "404 Not Found"

◎ main
└─ ✡ server.listen(8080, handleRequest)
   └─ ☆ "Server running on port 8080"
```

### 12. データベースアクセス
```
◎ main
├─ ✡ db = connect("mysql://localhost/grimoire")
│
├─ ○ (try)
│  ├─ □ $query = "SELECT * FROM spells WHERE level > ?"
│  ├─ □ []results = db.query(query, [5])
│  └─ ⬟ (spell in results)
│     └─ ☆ spell.name + " (Level " + spell.level + ")"
│
└─ ○! (catch)
   └─ ☆ "データベースエラー"
```

### 13. ファイル処理
```
○ processFile($filename)
├─ ☆ file = open(filename, "r")
├─ □ #lineCount = 0
├─ ⬟ (line in file.lines())
│  ├─ □ lineCount++
│  └─ △ (line.contains("魔法"))
│     └─ ☆ "Found magic at line " + lineCount
└─ file.close()

◎ main
└─ processFile("spellbook.txt")
```

## 視覚的な特殊パターン

### 14. 円形の依存関係
```
    ○ A ←─┐
    ↓     │
    ○ B   │
    ↓     │
    ○ C ──┘
```
循環参照を視覚的に表現

### 15. ツリー構造
```
         ○ root
        /│\
       / │ \
      ○  ○  ○
     /│\ │ /│\
    ○ ○ ○○○ ○ ○
```
階層的なデータ構造

## アルゴリズムとデータ構造

### 16. バブルソート
```
◎ main
├─ □ []array = [5, 2, 8, 1, 9]
├─ □ #n = array.length
├─ ⬟ (i = 0; i < n; i++)
│  └─ ⬟ (j = 0; j < n-i-1; j++)
│     └─ △ (array[j] > array[j+1])
│        └─true→ ○ swap
│                ├─ □ #temp = array[j]
│                ├─ □ array[j] = array[j+1]
│                └─ □ array[j+1] = temp
└─ ☆ "Sorted: " + array
```

### 17. 素数判定
```
○ isPrime(#n)
├─ △ (n <= 1) → false
├─ △ (n == 2) → true
├─ △ (n % 2 == 0) → false
├─ □ #sqrt_n = Math.sqrt(n)
└─ ⬟ (i = 3; i <= sqrt_n; i += 2)
   └─ △ (n % i == 0) → false
   └─ ☆ true

◎ main
├─ □ []testNumbers = [2, 3, 4, 17, 20, 29, 100, 101]
└─ ⬟ (num in testNumbers)
   └─ △ isPrime(num)
      ├─true→ ☆ num + " は素数"
      └─false→ ☆ num + " は素数ではない"
```

### 18. 文字列反転
```
○ reverseString($str)
├─ □ []chars = str.toArray()
├─ □ #left = 0
├─ □ #right = chars.length - 1
└─ ⬟ (left < right)
   ├─ □ #temp = chars[left]
   ├─ □ chars[left] = chars[right]
   ├─ □ chars[right] = temp
   ├─ □ left++
   └─ □ right--
└─ ☆ chars.join("")

◎ main
├─ □ $text = "Grimoire"
├─ □ $reversed = reverseString(text)
└─ ☆ text + " → " + reversed
```

### 19. スタック実装
```
○ Stack()
├─ □ []items = []
├─ ○ push(element)
│  └─ items.append(element)
├─ ○ pop()
│  ├─ △ isEmpty() → null
│  └─ items.pop()
├─ ○ peek()
│  ├─ △ isEmpty() → null
│  └─ items[-1]
├─ ○ isEmpty()
│  └─ items.length == 0
└─ ○ size()
   └─ items.length

◎ main
├─ □ stack = new Stack()
├─ stack.push(10)
├─ stack.push(20)
├─ stack.push(30)
├─ ☆ "Top: " + stack.peek()
├─ ☆ "Popped: " + stack.pop()
└─ ☆ "Size: " + stack.size()
```

### 20. ユークリッドの互除法（最大公約数）
```
○ gcd(#a, #b)
└─ ⬟ (b != 0)
   ├─ □ #remainder = a % b
   ├─ □ a = b
   └─ □ b = remainder
└─ ☆ a

○ lcm(#a, #b)
└─ ☆ (a * b) / gcd(a, b)

◎ main
├─ □ #x = 48
├─ □ #y = 18
├─ □ #g = gcd(x, y)
├─ □ #l = lcm(x, y)
├─ ☆ "GCD(" + x + "," + y + ") = " + g
└─ ☆ "LCM(" + x + "," + y + ") = " + l
```

---

これらの例は、Grimoireの表現力と視覚的な明確さを示しています。各サンプルの詳細な説明は、`examples/`ディレクトリ内の対応する`.grim.md`ファイルを参照してください。