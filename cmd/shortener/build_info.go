package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func getGitCommitHash() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func getGitCommitName() (string, error) {
	out, err := exec.Command("git", "log", "-1", "--format=%s").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func getGitBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func getGitLastCommitDate() (string, error) {
	out, err := exec.Command("git", "log", "-1", "--format=%cI").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func fillBuildInfoFromGit() {
	if BuildCommitHash == "N/A" {
		if v, err := getGitCommitHash(); err == nil {
			BuildCommitHash = v
		}
	}

	if BuildCommitName == "N/A" {
		if v, err := getGitCommitName(); err == nil {
			BuildCommitName = v
		}
	}

	if BuildBranch == "N/A" {
		if v, err := getGitBranch(); err == nil {
			BuildBranch = v
		}
	}

	if BuildDate == "N/A" {
		if v, err := getGitLastCommitDate(); err == nil {
			BuildDate = v
		}
	}
}

func printBuildInfo() {
	fillBuildInfoFromGit()

	for _, field := range []struct {
		label string
		value string
	}{
		{"Build version", BuildVersion},
		{"Build date", BuildDate},
		{"Build commit hash", BuildCommitHash},
		{"Build commit name", BuildCommitName},
		{"Build branch", BuildBranch},
	} {
		fmt.Printf("%s: %s\n", field.label, field.value)
	}
}
