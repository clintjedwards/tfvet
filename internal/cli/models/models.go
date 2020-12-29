// Package models represents data structure which are shared between packages.
// Changes made here will regenerate the protobuf definitions and maybe require an version
// bump for the sdk.
package models

// Ruleset represents a packaged set of rules that govern what tfvet checks for.
type Ruleset struct {
	Name       string `hcl:"name,label"`
	Version    string `hcl:"version"`
	Repository string `hcl:"repository"`
	Enabled    bool   `hcl:"enabled"`
	Rules      []Rule `hcl:"rule,block"`
}

// Severity is used to convey how serious the offending error is. This is passed in the output
// of tfvet linting errors so that downstream tools can choose whether to surface them or not.
type Severity string

const (
	SeverityUnknown  Severity = "unknown"
	SeverityInfo              = "info"
	SeverityWarning           = "warning"
	SeverityError             = "error"
	SeverityCritical          = "critical"
)

// Rule represents a single lint check within a ruleset.
type Rule struct {
	ID       string `hcl:"id,label"`
	Name     string `hcl:"name"`
	Short    string `hcl:"short"`
	Long     string `hcl:"long"`
	Severity string `hcl:"severity"`
	Link     string `hcl:"link"`
	Enabled  bool   `hcl:"enabled"`
}
