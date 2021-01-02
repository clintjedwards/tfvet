package ruleset

import "github.com/spf13/cobra"

// CmdCreate creates a skeleton ruleset
var CmdCreate = &cobra.Command{
	Use:   "generate",
	Short: "Manage linting rulesets",
	Long: `Manage linting rulesets.

Rulesets are a grouping of rules that are used to lint documents.

The ruleset subcommand allows you to retrieve, remove, and otherwise manipulate particular rulesets.`,
}
