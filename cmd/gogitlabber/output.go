package main

import (
	"fmt"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"github.com/scornet256/go-logger"
)

var bar *progressbar.ProgressBar

func progressBar() {

	// configure progressbar
	bar = progressbar.NewOptions(100,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription("..."),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionShowCount(),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	logger.Print("Initialize progressbar", nil)
}

func printSummary(stats *GitStats) {

	// print stats
	fmt.Println("")
	fmt.Printf(
		"Summary:\n"+
			" Cloned repositories: %v\n"+
			" Pulled repositories: %v\n"+
			" Errors: %v\n",
		stats.clonedCount,
		stats.pulledCount,
		stats.errorCount,
	)
}

func printPullErrorUnstaged(stats *GitStats) {
	if len(stats.pullErrorMsgUnstaged) > 0 {
		for _, repo := range stats.pullErrorMsgUnstaged {
			fmt.Printf("❕%s has unstaged changes.\n", repo)
		}
	}
}

func printPullErrorUncommitted(stats *GitStats) {
	if len(stats.pullErrorMsgUncommitted) > 0 {
		for _, repo := range stats.pullErrorMsgUncommitted {
			fmt.Printf("❕%s has uncommitted changes.\n", repo)
		}
	}
}
