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

func printSummary() {

	fmt.Println("")
	fmt.Printf(
		"Summary:\n"+
			" Cloned repositories: %v\n"+
			" Pulled repositories: %v\n"+
			" Errors: %v\n",
		clonedCount,
		pulledCount,
		errorCount,
	)
}

func printPullErrorUnstaged(pullErrorMsgUnstaged []string) {
	if len(pullErrorMsgUnstaged) > 0 {
		for _, repo := range pullErrorMsgUnstaged {
			fmt.Printf("❕%s has unstaged changes.\n", repo)
		}
	}
}

func printPullErrorUncommitted(pullErrorMsgUncommitted []string) {
	if len(pullErrorMsgUncommitted) > 0 {
		for _, repo := range pullErrorMsgUncommitted {
			fmt.Printf("❕%s has uncommitted changes.\n", repo)
		}
	}
}
