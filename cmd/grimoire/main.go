package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ayutaz/grimoire/internal/cli"
	grimoireErrors "github.com/ayutaz/grimoire/internal/errors"
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
		// Format error output based on error type
		if _, ok := err.(*grimoireErrors.EnhancedError); ok {
			// Enhanced error already has proper formatting
			fmt.Fprintln(os.Stderr, err)
		} else {
			// Legacy error format
			fmt.Fprintf(os.Stderr, i18n.T("error.error_prefix"), err)
		}
		return 1
	}

	// デバッグモードの場合は実行時間を表示
	if os.Getenv("GRIMOIRE_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, i18n.T("error.execution_time"), time.Since(start))
	}

	return 0
}
