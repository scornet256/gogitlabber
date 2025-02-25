package main

import (
  "log"
	"os"
	"os/exec"
)

func verifyGitAvailable() {
	_, err := exec.LookPath("git")
	if err != nil {
    log.Fatalf("Error: could not find git in path")
		os.Exit(1)
	}
}
