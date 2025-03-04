package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func checkoutRepositories(repositories []Repository) {
	for _, repo := range repositories {

		// get repository name + create repo destination
		repoName := string(repo.PathWithNamespace)
		repoDestination := repoDestinationPre + repoName

		// make gitlab url
		url := fmt.Sprintf("https://gitlab-token:%s@%s/%s.git", gitlabToken, gitlabHost, repoName)

		// check current status of repoDestination
		checkRepo := func(repoDestination string) string {
			checkCmd := exec.Command("git", "-C", repoDestination, "remote", "-v")
			checkOutput, _ := checkCmd.CombinedOutput()

			return string(checkOutput)
		}
		repoStatus := checkRepo(repoDestination)

		// report error if not cloned or pulled repository
		// clone repository if it does not exist
		switch {
		case strings.Contains(string(repoStatus), "No such file or directory"):

			// update the progress bar
			descriptionPrefixPre := "Cloning repository "
			descriptionPrefix := descriptionPrefixPre + repoName + " ..."
			bar.Describe(descriptionPrefix)

			// clone the repo
			cloneRepository := func(repoDestination string, url string) (string, error) {
				cloneCmd := exec.Command("git", "clone", url, repoDestination)
				cloneOutput, err := cloneCmd.CombinedOutput()

				return string(cloneOutput), err
			}
			_, err := cloneRepository(repoDestination, url)
			if err != nil {
				log.Printf("error: %v\n", err)
			}
			clonedCount = clonedCount + 1
			progressBarAdd(1)

		// pull the latest
		case strings.Contains(string(repoStatus), url):
			pullRepository(repoName, repoDestination)
			progressBarAdd(1)

		default:
			log.Printf("error: decided not to clone or pull repository %v\n", repoName)
			log.Printf("error: this is why: %v\n", repoStatus)
			errorCount = errorCount + 1
			progressBarAdd(1)
		}
	}
}

func pullRepository(repoName string, repoDestination string) {

	// update the progress bar
	descriptionPrefixPre := "Pulling repository "
	descriptionPrefix := descriptionPrefixPre + repoName + " ..."
	bar.Describe(descriptionPrefix)

	// find remote
	findRemote := func(repoDestination string) (string, error) {
		remoteCmd := exec.Command("git", "-C", repoDestination, "remote", "show")
		remoteOutput, err := remoteCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("finding remote: %v\n", err)
		}

		remote := strings.Split(strings.TrimSpace(string(remoteOutput)), "\n")[0]
		return remote, nil
	}
	remote, _ := findRemote(repoDestination)

	// pull repository
	pullCmd := exec.Command("git", "-C", repoDestination, "pull", remote)
	pullOutput, err := pullCmd.CombinedOutput()
	pulledCount = pulledCount + 1

	if err != nil {
		errorCount = errorCount + 1
		pulledCount = pulledCount - 1

		switch {
		case strings.Contains(string(pullOutput), "You have unstaged changes"):
			pullErrorMsg = append(pullErrorMsg, repoDestination)

		default:
			log.Printf("error: pulling %v\n", err)
		}
	}
}
