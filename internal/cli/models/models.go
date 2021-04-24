// Package models represents data structure which are shared between packages.
// Changes made here will regenerate the protobuf definitions and maybe require an version
// bump for the sdk.
package models

// Ruleset represents a packaged set of rules that govern what tfvet checks for.
type Ruleset struct {
	Name       string `hcl:"name,label" json:"name"`
	Version    string `hcl:"version" json:"version"`
	Repository string `hcl:"repository" json:"repository"`
	Enabled    bool   `hcl:"enabled" json:"enabled"`
	Rules      []Rule `hcl:"rule,block" json:"rules"`
}

// Rule represents a single lint check within a ruleset.
type Rule struct {
	ID      string `hcl:"id,label" json:"id"`
	Name    string `hcl:"name" json:"name"`
	Short   string `hcl:"short" json:"short"`
	Long    string `hcl:"long" json:"long"`
	Link    string `hcl:"link" json:"link"`
	Enabled bool   `hcl:"enabled" json:"enabled"`
}
