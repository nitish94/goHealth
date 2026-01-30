package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"goHealth/internal/checks"
	"goHealth/internal/doctor"
	"goHealth/internal/report"

	"golang.org/x/tools/go/packages"
)

func main() {
	output := flag.String("o", "", "Output file for HTML report")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: goHealth [-o output.html] check <path>")
		os.Exit(1)
	}

	command := args[0]
	if command != "check" {
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	path := "."
	if len(args) > 1 {
		path = args[1]
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
		&checks.TimeComparison{},
		&checks.EmptySpin{},
		// Add more checks here
	}

	var all []doctor.Diagnosis

	for _, pkg := range pkgs {
		for _, check := range registry {
			results := check.Run(pkg)
			all = append(all, results...)
		}
	}

	if *output != "" {
		if err := generateHTMLReport(all, *output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate HTML report: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("HTML report generated: %s\n", *output)
	} else {
		report.Render(all)
	}
}

func generateHTMLReport(diagnoses []doctor.Diagnosis, outputFile string) error {
	// Sort by severity
	sort.Slice(diagnoses, func(i, j int) bool {
		if diagnoses[i].Severity == doctor.SeverityCritical && diagnoses[j].Severity != doctor.SeverityCritical {
			return true
		}
		return false
	})

	html := `<html>
<head>
<title>Go Health Report</title>
</head>
<body>
<h1>Go Health Inspector Report</h1>
<p>Found ` + fmt.Sprintf("%d", len(diagnoses)) + ` issues.</p>
`

	for _, d := range diagnoses {
		severity := string(d.Severity)
		html += "<h2>" + severity + ": " + d.Message + "</h2>\n"
		html += "<p><strong>Location:</strong> " + d.File + ":" + fmt.Sprintf("%d", d.Line) + "</p>\n"
		if d.CodeSnippet != "" {
			html += "<p><strong>Code:</strong> " + d.CodeSnippet + "</p>\n"
		}
		html += "<p><strong>Why this matters:</strong></p>\n"
		html += "<p>" + strings.ReplaceAll(d.WhyItMatters, "\n", "<br>") + "</p>\n"
		if d.Suggestion != "" {
			html += "<p><strong>Suggestion:</strong></p>\n"
			html += "<p>" + strings.ReplaceAll(d.Suggestion, "\n", "<br>") + "</p>\n"
		}
		html += "<p></p>\n" // padding
	}

	html += "</body>\n</html>\n"

	return os.WriteFile(outputFile, []byte(html), 0644)
}
