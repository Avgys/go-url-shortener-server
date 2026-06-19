package main

import (
	"os/exec"
	"testing"
)

func TestGitCommands(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	if out, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output(); err != nil ||
		string(out) != "true\n" {
		t.Skip("not inside a git repository")
	}

	hash, err := getGitCommitHash()
	if err != nil {
		t.Fatalf("getGitCommitHash: %v", err)
	}
	if hash == "" {
		t.Fatal("getGitCommitHash returned empty string")
	}

	name, err := getGitCommitName()
	if err != nil {
		t.Fatalf("getGitCommitName: %v", err)
	}
	if name == "" {
		t.Fatal("getGitCommitName returned empty string")
	}

	branch, err := getGitBranch()
	if err != nil {
		t.Fatalf("getGitBranch: %v", err)
	}
	if branch == "" {
		t.Fatal("getGitBranch returned empty string")
	}

	date, err := getGitLastCommitDate()
	if err != nil {
		t.Fatalf("getGitLastCommitDate: %v", err)
	}
	if date == "" {
		t.Fatal("getGitLastCommitDate returned empty string")
	}
}

func TestFillBuildInfoFromGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	BuildCommitHash = "N/A"
	BuildCommitName = "N/A"
	BuildBranch = "N/A"
	BuildDate = "N/A"

	fillBuildInfoFromGit()

	if BuildCommitHash == "N/A" {
		t.Fatal("BuildCommitHash was not filled")
	}
	if BuildCommitName == "N/A" {
		t.Fatal("BuildCommitName was not filled")
	}
	if BuildBranch == "N/A" {
		t.Fatal("BuildBranch was not filled")
	}
	if BuildDate == "N/A" {
		t.Fatal("BuildDate was not filled")
	}
}
