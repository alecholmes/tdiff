package importer

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRecursiveImport(t *testing.T) {
	graph, err := DefaultRecursiveImport("github.com/alecholmes/tdiff/importer/test_packages/a")
	if err != nil {
		t.Fatal(err)
	}

	expectedDepMap := map[string][]string{
		"github.com/alecholmes/tdiff/importer/test_packages/a": []string{
			"github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa",
			"github.com/alecholmes/tdiff/importer/test_packages/b",
			"github.com/alecholmes/tdiff/importer/test_packages/vendor/p",
		},
		"github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa": []string{
			"github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa/vendor/p",
			"github.com/alecholmes/tdiff/importer/test_packages/b",
		},
		"github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa/vendor/p": []string{},
		"github.com/alecholmes/tdiff/importer/test_packages/b": []string{
			"unsafe",
		},
		"github.com/alecholmes/tdiff/importer/test_packages/vendor/p": []string{},
		"unsafe": []string{},
	}
	if actual := graph.ToMap(); !reflect.DeepEqual(expectedDepMap, actual) {
		t.Fatalf("Expected dep graph `%v` but got `%v`", expectedDepMap, actual)
	}

	expectedAPaths := map[string]string{
		"github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa": "github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa",
		"github.com/alecholmes/tdiff/importer/test_packages/b":        "github.com/alecholmes/tdiff/importer/test_packages/b",
		"p": "github.com/alecholmes/tdiff/importer/test_packages/vendor/p",
	}
	if actual := graph.Packages["github.com/alecholmes/tdiff/importer/test_packages/a"].ImportVendoredPaths; !reflect.DeepEqual(actual, expectedAPaths) {
		t.Errorf("Expected ImportVendoredPaths %v but got %v", expectedAPaths, actual)
	}

	expectedAAAPaths := map[string]string{
		"p": "github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa/vendor/p",
		"github.com/alecholmes/tdiff/importer/test_packages/b": "github.com/alecholmes/tdiff/importer/test_packages/b",
	}
	if actual := graph.Packages["github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa"].ImportVendoredPaths; !reflect.DeepEqual(actual, expectedAAAPaths) {
		t.Errorf("Expected ImportVendoredPaths %v but got %v", expectedAAAPaths, actual)
	}
}

func TestVendorPaths(t *testing.T) {
	testCases := []struct {
		packageName      string
		importPath       string
		expectedPackages []string
	}{
		{
			packageName: "x/y",
			importPath:  "",
			expectedPackages: []string{
				"vendor/x/y",
				"x/y",
			},
		},
		{
			packageName: "x/y",
			importPath:  "a",
			expectedPackages: []string{
				"a/vendor/x/y",
				"vendor/x/y",
				"x/y",
			},
		},
		{
			packageName: "x/y",
			importPath:  "a/b",
			expectedPackages: []string{
				"a/b/vendor/x/y",
				"a/vendor/x/y",
				"vendor/x/y",
				"x/y",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("packageName=%s importPath=%s", tc.packageName, tc.importPath), func(t *testing.T) {
			if actualPaths := vendorPaths(tc.packageName, tc.importPath); !reflect.DeepEqual(tc.expectedPackages, actualPaths) {
				t.Errorf("Expected paths %v but got %v", tc.expectedPackages, actualPaths)
			}
		})
	}
}
