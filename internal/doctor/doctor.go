package doctor

import "golang.org/x/tools/go/packages"

type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

type Diagnosis struct {
	Severity     Severity
	Message      string
	WhyItMatters string
	File         string
	Line         int
	CodeSnippet  string
}

type Check interface {
	Name() string
	Run(pkg *packages.Package) []Diagnosis
}
