package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

// add a mutex to safely increment shared counters
var mu sync.Mutex

func checkoutRepositories(repositories []Repository, concurrency int) {

	// create a waitgroup + semaphore channel
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	// manage all repositories found
	for _, repo := range repositories {

		// increment waitgroup counter + acquire semaphore slot
		wg.Add(1)
		semaphore <- struct{}{}

		// start go routine per repo
		go func(repo Repository) {

			// ensure we release the semaphore and close the goroutine
			defer func() {
				<-semaphore
				wg.Done()
			}()

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
					log.Printf("ERROR: %v\n", err)
				}

				// set a lock, increment counters and unlock
				mu.Lock()
				clonedCount++
				mu.Unlock()
				progressBarAdd(1)

			// pull the latest
			case strings.Contains(string(repoStatus), url):
				pullRepository(repoName, repoDestination)
				progressBarAdd(1)

			default:
				log.Printf("ERROR: decided not to clone or pull repository %v\n", repoName)
				log.Printf("ERROR: this is why: %v\n", repoStatus)

				// set a lock, increment counters and unlock
				mu.Lock()
				errorCount++
				mu.Unlock()
				progressBarAdd(1)
			}
		}(repo)
	}

	// wait for goroutines
	wg.Wait()
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

	// set a lock, increment counters and unlock
	mu.Lock()
	pulledCount++
	mu.Unlock()

	if err != nil {

		// set a lock, increment counters and unlock
		mu.Lock()
		errorCount++
		pulledCount--
		mu.Unlock()

		switch {
		case strings.Contains(string(pullOutput), "You have unstaged changes"):
			pullErrorMsg = append(pullErrorMsg, repoDestination)

		default:
			log.Printf("ERROR: pulling %v\n", err)
		}
	}
}
