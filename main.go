package main

import (
	"fmt"
	"os"
	"path/filepath"

	"goHealth/internal/checks"
	"goHealth/internal/doctor"
	"goHealth/internal/report"

	"golang.org/x/tools/go/packages"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: goHealth check <path>")
		os.Exit(1)
	}

	command := os.Args[1]
	if command != "check" {
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	// Ensure path is absolute for clearer reporting
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}

	fmt.Printf("Health Inspector visiting: %s\n", path)

	// Load packages with type info
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, path+"/...")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Checkup failed: %v\n", err)
		os.Exit(1)
	}

	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	// Register Checks
	registry := []doctor.Check{
		&checks.SleepInLoop{},
		&checks.ContextInStruct{},
		&checks.UnclosedBody{},
		&checks.SQLInjection{},
		&checks.RowsClose{},
		&checks.TimebombHttpClient{},
		&checks.BrokenLinkContext{},
		&checks.ZombieTransaction{},
		&checks.ExitInLibrary{},
		&checks.MemoryDOSGuard{},
		&checks.SilencedErrors{},
		&checks.SliceAppendRace{},
		&checks.WeakRandomness{},
		// Add more checks here
	}

	var all []doctor.Diagnosis

	for _, pkg := range pkgs {
		for _, check := range registry {
			results := check.Run(pkg)
			all = append(all, results...)
		}
	}

	report.Render(all)
}
