package main

import (
	"fmt"
	"log"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

var bar *progressbar.ProgressBar

func progressBar(repositories []Repository) {
	repoCount := len(repositories)

	// make progressbar
	barPrefix := "Getting your one and only repository..."
	if repoCount > 1 {
		barPrefix = "Getting your repositories..."
	}

	bar = progressbar.NewOptions(
		repoCount,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(barPrefix),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

func progressBarAdd(amount int) {
	if err := bar.Add(amount); err != nil {
    log.Printf("ERROR: Progress bar update error: %v\n", err)
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
