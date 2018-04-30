package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alecholmes/tdiff/app"
	"github.com/alecholmes/tdiff/importer"
)

var (
	// Required input flags
	packageFlag = flag.String("package", "", "Package to find reachable diff from")
	shaFlag     = flag.String("sha", "", "Git SHA")

	// Optional flags
	artifactsFlag = flag.Bool("artifacts", false, "If true, includes changed non-Go source files under the package directory, recursive")
	verboseFlag   = flag.Bool("verbose", false, "If set, log verbose debugging information")

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

	includePaths := *jsonFlag
	logger := app.NoLogging
	if *verboseFlag {
		logger = log.Printf
	}

	differ := app.NewDiffer(os.Getenv("GOPATH"), importer.DefaultRecursiveImport, *commitsFlag, includePaths, logger)

	summary, err := differ.Diff(*packageFlag, *shaFlag, *artifactsFlag)
	if err != nil {
		log.Fatal(err)
	}

	if *packagesFlag {
		for _, pkg := range summary.Packages {
			fmt.Println(pkg.ImportPath)
		}
	}

	if *filesFlag {
		for _, file := range summary.Files {
			fmt.Println(file)
		}
	}

	if *commitsFlag {
		for _, commit := range summary.Commits {
			fmt.Printf("%s %s\n", commit.SHA, commit.Description)
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
