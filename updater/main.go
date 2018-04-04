package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"syscall"

	"gopkg.in/src-d/go-git.v4"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Capture panics instead of closing the window at a superhuman speed before the user can read the message on Windows
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
			debug.PrintStack()
			pressAnyKey(scanner)
			return
		}
	}()

	updater(scanner)
}

func pressAnyKey(scanner *bufio.Scanner) {
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}

// The bool return is a little trick to condense two lines onto one
func logError(err error) bool {
	if err == nil {
		return true
	}
	fmt.Println(err)
	debug.PrintStack()
	return false
}

func updater(scanner *bufio.Scanner) bool {
	fmt.Println("Welcome to Gosora's Upgrader")
	fmt.Println("We're going to check for new updates, please wait patiently")

	repo, err := git.PlainOpen(".")
	if err != nil {
		return logError(err)
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return logError(err)
	}

	err = workTree.Pull(&git.PullOptions{Force: true})
	if err == git.NoErrAlreadyUpToDate {
		fmt.Println("You are already up-to-date")
		return true
	} else if err != nil && err != git.ErrUnstagedChanges { // fixes a bug in git where it refuses to update the files
		return logError(err)
	}

	// The unstaged files are particularly resistant, so blast them away at full force
	err = workTree.Reset(&git.ResetOptions{Mode: git.HardReset})
	if err != nil {
		return logError(err)
	}

	fmt.Println("Updated to the latest commit")
	headRef, err := repo.Head()
	if err != nil {
		return logError(err)
	}

	// Get information about the commit
	commit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return logError(err)
	}
	fmt.Println("Commit details:", commit)

	switch runtime.GOOS {
	case "windows":
		err = syscall.Exec("./patcher.bat", []string{}, os.Environ())
	default: //linux, etc.
		err = syscall.Exec("./patcher-linux", []string{}, os.Environ())
	}
	return logError(err)
}
