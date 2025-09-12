package main

import (
	"fmt"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"github.com/scornet256/go-logger"
)

var bar *progressbar.ProgressBar

// make progressbar
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

// update progressbar
func updateProgressBar(repoCount int) error {
	if config.Debug {
		return nil // Skip progress bar in debug mode
	}
	logger.Print("Resetting progress bar", nil)
	if err := bar.Set(0); err != nil {
		return fmt.Errorf("resetting progress bar: %w", err)
	}
	logger.Print("Setting progress bar maximum", nil)
	bar.ChangeMax(repoCount)
	return nil
}

// print summary
func printSummary(stats *GitStats) {
	// print stats
	fmt.Println("")
	fmt.Printf(
		"Summary:\n"+
			" Cloned repositories: %v\n"+
			" Pulled repositories: %v\n"+
			" Errors: %v\n\n",
		stats.clonedCount,
		stats.pulledCount,
		stats.errorCount,
	)
}

// print pull errors unstaged
func printPullErrorUnstaged(stats *GitStats) {
	if len(stats.pullErrorMsgUnstaged) > 0 {
		fmt.Println("Repositories with unstaged changes:")
		for _, repo := range stats.pullErrorMsgUnstaged {
			fmt.Printf("❕ %s has unstaged changes.\n", repo)
		}
		fmt.Println()
	}
}

// print pull errors uncommited
func printPullErrorUncommitted(stats *GitStats) {
	if len(stats.pullErrorMsgUncommitted) > 0 {
		fmt.Println("Repositories with uncommitted changes:")
		for _, repo := range stats.pullErrorMsgUncommitted {
			fmt.Printf("❕ %s has uncommitted changes.\n", repo)
		}
		fmt.Println()
	}
}

// print all errors
func printAllErrors(stats *GitStats) {
	printPullErrorUnstaged(stats)
	printPullErrorUncommitted(stats)
	printGeneralErrors(stats)
}

// print general errors
func printGeneralErrors(stats *GitStats) {
	if len(stats.generalErrors) > 0 {
		fmt.Println("Repositories with errors:")
		for _, repo := range stats.generalErrors {
			fmt.Printf("❌ %s failed to process.\n", repo)
		}
		fmt.Println()
	}
}

// check for errors
func hasErrors(stats *GitStats) bool {
	return len(stats.pullErrorMsgUnstaged) > 0 ||
		len(stats.pullErrorMsgUncommitted) > 0 ||
		len(stats.generalErrors) > 0
}

// print detailed summary
func printDetailedSummary(stats *GitStats) {
	printSummary(stats)

	if hasErrors(stats) {
		fmt.Println("Error Details:")
		printAllErrors(stats)
	}
}
