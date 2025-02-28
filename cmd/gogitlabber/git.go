package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func checkoutRepositories(repositories []Repository) {
	for _, repo := range repositories {

		// create clone
		repoName := string(repo.PathWithNamespace)
		url := getGitlabURL(gitlabToken, gitlabHost, repoName)

		// create repository destination
		repoDestination := repoDestinationPre + repoName

		// create and update bar description
		descriptionPrefixPre := "Cloning repository "
		descriptionPrefix := descriptionPrefixPre + repoName + " ..."
		bar.Describe(descriptionPrefix)

		// clone the repo
    cloneRepository := func(repoDestination string, gitlabUrl string) (string, error) {
      cloneCmd := exec.Command("git", "clone", gitlabUrl, repoDestination)
      cloneOutput, err := cloneCmd.CombinedOutput()

      return string(cloneOutput), err
    }
		cloneOutput, err := cloneRepository(repoDestination, url)

    // try to pull if clone didnt work 
		if err != nil {

			// if repo already exists, try to pull the latest changes
			if strings.Contains(string(cloneOutput),
				"already exists and is not an empty directory") {

				descriptionPrefixPre := "Pulling repository "
				descriptionPrefix := descriptionPrefixPre + repoName + " ..."
				bar.Describe(descriptionPrefix)

				pullRepository(repoDestination)
				pulledCount = pulledCount + 1
				continue
			}

			// in case cloning failed and the directory does not exist
			// print the clone error and continue
			log.Printf("failed to clone %s: %v", repoName, err)
			errorCount = errorCount + 1
			bar.Add(1)
			continue
		}

		// finish the clone
		clonedCount = clonedCount + 1
		bar.Add(1)
	}
}

func pullRepository(repoDestination string) (string, error) {

  // find remote
  findRemote := func(repoDestination string) (string, error) {
    remoteCmd := exec.Command("git", "-C", repoDestination, "remote", "show")
    remoteOutput, err := remoteCmd.CombinedOutput()
    if err != nil {
      return "", fmt.Errorf("finding remote: %v", err)
    }

    remote := strings.Split(strings.TrimSpace(string(remoteOutput)), "\n")[0]
    return remote, nil
  }
	remote, _ := findRemote(repoDestination)

  // pull repository
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
