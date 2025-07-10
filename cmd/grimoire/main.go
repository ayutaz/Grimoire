package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ayutaz/grimoire/internal/cli"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	start := time.Now()

	// CLIの実行
	if err := cli.Execute(version, commit, date); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// デバッグモードの場合は実行時間を表示
	if os.Getenv("GRIMOIRE_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "Execution time: %v\n", time.Since(start))
	}
}