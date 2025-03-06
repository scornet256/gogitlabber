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
	bar = progressbar.NewOptions(2,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription("Logging into Gitlab..."),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionShowCount(),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// initialize progressbar
	logger.Print("Initialize progressbar", nil)
	err := bar.RenderBlank()
	progressBarAdd(1)
	if err != nil {
		logger.Fatal("Initialization of the progressbar failed", err)
	}
}

func progressBarAdd(amount int) {
	logger.Print("BAR: Progressing the bar", nil)
	if err := bar.Add(amount); err != nil {
		logger.Print("BAR: Could not update the bar", err)
	}
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

func printPullError(pullErrorMsg []string) {
	if len(pullErrorMsg) > 0 {
		for _, repo := range pullErrorMsg {
			fmt.Printf("❕%s has unstaged changes.\n", repo)
		}
	}
}
