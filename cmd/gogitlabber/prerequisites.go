package main

import (
	"os/exec"
)

func verifyGitAvailable() error {
	_, err := exec.LookPath("git")
	if err != nil {
    return err
	}
  return nil
}
