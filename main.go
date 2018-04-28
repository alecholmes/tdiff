package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecholmes/tdiff/importer"
	"github.com/alecholmes/tdiff/lib"
)

type Package struct {
	Name         string   `json:"name"`
	PathFromRoot []string `json:"pathFromRoot,omitempty"`
}

type Commit struct {
	SHA              string     `json:"sha"`
	Description      string     `json:"description"`
	RelevantPackages []*Package `json:"relevantPackages,omitempty"`
}

type Summary struct {
	Packages []*Package `json:"packages,omitempty"`
	Commits  []*Commit  `json:"commits,omitempty"`
	Files    []string   `json:"files,omitempty"`
}

var (
	packageFlag   = flag.String("package", "", "Package to find reachable diff from")
	shaFlag       = flag.String("sha", "", "Git SHA")
	artifactsFlag = flag.Bool("artifacts", false, "If true, includes changed non-Go source files under the package directory, recursive")

	verboseFlag = flag.Bool("verbose", false, "If set, log verbose debugging information")

	// Output format flags
	packagesFlag = flag.Bool("packages", false, "If set, all relevant changed packages printed")
	filesFlag    = flag.Bool("files", false, "If set, all relevant changed files are printed")
	commitsFlag  = flag.Bool("commits", false, "If set, all relevant commits are printed")
	jsonFlag     = flag.Bool("json", false, "If set, JSON object representing all changes is printed")
)

func main() {
	flag.Parse()
	if len(*packageFlag) == 0 || len(*shaFlag) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// TODO: Check that exactly one of the output flags is set.
	packageNamer, git, err := newGitPackageNamer(*packageFlag, *verboseFlag)
	if err != nil {
		log.Fatal(err)
	}

	// Find all files that changed since the given SHA.
	// Not all files will be relevant, as some will be in unreachable packages.
	files, err := git.DiffFiles(*shaFlag, "HEAD")
	if err != nil {
		log.Fatal(err)
	}

	// Determine all the packages with changes.
	changedPackageFiles := make(map[string][]string)
	changedFilePackage := make(map[string]string)
	var changedArtifactFiles []string
	for _, file := range files {
		packageName := packageNamer(filepath.Dir(file))
		if strings.HasSuffix(file, ".go") {
			changedPackageFiles[packageName] = append(changedPackageFiles[packageName], file)
			changedFilePackage[file] = packageName
		} else if *artifactsFlag && selfOrChildPackage(packageName, *packageFlag) {
			changedArtifactFiles = append(changedArtifactFiles, file)
		}
	}

	// Find all packages recursively reachable from the given root package.
	//reachablePackages, err := recursiveDeps(packageFlag)
	reachablePackages, packageGraph, err := newRecursiveDeps(*packageFlag)
	if err != nil {
		log.Fatal(err)
	}

	// Add given root package to the reachable set
	reachablePackages = append(reachablePackages, *packageFlag)

	// Determine relevant packages, where relevant is a package reachable from
	// the root package that has also changed.
	relevantPackages := make(stringSet)
	for _, pkg := range reachablePackages {
		if _, ok := changedPackageFiles[pkg]; ok {
			relevantPackages.add(pkg)
		}
	}

	summary := new(Summary)
	packageSummaries := make(map[string]*Package)

	if *packagesFlag || *jsonFlag {
		outPackages := relevantPackages.slice()
		sort.Strings(outPackages)

		for _, pkg := range outPackages {
			shortestPath, err := packageGraph.ShortestPath(*packageFlag, pkg)
			if err != nil {
				log.Fatal(err)
			}
			if len(shortestPath) == 0 {
				log.Fatalf("Expected path between %s and %s", *packageFlag, pkg)
			}

			if *jsonFlag {
				packageSummary := &Package{
					Name:         pkg,
					PathFromRoot: shortestPath,
				}
				summary.Packages = append(summary.Packages, packageSummary)
				packageSummaries[pkg] = packageSummary
			}

			if *packagesFlag {
				fmt.Println(pkg)
			}
		}
	}

	if *filesFlag || *jsonFlag {
		outFiles := changedArtifactFiles
		for pkg := range relevantPackages {
			outFiles = append(outFiles, changedPackageFiles[pkg]...)
		}

		sort.Strings(outFiles)
		for _, file := range outFiles {
			if *jsonFlag {
				summary.Files = append(summary.Files, file)
			}

			if *filesFlag {
				fmt.Println(file)
			}
		}
	}

	if *commitsFlag || *jsonFlag {
		commits, err := git.Commits(*shaFlag, "HEAD")
		if err != nil {
			log.Fatal(err)
		}

		// Determine all files that changed in the relevant subset of changed packages.
		relevantFiles := make(stringSet)
		for pkg := range relevantPackages {
			relevantFiles.add(changedPackageFiles[pkg]...)
		}
		relevantFiles.add(changedArtifactFiles...)

		var relevantCommits []lib.GitCommit
		for _, commit := range commits {
			commitFiles, err := git.CommitFiles(commit.SHA)
			if err != nil {
				log.Fatal(err)
			}

			// The commit should be included if any files in it were part of a relevant changed package.
			relevant := false
			for _, file := range commitFiles {
				if relevantFiles.contains(file) {
					relevant = true
					break
				}
			}
			if relevant {
				relevantCommits = append(relevantCommits, commit)

				if *jsonFlag {
					commitPackageSet := make(stringSet)
					for _, file := range commitFiles {
						commitPackageSet.add(changedFilePackage[file])
					}

					commitPackages := commitPackageSet.slice()
					sort.Strings(commitPackages)
					var commitPackageSummaries []*Package
					for _, commitPackage := range commitPackages {
						if summary, ok := packageSummaries[commitPackage]; ok {
							commitPackageSummaries = append(commitPackageSummaries, summary)
						}
					}

					summary.Commits = append(summary.Commits, &Commit{
						SHA:              commit.SHA,
						Description:      commit.Description,
						RelevantPackages: commitPackageSummaries,
					})
				}
			}
		}

		if *commitsFlag {
			for _, commit := range relevantCommits {
				fmt.Printf("%s %s\n", commit.SHA, commit.Description)
			}
		}
	}

	if *jsonFlag {
		body, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(body))
	}
}

// stringSet represents an unordered set of strings.
type stringSet map[string]bool

// add puts the given values into the set.
func (s stringSet) add(values ...string) {
	for _, v := range values {
		s[v] = true
	}
}

// contains returns true iff the set contains the given value.
// s.contains("foo") and s["foo"] are equivalent.
func (s stringSet) contains(value string) bool {
	return s[value]
}

// slice returns an arbitrarily ordered slice that contains each value in the set.
func (s stringSet) slice() []string {
	values := make([]string, 0, len(s))
	for v := range s {
		values = append(values, v)
	}

	return values
}

// goPackagerNamer determines a full package name given a partial package name.
type goPackagerNamer func(partialPackageName string) string

// newGitPackageNamer creates a goPackagerNamer and Git based on the current Go environment
// and a given package name.
// This function will attempt to use the GOPATH environment variable to figure out the root of
// the Git repository the given package lives under.
func newGitPackageNamer(packageName string, verbose bool) (goPackagerNamer, *lib.Git, error) {
	srcDir := fmt.Sprintf("%s/src/", os.Getenv("GOPATH"))
	if verbose {
		log.Printf("Using source directory: %s", srcDir)
	}

	git, err := lib.NewGitInDir(fmt.Sprintf("%s%s", srcDir, packageName))
	if err != nil {
		return nil, nil, err
	}

	if verbose {
		log.Printf("Using git root: %s", git.RootDir)
	}

	if !strings.HasPrefix(git.RootDir, srcDir) {
		return nil, nil, fmt.Errorf("Expected git root to be under %s; working directory is %s", srcDir, git.RootDir)
	}

	packagePrefix := git.RootDir[len(srcDir):]
	if verbose {
		log.Printf("Prefixing packages with: %s", packagePrefix)
	}

	return func(relativePackage string) string {
		return fmt.Sprintf("%s/%s", packagePrefix, relativePackage)
	}, git, nil
}

// Returns true if the given packageName is either the same package or a child
// of selfOrParentPackageName.
// selfOrParentPackageName("a/b", "a/b") == true
// selfOrParentPackageName("a/b/c", "a/b") == true
// selfOrParentPackageName("a", "a/b") == false
func selfOrChildPackage(packageName, selfOrParentPackageName string) bool {
	if !strings.HasPrefix(packageName, selfOrParentPackageName) {
		return false
	} else if len(packageName) == len(selfOrParentPackageName) {
		return true
	}

	return packageName[len(selfOrParentPackageName)] == '/'
}

func newRecursiveDeps(packageName string) ([]string, *importer.PackageGraph, error) {
	graph, err := importer.DefaultRecursiveImport(packageName)
	if err != nil {
		return nil, nil, err
	}

	packageNames := make([]string, 0, len(graph.Packages))
	for name := range graph.Packages {
		packageNames = append(packageNames, name)
	}

	return packageNames, graph, nil
}

func recursiveDeps(root *string) ([]string, error) {
	args := []string{"list", "-f", "'{{.Deps}}'"}
	if root != nil {
		args = append(args, *root)
	}
	out, err := lib.RunCommand("go", args...)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}

	str := string(out)
	// This assumes that `go list` outputs as Go-formatted array of strings.
	// E.g. '[]', '[foo]', '[foo bar]'
	str = strings.Trim(str, "'[]\n")

	return strings.Split(str, " "), nil
}

// goDirs take a list of filenames, filters out ones that are not Go source files,
// and returns the set of parent directories of these files.
func goDirs(files []string) stringSet {
	dirSet := make(stringSet)
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			dirSet.add(filepath.Dir(file))
		}
	}

	return dirSet
}
