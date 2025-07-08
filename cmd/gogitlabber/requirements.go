package main

import (
	"fmt"
	"os/exec"
)

// verify if git is available
func verifyGitAvailable() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed or not in PATH: %w", err)
	}
	return nil
}
