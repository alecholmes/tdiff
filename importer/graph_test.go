package importer

import (
	"fmt"
	"go/build"
	"reflect"
	"testing"
)

func TestShortestPast(t *testing.T) {
	addFakePackage := func(graph *PackageGraph, name string, imports ...string) *Package {
		pkg := &Package{
			Package: &build.Package{
				ImportPath: name,
				Imports:    imports,
			},
			ImportVendoredPaths: make(map[string]string),
		}

		for _, importPath := range imports {
			pkg.ImportVendoredPaths[importPath] = importPath
		}

		graph.Packages[name] = pkg

		return pkg
	}

	// A -> [B, G]
	// B -> [G]
	// G -> [D]
	// D -> []
	// X -> [B, Y]
	// Y -> []
	graph := &PackageGraph{Packages: make(map[string]*Package)}
	addFakePackage(graph, "A", "B", "G")
	addFakePackage(graph, "B", "G")
	addFakePackage(graph, "G", "D")
	addFakePackage(graph, "D")
	addFakePackage(graph, "X", "Y", "B")
	addFakePackage(graph, "Y")

	testCases := []struct {
		from string
		to   string
		path []string
	}{
		{from: "A", to: "A", path: []string{"A"}},
		{from: "A", to: "B", path: []string{"A", "B"}},
		{from: "A", to: "G", path: []string{"A", "G"}},
		{from: "A", to: "D", path: []string{"A", "G", "D"}},
		{from: "X", to: "A", path: nil},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("from=%s to=%s path=%v", tc.from, tc.to, tc.path), func(t *testing.T) {
			path, err := graph.ShortestPath(tc.from, tc.to)
			if err != nil {
				t.Fatal(err)
			}
			if expectedPath := Path(tc.path); !reflect.DeepEqual(expectedPath, path) {
				t.Fatalf("Expected path %v but got %v", tc.path, path)
			}
		})
	}

	if _, err := graph.ShortestPath("A", "does not exist"); err == nil {
		t.Errorf("Expected error but got none")
	}
	if _, err := graph.ShortestPath("does not exist", "A"); err == nil {
		t.Errorf("Expected error but got none")
	}
}
