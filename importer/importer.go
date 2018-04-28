package importer

import (
	"fmt"
	"go/build"
	"os"
	"strings"
)

// DefaultRecursiveImport imports a Go package and all of its reachable dependencies.
// The environment's GOPATH and default build context are used.
func DefaultRecursiveImport(importPath string) (*PackageGraph, error) {
	buildCtx := build.Default
	return RecursiveImport(importPath, os.Getenv("GOPATH"), &buildCtx)
}

// RecursiveImport imports a Go package and all of its reachable dependencies.
func RecursiveImport(importPath string, goPath string, buildContext *build.Context) (*PackageGraph, error) {
	importer := newRecursiveImporter(goPath, buildContext)
	if err := importer.importPackage(nil, importPath); err != nil {
		return nil, err
	}

	return &PackageGraph{
		Packages: importer.packages,
	}, nil
}

type recursiveImporter struct {
	goPath       string
	buildContext *build.Context
	packages     map[string]*Package
}

func newRecursiveImporter(goPath string, buildContext *build.Context) *recursiveImporter {
	return &recursiveImporter{
		goPath:       goPath,
		buildContext: buildContext,
		packages:     make(map[string]*Package),
	}
}

// importPackage imports the package identified by importPath.
// If parentPkg is given, it uses the parent package's import path to determine
// possible vendored packages to import for importPath.
func (r *recursiveImporter) importPackage(parentPkg *Package, importPath string) error {
	pkg, err := r.vendoredImport(parentPkg, importPath)
	if err != nil {
		return err
	} else if pkg == nil {
		return nil
	}

	if err := r.importAll(pkg, pkg.Imports); err != nil {
		return err
	}

	if err := r.importAll(pkg, pkg.TestImports); err != nil {
		return err
	}

	return r.importAll(pkg, pkg.XTestImports)
}

func (r *recursiveImporter) importAll(parentPkg *Package, importPaths []string) error {
	for _, importPath := range importPaths {
		if err := r.importPackage(parentPkg, importPath); err != nil {
			return fmt.Errorf("%s > %v", packageImportPath(parentPkg), err)
		}
	}

	return nil
}

// vendoredImport attempts to import a possibly vendored package relative to a parent package.
// If the package does not exist, nil is returned.
func (r *recursiveImporter) vendoredImport(parentPkg *Package, importPath string) (*Package, error) {
	// Ignore cgo import
	if importPath == "C" {
		return nil, nil
	}

	possibleImportPaths := vendorPaths(importPath, packageImportPath(parentPkg))

	var goPkg *build.Package
	var lastErr error
	for _, possibleImportPath := range possibleImportPaths {
		if _, ok := r.packages[possibleImportPath]; ok {
			parentPkg.ImportVendoredPaths[importPath] = possibleImportPath
			return nil, nil
		}

		goPkg, lastErr = r.rawImport(possibleImportPath)
		if lastErr == nil {
			pkg := &Package{
				Package:             goPkg,
				ImportVendoredPaths: make(map[string]string),
			}
			if parentPkg != nil {
				parentPkg.ImportVendoredPaths[importPath] = possibleImportPath
			}
			r.packages[possibleImportPath] = pkg
			return pkg, nil
		}
	}
	return nil, lastErr
}

// rawImport attempts to import a build.Package based on an import path.
// If the package does not exist, nil is returned.
func (r *recursiveImporter) rawImport(importPath string) (*build.Package, error) {
	// https: //github.com/golang/go/issues/17326
	lookupName := importPath
	if strings.HasPrefix(importPath, "golang_org/") {
		lookupName = fmt.Sprintf("vendor/%s", importPath)
	}

	pkg, err := r.buildContext.Import(lookupName, r.goPath, 0)
	if err != nil {
		return nil, err
	}

	if len(pkg.Name) == 0 {
		return nil, nil
	}

	return pkg, nil
}

func packageImportPath(pkg *Package) string {
	if pkg == nil {
		return ""
	}

	return pkg.ImportPath
}

// vendorPaths returns all possible import paths for a package.
// importPath is the import path of the package to import, and
// fromImportPath is the parent package that is importing the package identified by importPath.
//
// For example, if package "a/b" (fromImportPath="a/b") wants to import packageName="target", it
// may try the following ordered vendored import paths:
// [a/b/vendor/target, a/vendor/target, vendor/target, target]
func vendorPaths(importPath, fromImportPath string) []string {
	var parts []string
	if len(fromImportPath) > 0 {
		parts = strings.Split(fromImportPath, "/")
	}

	paths := make([]string, 0, len(parts)+2)
	for i := len(parts); i >= 0; i-- {
		paths = append(paths, strings.Join(append(append(parts[0:i], "vendor"), importPath), "/"))
	}

	return append(paths, importPath)
}
