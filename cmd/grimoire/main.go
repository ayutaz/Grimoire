package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ayutaz/grimoire/internal/cli"
	"github.com/ayutaz/grimoire/internal/i18n"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	start := time.Now()

	// CLIの実行
	if err := cli.Execute(version, commit, date); err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("error.error_prefix"), err)
		return 1
	}

	// デバッグモードの場合は実行時間を表示
	if os.Getenv("GRIMOIRE_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, i18n.T("error.execution_time"), time.Since(start))
	}

	return 0
}
