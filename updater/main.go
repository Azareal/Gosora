package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime/debug"

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

	err := updater(scanner)
	if err != nil {
		fmt.Println(err)
	}
}

func pressAnyKey(scanner *bufio.Scanner) {
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}

func updater(scanner *bufio.Scanner) error {
	fmt.Println("Welcome to Gosora's Upgrader")
	fmt.Print("We're going to check for new updates, please wait patiently")

	repo, err := git.PlainOpen("./.git")
	if err != nil {
		return err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = workTree.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		return err
	}

	fmt.Println("Updated to the latest commit")
	headRef, err := repo.Head()
	if err != nil {
		return err
	}

	fmt.Println("Commit details:")
	commit, err := repo.CommitObject(headRef.Hash())
	return err
}
