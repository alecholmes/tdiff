package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func DiffFiles(fromSHA, toSHA string) ([]string, error) {
	out, err := RunCommand("git", "diff", "--name-only", fmt.Sprintf("%s..%s", fromSHA, toSHA))
	if err != nil {
		return nil, err
	}

	var files []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

type GitCommit struct {
	SHA         string
	Description string
}

// TODO: does this handle commit messages with newlines?
func Commits(fromSHA, toSHA string) ([]GitCommit, error) {
	// TODO: no-merges as default is weird

	out, err := RunCommand("git", "log", `--pretty=format:%H;%s`, "--no-merges", fmt.Sprintf("%s..%s", fromSHA, toSHA))
	if err != nil {
		return nil, err
	}

	var commits []GitCommit
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ";", 2)
		commits = append(commits, GitCommit{SHA: parts[0], Description: parts[1]})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commits, nil
}

func CommitFiles(sha string) ([]string, error) {
	out, err := RunCommand("git", "diff-tree", "--no-commit-id", "--name-only", "-r", sha)
	if err != nil {
		return nil, err
	}

	var files []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
