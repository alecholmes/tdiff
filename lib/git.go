package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// GitCommit describes a Git commit.
type GitCommit struct {
	SHA         string // Full SHA of the commit
	Description string // Commit description
}

// Git represents a Git local repository.
type Git struct {
	RootDir string // Root directory of the Git repository
}

// NewGit creates a new Git for the repository related to the current working directory.
// If the working directory is not in a Git repository then an error is returned.
func NewGit() (*Git, error) {
	return NewGitInDir("")
}

// NewGitInDir creates a new Git for the repository containing the given directory.
// If the given directory is not inside of a Git repository then an error is returned.
func NewGitInDir(dir string) (*Git, error) {
	rootDir, err := gitRootDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Unable to determine root of Git repository: %v", err)
	}

	return &Git{RootDir: rootDir}, nil
}

// DiffFiles returns list of files that were changed after fromSHA through toSHA. E.g. (fromSha, toSHA].
// The file names are relative to the root of the Go repository.
func (g *Git) DiffFiles(fromSHA, toSHA string) ([]string, error) {
	out, err := g.runGitCommand("diff", "--name-only", fmt.Sprintf("%s..%s", fromSHA, toSHA))
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

// Commits returns a list of commits after fromSHA through toSHA. E.g. (fromSha, toSHA].
// Commits are ordered from newest to older.
func (g *Git) Commits(fromSHA, toSHA string) ([]GitCommit, error) {
	// TODO: Is no-merges as default weird?
	out, err := g.runGitCommand("log", `--pretty=format:%H;%s`, "--no-merges", fmt.Sprintf("%s..%s", fromSHA, toSHA))
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

// CommitFiles returns the list of files changed in the commit of the given SHA.
// The file names are relative to the root of the Go repository.
func (g *Git) CommitFiles(sha string) ([]string, error) {
	out, err := g.runGitCommand("diff-tree", "--no-commit-id", "--name-only", "-r", sha)
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

func (g *Git) runGitCommand(args ...string) ([]byte, error) {
	args = append([]string{"-C", g.RootDir}, args...)
	return RunCommand("git", args...)
}

func gitRootDir(dir string) (string, error) {
	var args []string
	if len(dir) > 0 {
		args = []string{"-C", dir}
	}
	args = append(args, "rev-parse", "--show-toplevel")
	out, err := RunCommand("git", args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
