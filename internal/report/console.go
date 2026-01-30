package report

import (
	"fmt"
	"sort"
	"strings"

	"goHealth/internal/doctor"
)

func Render(diagnoses []doctor.Diagnosis) {
	if len(diagnoses) == 0 {
		fmt.Println("\n✅  Doctor's Orders: clear bill of health! No critical issues found.")
		return
	}

	fmt.Printf("\n🩺  Doctor's Findings (%d issues):\n\n", len(diagnoses))

	// Group by file? Or just list? Listing is fine for now.
	// Let's sort by severity first.
	sort.Slice(diagnoses, func(i, j int) bool {
		// Critical first
		if diagnoses[i].Severity == doctor.SeverityCritical && diagnoses[j].Severity != doctor.SeverityCritical {
			return true
		}
		return false
	})

	for _, d := range diagnoses {
		printDiagnosis(d)
		fmt.Println(strings.Repeat("-", 40))
	}
}

func printDiagnosis(d doctor.Diagnosis) {
	icon := "⚠️"
	color := "\033[33m" // Yellow
	reset := "\033[0m"

	if d.Severity == doctor.SeverityCritical {
		icon = "🚨"
		color = "\033[31m" // Red
	}

	fmt.Printf("%s %s[%s]%s %s\n", icon, color, d.Severity, reset, d.Message)
	fmt.Printf("   📍 Location: %s:%d\n", d.File, d.Line)

	if d.CodeSnippet != "" {
		fmt.Printf("   📝 Code: %s\n", strings.TrimSpace(d.CodeSnippet))
	}

	fmt.Printf("\n   🎓 %sWhy this matters:%s\n", "\033[1m", reset) // Bold
	fmt.Printf("   %s\n\n", d.WhyItMatters)
}
