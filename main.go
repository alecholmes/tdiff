package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/alecholmes/tdiff/app"
	"github.com/alecholmes/tdiff/importer"
)

var (
	// Required input flags
	packageFlag = flag.String("package", "", "Package to find reachable diff from")
	shaFlag     = flag.String("sha", "", "Git SHA after which changes will be considered (exclusive)")

	// Optional flags
	artifactsFlag = flag.Bool("artifacts", false, "If true, includes changed non-Go source files under the package directory, recursive")
	verboseFlag   = flag.Bool("verbose", false, "If set, log verbose debugging information")

	// Output format flags
	packagesFlag = flag.Bool("packages", false, "If set, all relevant changed packages printed")
	filesFlag    = flag.Bool("files", false, "If set, all relevant changed files are printed")
	commitsFlag  = flag.Bool("commits", false, "If set, all relevant commits are printed")
	jsonFlag     = flag.Bool("json", false, "If set, JSON object representing all changes is printed")
	htmlFlag     = flag.Bool("html", false, "If set, an HTML summary is written to a temp file")
)

func main() {
	flag.Parse()
	if len(*packageFlag) == 0 || len(*shaFlag) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	includePaths := *jsonFlag || *htmlFlag
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

	if *htmlFlag {
		fileName, err := writeHTML(summary)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(fileName)
	}
}

func writeHTML(summary *app.Summary) (string, error) {
	body, err := app.HTML(summary)
	if err != nil {
		return "", err
	}

	file, err := ioutil.TempFile("", summary.SHA)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.Write(body); err != nil {
		return "", err
	}
	file.Close()

	newName := fmt.Sprintf("%s.html", file.Name())
	if err := os.Rename(file.Name(), newName); err != nil {
		return "", err
	}

	return newName, nil
}
