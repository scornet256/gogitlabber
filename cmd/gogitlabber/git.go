package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func cloneRepository(repoDestination string, gitlabUrl string) (string, error) {

	cloneCmd := exec.Command("git", "clone", gitlabUrl, repoDestination)
	cloneOutput, err := cloneCmd.CombinedOutput()

	return string(cloneOutput), err
}

func findRemote(repoDestination string) (string, error) {

	remoteCmd := exec.Command("git", "-C", repoDestination, "remote", "show")
	remoteOutput, err := remoteCmd.CombinedOutput()
	remote := strings.Split(strings.TrimSpace(string(remoteOutput)), "\n")[0]

	if err != nil {
		log.Printf("\n❌ error finding remote for: %s\n", err)
	}

	return remote, err
}

func pullRepositories(repoDestination string) (string, error) {

	remote, err := findRemote(repoDestination)
	pullCmd := exec.Command("git", "-C", repoDestination, "pull", remote)
	pullOutput, err := pullCmd.CombinedOutput()

	if err != nil {
		errorCount = errorCount + 1
		if strings.Contains(string(pullOutput), "You have unstaged changes") {
			pullErrorMsg = append(pullErrorMsg, repoDestination)
		}
	}

	bar.Add(1)
	return string(pullOutput), err
}

func printPullError(pullErrorMsg []string) {
	if len(pullErrorMsg) > 0 {
		for _, repo := range pullErrorMsg {
			fmt.Printf("❕%s has unstaged changes.\n", repo)
		}
	}
}
