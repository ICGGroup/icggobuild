package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func getRepoRootPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	path := ""

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return path, err
	}

	scanner := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		return path, err
	}

	for scanner.Scan() {
		path = string(scanner.Bytes())
	}

	if err := cmd.Wait(); err != nil {
		return path, err
	}

	return path, nil

}

func getCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	path := ""

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return path, err
	}

	scanner := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		return path, err
	}

	for scanner.Scan() {
		path = string(scanner.Bytes())
	}

	if err := cmd.Wait(); err != nil {
		return path, err
	}

	return path, nil

}

type gitChange struct {
	status string
	path   string
}

func getChanges() ([]gitChange, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	changes := []gitChange{}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return changes, err
	}

	scanner := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		return changes, err
	}

	for scanner.Scan() {
		output := scanner.Bytes()
		status := string(output[:2])
		file := string(output[3:])

		changes = append(changes, gitChange{status: status, path: file})

	}

	if err := cmd.Wait(); err != nil {
		return changes, err
	}

	return changes, nil

}

func goBuild(args ...string) error {
	// Get the path of the current directory
	hash, err := getCommitHash()
	if err != nil {
		return err
	}

	fmtDt := time.Now().UTC().Format(time.RFC3339)

	goArgs := []string{"build", "-ldflags", fmt.Sprintf("-X main.buildDate=%s -X main.commitHash=%s", fmtDt, hash)}

	goArgs = append(goArgs, args...)
	cmd := exec.Command("go", goArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil

}

func main() {
	argsWithoutProg := os.Args[1:]

	// Get the path of the current directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Get the path of the repository
	repoPath, err := getRepoRootPath()
	if err != nil {
		log.Fatal(err)
	}

	// Get the modified files
	changes, err := getChanges()
	if err != nil {
		log.Fatal(err)
	}

	// Loop over the files extracting those that are in the current directory or below
	commitChanges := []gitChange{}
	for _, c := range changes {
		af := path.Join(repoPath, c.path)
		if strings.HasPrefix(af, cwd) {
			commitChanges = append(commitChanges, c)
		}
	}

	if len(commitChanges) > 0 {
		fmt.Println("The following changes must be committed prior to build:\n")
		for _, c := range commitChanges {
			fmt.Printf("%s %s\n", c.status, c.path)
		}
		fmt.Println("")
		os.Exit(2)
		return
	}

	// Do the build
	err = goBuild(argsWithoutProg...)
	if err != nil {
		os.Exit(2)
	}

}
