package main

import (
	"fmt"
	"gogitlabber/cmd/gogitlabber/logging"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
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
	logging.Print("Initialize progressbar", nil)
	err := bar.RenderBlank()
	progressBarAdd(1)
	if err != nil {
		logging.Fatal("Initialization of the progressbar failed", err)
	}
}

func progressBarAdd(amount int) {
	logging.Print("BAR: Progressing the bar", nil)
	if err := bar.Add(amount); err != nil {
		logging.Print("BAR: Could not update the bar", err)
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
