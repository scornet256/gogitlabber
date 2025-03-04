package main

import (
	"log"
	"os/exec"
)

func verifyGitAvailable() {
	_, err := exec.LookPath("git")
	if err != nil {
		log.Fatal("could not find git in path")
	}
}
