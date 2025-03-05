package main

import (
	"errors"
	"testing"
)

// new type for the lookup function
type lookPathFunc func(file string) (string, error)

// modified function to accept a lookup function
func verifyGitAvailableWithLookup(lookPath lookPathFunc) error {
	_, err := lookPath("git")
	if err != nil {
		return errors.New("could not find git in path")
	}
	return nil
}

func TestVerifyGitAvailable(t *testing.T) {

	// test case 1: git is available
	err := verifyGitAvailableWithLookup(func(file string) (string, error) {
		return "/usr/bin/git", nil
	})
	if err != nil {
		t.Errorf("expected no error when git is available, got: %v", err)
	}

	// test case 2: git is not available
	err = verifyGitAvailableWithLookup(func(file string) (string, error) {
		return "", errors.New("git not found")
	})
	if err == nil {
		t.Error("expected error when git is not available, got nil")
	}
}
