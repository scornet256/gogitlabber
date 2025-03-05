package main

import (
	"fmt"
	"log"

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
	logPrint("Initialize progressbar", nil)
	err := bar.RenderBlank()
	progressBarAdd(1)
	if err != nil {
		logFatal("Initialization of the progressbar failed", err)
	}
}

func progressBarAdd(amount int) {
	if err := bar.Add(amount); err != nil {
		logPrint("ERROR: Progress bar update error: %v\n", err)
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

func logPrint(message string, err error) {
	if debug {
		if err != nil {
			log.Printf("gogitlabber | DEBUG: %v error: %v\n", message, err)
		}
		if err == nil {
			log.Printf("gogitlabber | DEBUG: %v\n", message)
		}
	}
}

func logFatal(message string, err error) {
	if err != nil {
		log.Fatalf("gogitlabber | FATAL: %v error: %v\n", message, err)
	}
	log.Fatalf("gogitlabber | FATAL: %v\n", message)
}
