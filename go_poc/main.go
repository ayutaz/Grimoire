package main

import (
    "fmt"
    "image"
    "image/color"
    "log"
    "os"
    "time"
    
    "gocv.io/x/gocv"
    "github.com/spf13/cobra"
)

type SymbolType int

const (
    OuterCircle SymbolType = iota
    Circle
    Square
    Triangle
    Star
)

type Symbol struct {
    Type       SymbolType
    Position   image.Point
    Size       float64
    Confidence float64
}

type MagicCircleDetector struct {
    minContourArea float64
    circleThreshold float64
}

func NewMagicCircleDetector() *MagicCircleDetector {
    return &MagicCircleDetector{
        minContourArea: 100.0,
        circleThreshold: 0.8,
    }
}

func (d *MagicCircleDetector) DetectSymbols(imagePath string) ([]Symbol, error) {
    start := time.Now()
    
    // 画像を読み込み
    img := gocv.IMRead(imagePath, gocv.IMReadColor)
    if img.Empty() {
        return nil, fmt.Errorf("cannot read image: %s", imagePath)
    }
    defer img.Close()
    
    fmt.Printf("Image loaded in %v\n", time.Since(start))
    
    // グレースケール変換
    gray := gocv.NewMat()
    defer gray.Close()
    gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
    
    // 二値化
    binary := gocv.NewMat()
    defer binary.Close()
    gocv.Threshold(gray, &binary, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)
    
    // 輪郭検出
    contours := gocv.FindContours(binary, gocv.RetrievalExternal, gocv.ChainApproxSimple)
    
    symbols := []Symbol{}
    
    // 外円を検出
    if outerCircle := d.findOuterCircle(contours); outerCircle != nil {
        symbols = append(symbols, *outerCircle)
    }
    
    fmt.Printf("Total detection time: %v\n", time.Since(start))
    
    return symbols, nil
}

func (d *MagicCircleDetector) findOuterCircle(contours [][]image.Point) *Symbol {
    if len(contours) == 0 {
        return nil
    }
    
    // 最大の輪郭を見つける
    maxArea := 0.0
    maxIdx := -1
    
    for i, contour := range contours {
        area := gocv.ContourArea(contour)
        if area > maxArea {
            maxArea = area
            maxIdx = i
        }
    }
    
    if maxIdx == -1 || maxArea < d.minContourArea*10 {
        return nil
    }
    
    // 重心を計算
    moments := gocv.Moments(contours[maxIdx], false)
    centerX := moments["m10"] / moments["m00"]
    centerY := moments["m01"] / moments["m00"]
    
    return &Symbol{
        Type:       OuterCircle,
        Position:   image.Point{X: int(centerX), Y: int(centerY)},
        Size:       100.0, // 仮の値
        Confidence: 0.9,
    }
}

func runProgram(imagePath string) error {
    detector := NewMagicCircleDetector()
    symbols, err := detector.DetectSymbols(imagePath)
    if err != nil {
        return err
    }
    
    // シンプルなHello Worldの判定
    for _, symbol := range symbols {
        if symbol.Type == OuterCircle {
            fmt.Println("Hello, World!")
            break
        }
    }
    
    return nil
}

func compileProgram(imagePath string, outputPath string) error {
    detector := NewMagicCircleDetector()
    _, err := detector.DetectSymbols(imagePath)
    if err != nil {
        return err
    }
    
    pythonCode := "print('Hello, World!')"
    
    if outputPath != "" {
        return os.WriteFile(outputPath, []byte(pythonCode), 0644)
    }
    
    fmt.Println(pythonCode)
    return nil
}

func main() {
    var rootCmd = &cobra.Command{
        Use:   "grimoire",
        Short: "A visual programming language using magic circles",
    }
    
    var runCmd = &cobra.Command{
        Use:   "run [image]",
        Short: "Run a Grimoire program",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return runProgram(args[0])
        },
    }
    
    var output string
    var compileCmd = &cobra.Command{
        Use:   "compile [image]",
        Short: "Compile a Grimoire program to Python",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return compileProgram(args[0], output)
        },
    }
    compileCmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
    
    rootCmd.AddCommand(runCmd, compileCmd)
    
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}