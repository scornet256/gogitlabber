package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/scornet256/go-logger"
)

// add a mutex to safely increment shared counters
var mu sync.Mutex

func checkoutRepositories(repositories []Repository) {

	// create a waitgroup + semaphore channel
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.Concurrency)

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
			repoDestination := config.Destination + repoName

			// log activity
			logger.Print("Starting on repository: "+repoName, nil)

			// make git url
			url := fmt.Sprintf("https://%s-token:%s@%s/%s.git", config.GitBackend, config.GitToken, config.GitHost, repoName)

			// check current status of repoDestination
			checkRepo := func(repoDestination string) string {
				checkCmd := exec.Command("git", "-C", repoDestination, "remote", "-v")
				checkOutput, _ := checkCmd.CombinedOutput()
				logger.Print("Checking status for repository: "+repoName, nil)

				return string(checkOutput)
			}
			repoStatus := checkRepo(repoDestination)

			// report error if not cloned or pulled repository
			// clone repository if it does not exist
			switch {
			case strings.Contains(string(repoStatus), "No such file or directory"):

				// log activity
				logger.Print("Decided to clone repository: "+repoName, nil)

				// clone the repo
				cloneRepository := func(repoDestination string, url string) (string, error) {
					cloneCmd := exec.Command("git", "clone", url, repoDestination)
					cloneOutput, err := cloneCmd.CombinedOutput()

					// set username and email
					setGitUserName(repoName, repoDestination)
					setGitUserMail(repoName, repoDestination)

					logger.Print("Cloning repository: "+repoName+" to "+repoDestination, nil)

					return string(cloneOutput), err
				}
				_, err := cloneRepository(repoDestination, url)
				if err != nil {
					logger.Print("ERROR: %v\n", err)
				}

				// set a lock, increment counters, update progressbar  and unlock
				mu.Lock()
				clonedCount++
				if !config.Debug {
					_ = bar.Add(1)
				}
				mu.Unlock()

			// pull the latest
			case strings.Contains(string(repoStatus), url):
				logger.Print("Decided to pull repository: "+repoName, nil)
				pullRepository(repoName, repoDestination)

				// set username and email
				setGitUserName(repoName, repoDestination)
				setGitUserMail(repoName, repoDestination)

				if !config.Debug {
					_ = bar.Add(1)
				}

			default:
				logger.Print("ERROR: decided not to clone or pull repository: "+repoName, nil)
				logger.Print("ERROR: this is why: "+repoStatus, nil)

				// set a lock, increment counters and unlock
				mu.Lock()
				errorCount++
				if !config.Debug {
					_ = bar.Add(1)
				}
				mu.Unlock()
			}
		}(repo)
	}

	// wait for goroutines
	wg.Wait()
}

func pullRepository(repoName string, repoDestination string) {

	// log activity
	logger.Print("Pulling repository: "+repoName+" at "+repoDestination, nil)

	// find remote
	findRemote := func(repoDestination string) (string, error) {
		remoteCmd := exec.Command("git", "-C", repoDestination, "remote", "show")
		remoteOutput, err := remoteCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("finding remote: %v", err)
		}

		logger.Print("Finding remote for repository: "+repoName+" at "+repoDestination, nil)
		remote := strings.Split(strings.TrimSpace(string(remoteOutput)), "\n")[0]
		logger.Print("Found remote; "+remote+" for repository: "+repoName+" at "+repoDestination, nil)
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
			pullErrorMsgUnstaged = append(pullErrorMsgUnstaged, repoDestination)
			logger.Print("Found unstaged changes in repository: "+repoName+" at "+repoDestination, nil)

		case strings.Contains(string(pullOutput), "Your index contains uncommitted changes"):
			pullErrorMsgUncommitted = append(pullErrorMsgUncommitted, repoDestination)
			logger.Print("Found uncommitted changes in repository: "+repoName+" at "+repoDestination, nil)

		default:
			logger.Print("ERROR: pulling "+repoName, nil)
		}
	}

	// log activity
	logger.Print("Pulled repository: "+repoName+" at "+repoDestination, nil)
}

// function to set the git user name
func setGitUserName(repoName string, repoDestination string) {

	gitUserNameCmd := exec.Command("git", "-C", repoDestination, "config", "user.name", config.GitUserName)
	gitUserNameCmd.CombinedOutput()

	logger.Print("Setting git username for: "+repoName, nil)
}

// function to set the git user mail
func setGitUserMail(repoName string, repoDestination string) {

	gitUserMailCmd := exec.Command("git", "-C", repoDestination, "config", "user.mail", config.GitUserMail)
	gitUserMailCmd.CombinedOutput()

	logger.Print("Setting git email for: "+repoName, nil)
}
