package importer

import (
	"fmt"
	"go/build"
	"sort"
)

// Path is an immutable path from one Go Package.
// A path of [A B C] represents: "A imports package B and B imports package C".
type Path []string

// Append creates a new Path with the part appended to it.
// [A B].Append(C) returns [A B C].
func (p Path) Append(part string) Path {
	newPath := make(Path, len(p)+1)
	copy(newPath, p)
	newPath[len(p)] = part

	return newPath
}

// Package wraps a Go build.Package and adds additional data about how paths
// used to import vendored dependencies.
type Package struct {
	*build.Package

	ImportVendoredPaths map[string]string // Package dep import path to real vendored import path
}

// AllImports returns a list of all package import paths imported by the current package.
// If vendored is true, the import paths returned will be the actual paths the Go build tool
// resolves.
func (p *Package) AllImports(vendored bool) []string {
	importPaths := append(append(append([]string(nil), p.Imports...), p.TestImports...), p.XTestImports...)

	if vendored {
		for i, importPath := range importPaths {
			// Check existence in ImportVendoredPaths due to "C" package
			if _, ok := p.ImportVendoredPaths[importPath]; ok {
				importPaths[i] = p.ImportVendoredPaths[importPath]
			}
		}
	}

	return importPaths
}

// PackageGraph represents all Go packages reachable from a single package.
type PackageGraph struct {
	Packages map[string]*Package // Packages, keyed by the vendored import path
}

// ToMap returns a map of package names to the package names of all dependencies.
// The dependencies are vendored import paths.
func (p *PackageGraph) ToMap() map[string][]string {
	depMap := make(map[string][]string)
	for name, pkg := range p.Packages {
		deps := make([]string, 0, len(pkg.ImportVendoredPaths))
		for _, importPath := range pkg.ImportVendoredPaths {
			deps = append(deps, importPath)
		}
		sort.Strings(deps)
		depMap[name] = deps
	}

	return depMap
}

// GoPackages returns all build.Packages in the graph.
func (p *PackageGraph) GoPackages() []*build.Package {
	packages := make([]*build.Package, 0, len(p.Packages))

	for _, pkg := range p.Packages {
		packages = append(packages, pkg.Package)
	}

	return packages
}

// ShortestPath returns the shortest import path from one package to another.
// If there is no path between the packages then nil is returned.
// If there are multiple equally short paths, the path chosen to return is not deterministic.
// If either the from or to package does not exist, an error is returned.
func (p *PackageGraph) ShortestPath(from, to string) (Path, error) {
	if _, ok := p.Packages[from]; !ok {
		return nil, fmt.Errorf("Import path `%s` does not exist in graph", from)
	}
	if _, ok := p.Packages[to]; !ok {
		return nil, fmt.Errorf("Import path `%s` does not exist in graph", to)
	}

	visited := make(map[string]bool)
	pathQueue := []Path{Path([]string{from})}
	for len(pathQueue) > 0 {
		curPath := pathQueue[0]
		pathQueue = pathQueue[1:]

		lastImportPath := curPath[len(curPath)-1]
		if lastImportPath == to {
			return curPath, nil
		}
		if lastImportPath == "C" {
			continue
		}
		lastPkg, ok := p.Packages[lastImportPath]
		if !ok {
			return nil, fmt.Errorf("Unexpected: package `%s` does not exist in graph (path=%v)", lastImportPath, curPath)
		}

		for _, importName := range lastPkg.AllImports(true) {
			if !visited[importName] {
				visited[importName] = true
				pathQueue = append(pathQueue, curPath.Append(importName))
			}
		}
	}

	return nil, nil
}
