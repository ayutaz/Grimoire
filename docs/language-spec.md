# Grimoire Language Specification v1.0

## Table of Contents
1. [Introduction](#introduction)
2. [Basic Shapes and Their Meanings](#basic-shapes)
3. [Symbols and Operators](#symbols-operators)
4. [Program Structure](#program-structure)
5. [Data Types](#data-types)
6. [Control Flow](#control-flow)
7. [Functions and Scope](#functions-scope)
8. [Error Handling](#error-handling)
9. [Advanced Features](#advanced-features)

## Introduction {#introduction}

Grimoire is a visual programming language where programs are expressed as hand-drawn magic circles. Each geometric shape, symbol, and their spatial relationships carry semantic meaning.

### Core Principles
- **Spatial Programming**: Position and connections determine execution flow
- **Shape Semantics**: Each shape has inherent meaning
- **Symbolic Density**: Complex operations can be expressed in compact symbols
- **Artistic Freedom**: Multiple valid ways to draw the same program

## Basic Shapes and Their Meanings {#basic-shapes}

### Primary Shapes

| Shape | Symbol | Meaning | Usage |
|-------|--------|---------|--------|
| Circle | ○ | Scope/Function | Defines boundaries and function spaces |
| Square | □ | Data/Variable | Stores values and state |
| Triangle | △ | Decision/Branch | Conditional execution |
| Pentagon | ⬟ | Loop/Iteration | Repetitive execution |
| Hexagon | ⬢ | Parallel/Async | Concurrent execution |
| Star (5-point) | ☆ | Input/Output | I/O operations |
| Star (6-point) | ✡ | Network/External | External connections |

### Shape Modifiers

- **Filled vs Empty**: Filled shapes are constants, empty shapes are variables
- **Double Border**: Protected/Immutable
- **Dashed Border**: Optional/Nullable
- **Wavy Border**: Lazy evaluation

## Symbols and Operators {#symbols-operators}

### Arithmetic Operators
```
+ : Addition (cross)
- : Subtraction (horizontal line)
× : Multiplication (X shape)
÷ : Division (horizontal line with dots)
% : Modulo (percentage symbol)
^ : Power (upward arrow)
```

### Logical Operators
```
∧ : AND (upward wedge)
∨ : OR (downward wedge)
¬ : NOT (hook symbol)
⊕ : XOR (circled plus)
→ : IMPLIES (arrow)
```

### Comparison Operators
```
= : Equals (parallel lines)
≠ : Not equals (crossed equals)
< : Less than
> : Greater than
≤ : Less than or equal
≥ : Greater than or equal
```

### Special Symbols
```
∞ : Infinity/Unbounded loop
∅ : Null/Void
π : Pi constant
λ : Lambda/Anonymous function
Σ : Sum/Reduce
∫ : Integrate/Map
```

## Program Structure {#program-structure}

### Entry Point
Every Grimoire program must have a main circle marked with a double border or the symbol `◎`.

```
◎ ← Main entry point
|
○ ← Function definitions
|
□ ← Global variables
```

### Execution Flow
1. Programs execute from the outermost main circle
2. Flow follows connecting lines clockwise by default
3. Counterclockwise flow indicates callbacks or returns
4. Crossed lines indicate parallel execution paths

### Example: Basic Program Structure
```
        ◎ (main)
       / | \
      /  |  \
     □   △   ○
  (data)(if)(func)
```

## Data Types {#data-types}

### Primitive Types
Indicated by shape colors or patterns:
- **Integer**: Solid black outline
- **Float**: Gray/gradient outline
- **String**: Quoted text or wavy outline
- **Boolean**: Dot in center (•) for true, empty for false
- **Rune**: Single character in triangle

### Composite Types
- **Array**: Connected squares `□-□-□`
- **Map/Dict**: Hexagon with internal divisions
- **Set**: Circle with dots inside
- **Tuple**: Grouped shapes in parentheses

### Type Inference
The compiler infers types from:
1. Shape patterns
2. Connected operators
3. Initial values
4. Context of use

## Control Flow {#control-flow}

### Conditional Execution (Triangle)
```
    △
   / \
  /   \
 ○     ○
(true)(false)
```

### Loops (Pentagon)
```
⬟ (5) ← Loop 5 times
|
○ ← Loop body

⬟ (∞) ← Infinite loop
|
○
```

### Pattern Matching (Nested Triangles)
```
    △
   /|\
  / | \
 △  △  △
(patterns)
```

## Functions and Scope {#functions-scope}

### Function Definition
```
○ function_name(params)
├─ □ (local vars)
├─ ⬟ (logic)
└─ ☆ (return)
```

### Scope Rules
1. Inner circles can access outer circle variables
2. Overlapping circles share scope
3. Dotted circle borders indicate closure capture

### Anonymous Functions (Lambda)
```
λ○ ← Lambda circle
|
expression
```

## Error Handling {#error-handling}

### Try-Catch Pattern
```
○ (try)
├─ ~~~~ (wavy line = might fail)
└─ ○! (catch circle with !)
```

### Error Types
- **!**: Generic error
- **!!**: Critical error
- **?**: Warning
- **...**: Timeout

## Advanced Features {#advanced-features}

### Metaprogramming
Drawing shapes within quoted circles creates code-generating code:
```
"○" ← Quoted circle generates circles
```

### Concurrency Patterns
```
   ⬢ (spawn 6 threads)
  /|\
 / | \
○  ○  ○ (parallel tasks)
 \ | /
  \|/
   ⬢ (join)
```

### Memory Management
- **Solid arrow**: Strong reference
- **Dashed arrow**: Weak reference
- **Dotted arrow**: Lazy reference
- **⊗**: Explicit deallocation

### Optimization Hints
Line thickness indicates optimization priority:
- Thin: Cold path
- Normal: Regular path
- Thick: Hot path (optimize aggressively)

### Debug Annotations
- **👁**: Watchpoint
- **🐛**: Breakpoint
- **📍**: Assertion
- **💭**: Comment bubble

## Symbol Combination Rules

1. **Adjacency**: Adjacent symbols combine operations
2. **Nesting**: Inner symbols modify outer symbols
3. **Overlapping**: Shared properties
4. **Distance**: Farther symbols have weaker binding

## File Format

Grimoire files use the `.grim` extension and can be:
1. **PNG/JPG**: Scanned or digital drawings
2. **SVG**: Vector format for precise shapes
3. **GRIM**: Binary compiled format

## Compilation Directives

Special symbols in the corners of the image:
- **Top-left**: Version number
- **Top-right**: Optimization level
- **Bottom-left**: Target platform
- **Bottom-right**: Debug info

---

*This specification is subject to change as the language evolves.*