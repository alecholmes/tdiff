package app

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecholmes/tdiff/importer"
	"github.com/alecholmes/tdiff/lib"
)

type Logger func(string, ...interface{})

func NoLogging(string, ...interface{}) {}

type Package struct {
	ImportPath   string   `json:"name"`
	PathFromRoot []string `json:"pathFromRoot"`
}

type Commit struct {
	SHA              string     `json:"sha"`
	Description      string     `json:"description"`
	RelevantPackages []*Package `json:"relevantPackages"` // This may be empty if only artifacts in non-Go subdirectories changed.
}

type Summary struct {
	RootImportPath string     `json:"rootImportPath"`
	SHA            string     `json:"sha"`
	Packages       []*Package `json:"packages"`
	Commits        []*Commit  `json:"commits"`
	Files          []string   `json:"files"`
}

type Differ struct {
	goPath         string
	importer       func(string) (*importer.PackageGraph, error)
	includeCommits bool
	includePaths   bool
	logger         Logger
}

func NewDiffer(goPath string, importer func(string) (*importer.PackageGraph, error), includeCommits, includePaths bool, logger Logger) *Differ {
	return &Differ{
		goPath:         goPath,
		importer:       importer,
		includeCommits: includeCommits,
		includePaths:   includePaths,
		logger:         logger,
	}
}

func (d *Differ) Diff(importPath, sha string, artifacts bool) (*Summary, error) {
	diff := diff{
		summary: Summary{
			RootImportPath: importPath,
			SHA:            sha,
		},
	}

	if err := diff.determineRelevantPackages(d.goPath, artifacts, d.logger); err != nil {
		return nil, err
	}

	if err := diff.createPackageSummaries(d.includePaths); err != nil {
		return nil, err
	}

	diff.determineRelevantFiles()

	if err := diff.determineCommits(); err != nil {
		return nil, err
	}

	return &diff.summary, nil
}

type diff struct {
	summary Summary

	git                  *lib.Git
	graph                *importer.PackageGraph
	relevantPackages     lib.StringSet       // Relevant packages that changed
	packageSummaries     map[string]*Package // Summaries by package import path
	changedArtifactFiles []string            // Artifacts that changed
	changedPackageFiles  map[string][]string // Files that changed by package
	changedFilePackage   map[string]string   // Package name by file
}

func (d *diff) determineRelevantPackages(goPath string, artifacts bool, logger Logger) error {
	packageNamer, git, err := newGitPackageNamer(d.summary.RootImportPath, goPath, logger)
	d.git = git
	if err != nil {
		return err
	}

	// Find all packages recursively reachable from the given root package.
	reachablePackages, packageGraph, err := recursiveDeps(d.summary.RootImportPath)
	if err != nil {
		return err
	}
	d.graph = packageGraph

	// Add given root package to the reachable set
	reachablePackages = append(reachablePackages, d.summary.RootImportPath)

	reachablePackageSet := make(lib.StringSet)
	reachablePackageSet.Add(reachablePackages...)

	// Find all files that changed since the given SHA.
	// Not all files will be relevant, as some will be in unreachable packages.
	files, err := git.DiffFiles(d.summary.SHA, "HEAD")
	if err != nil {
		return err
	}

	// Determine all the packages with changes.
	d.changedPackageFiles = make(map[string][]string)
	d.changedFilePackage = make(map[string]string)
	for _, file := range files {
		// For non-Go source files, some derived package names might not be actual Go packages.
		packageName := packageNamer(filepath.Dir(file))
		if reachablePackageSet.Contains(packageName) {
			d.changedPackageFiles[packageName] = append(d.changedPackageFiles[packageName], file)
			d.changedFilePackage[file] = packageName
		}
	}

	// Determine relevant packages, where relevant is a package reachable from
	// the root package that has also changed.
	d.relevantPackages = make(lib.StringSet)
	for _, pkg := range reachablePackages {
		if _, ok := d.changedPackageFiles[pkg]; ok {
			d.relevantPackages.Add(pkg)
		}
	}

	// Add any artifact files that changed
	if artifacts {
		for _, file := range files {
			if !strings.HasSuffix(file, ".go") {
				packageName := packageNamer(filepath.Dir(file))

				// Not the most efficient way of doing this...
				for _, reachablePackage := range reachablePackages {
					if importPathNestedWithin(packageName, reachablePackage) {
						d.changedArtifactFiles = append(d.changedArtifactFiles, file)

						// Artifact files don't necessarily live in a real Go package
						if reachablePackageSet.Contains(packageName) {
							d.relevantPackages.Add(packageName)
						}
					}
				}
			}
		}
	}

	return nil
}

func (d *diff) createPackageSummaries(includePaths bool) error {
	d.packageSummaries = make(map[string]*Package)
	outPackages := d.relevantPackages.Slice()
	sort.Strings(outPackages)

	for _, pkg := range outPackages {
		packageSummary := &Package{ImportPath: pkg}
		d.summary.Packages = append(d.summary.Packages, packageSummary)
		d.packageSummaries[pkg] = packageSummary

		if includePaths {
			shortestPath, err := d.graph.ShortestPath(d.summary.RootImportPath, pkg)
			if err != nil {
				return err
			}
			if len(shortestPath) == 0 {
				return fmt.Errorf("Expected path between %s and %s", d.summary.RootImportPath, pkg)
			}
			packageSummary.PathFromRoot = shortestPath
		}
	}

	return nil
}

func (d *diff) determineRelevantFiles() {
	outFileSet := make(lib.StringSet)
	outFileSet.Add(d.changedArtifactFiles...)
	for pkg := range d.relevantPackages {
		outFileSet.Add(d.changedPackageFiles[pkg]...)
	}

	outFiles := outFileSet.Slice()
	sort.Strings(outFiles)
	d.summary.Files = outFiles
}

func (d *diff) determineCommits() error {
	commits, err := d.git.Commits(d.summary.SHA, "HEAD")
	if err != nil {
		log.Fatal(err)
	}

	// Determine all files that changed in the relevant subset of changed packages.
	relevantFiles := make(lib.StringSet)
	for pkg := range d.relevantPackages {
		relevantFiles.Add(d.changedPackageFiles[pkg]...)
	}
	relevantFiles.Add(d.changedArtifactFiles...)

	var relevantCommits []lib.GitCommit
	for _, commit := range commits {
		commitFiles, err := d.git.CommitFiles(commit.SHA)
		if err != nil {
			return err
		}

		// The commit should be included if any files in it were part of a relevant changed package.
		relevant := false
		for _, file := range commitFiles {
			if relevantFiles.Contains(file) {
				relevant = true
				break
			}
		}
		if relevant {
			relevantCommits = append(relevantCommits, commit)

			commitPackageSet := make(lib.StringSet)
			for _, file := range commitFiles {
				commitPackageSet.Add(d.changedFilePackage[file])
			}

			commitPackages := commitPackageSet.Slice()
			sort.Strings(commitPackages)
			commitPackageSummaries := []*Package{}
			for _, commitPackage := range commitPackages {
				if summary, ok := d.packageSummaries[commitPackage]; ok {
					commitPackageSummaries = append(commitPackageSummaries, summary)
				}
			}

			d.summary.Commits = append(d.summary.Commits, &Commit{
				SHA:              commit.SHA,
				Description:      commit.Description,
				RelevantPackages: commitPackageSummaries,
			})

		}
	}

	return nil
}

// GoPackagerNamer determines a full package name given a partial package name.
type goPackagerNamer func(partialPackageName string) string

// NewGitPackageNamer creates a goPackagerNamer and Git based on the
// given package name and Go path.
// This function will attempt to use the Go path to determine the root of
// the Git repository the given package lives under.
func newGitPackageNamer(importPath, goPath string, logger Logger) (goPackagerNamer, *lib.Git, error) {
	srcDir := fmt.Sprintf("%s/src/", goPath)
	logger("Using source directory: %s", srcDir)

	git, err := lib.NewGitInDir(fmt.Sprintf("%s%s", srcDir, importPath))
	if err != nil {
		return nil, nil, err
	}

	logger("Using git root: %s", git.RootDir)

	if !strings.HasPrefix(git.RootDir, srcDir) {
		return nil, nil, fmt.Errorf("Expected git root to be under %s; working directory is %s", srcDir, git.RootDir)
	}

	packagePrefix := git.RootDir[len(srcDir):]
	logger("Prefixing packages with: %s", packagePrefix)

	return func(relativePackage string) string {
		if len(relativePackage) == 0 || relativePackage == "." {
			return packagePrefix
		}

		return fmt.Sprintf("%s/%s", packagePrefix, relativePackage)
	}, git, nil
}

// importPathNestedWithin returns true if the given import path is or starts with maybeOuterImportPath.
// importPathNestedWithin("a/b", "a/b") == true
// importPathNestedWithin("a/b/c", "a/b") == true
// importPathNestedWithin("a/x", "a/b") == true
// importPathNestedWithin("a", "a/b") == false
func importPathNestedWithin(importPath, maybeOuterImportPath string) bool {
	if !strings.HasPrefix(importPath, maybeOuterImportPath) {
		return false
	} else if len(importPath) == len(maybeOuterImportPath) {
		return true
	}

	return importPath[len(maybeOuterImportPath)] == '/'
}

func recursiveDeps(packageName string) ([]string, *importer.PackageGraph, error) {
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
