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

// set progressbar maximum to the number of repositories
func updateProgressBar(repoCount int) error {
	if bar == nil {
		return nil
	}
	if err := bar.Set(0); err != nil {
		return fmt.Errorf("resetting progress bar: %w", err)
	}
	bar.ChangeMax(repoCount)
	return nil
}

// print a titled list, skipped when empty
func printList(title, format string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Println(title)
	for _, item := range items {
		fmt.Printf(format, item)
	}
	fmt.Println()
}

// print detailed summary
func printDetailedSummary(stats *GitStats) {
	fmt.Printf(
		"\nSummary:\n"+
			" Cloned repositories: %d\n"+
			" Pulled repositories: %d\n"+
			" Errors: %d\n\n",
		stats.cloned,
		stats.pulled,
		stats.errorCount(),
	)

	if stats.errorCount() == 0 {
		return
	}

	fmt.Println("Error Details:")
	printList("Repositories with unstaged changes:", "• %s has unstaged changes.\n", stats.unstaged)
	printList("Repositories with uncommitted changes:", "• %s has uncommitted changes.\n", stats.uncommitted)
	printList("Repositories with errors:", "✗ %s\n", stats.failed)
}
