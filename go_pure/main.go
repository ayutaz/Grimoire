package main

import (
    "flag"
    "fmt"
    "image"
    "image/png"
    "log"
    "os"
    "time"
)

func main() {
    start := time.Now()
    
    // コマンドライン引数
    var (
        inputPath  = flag.String("i", "", "input image path")
        outputPath = flag.String("o", "", "output Python file path")
        debug      = flag.Bool("debug", false, "debug mode")
    )
    flag.Parse()
    
    if *inputPath == "" {
        if len(os.Args) > 1 {
            *inputPath = os.Args[1]
        } else {
            log.Fatal("Please provide an input image path")
        }
    }
    
    // 画像を読み込み
    file, err := os.Open(*inputPath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    img, _, err := image.Decode(file)
    if err != nil {
        log.Fatal(err)
    }
    
    // シンボルを検出
    detector := NewDetector()
    symbols := detector.Detect(img)
    
    if *debug {
        fmt.Printf("Detected %d symbols in %v\n", len(symbols), time.Since(start))
        for _, s := range symbols {
            fmt.Printf("  %+v\n", s)
        }
    }
    
    // パース
    parser := NewParser()
    ast := parser.Parse(symbols)
    
    // コンパイル
    compiler := NewCompiler()
    code := compiler.Compile(ast)
    
    // 出力
    if *outputPath != "" {
        err = os.WriteFile(*outputPath, []byte(code), 0644)
        if err != nil {
            log.Fatal(err)
        }
    } else {
        // 直接実行
        fmt.Print(code)
    }
    
    if *debug {
        fmt.Printf("Total execution time: %v\n", time.Since(start))
    }
}