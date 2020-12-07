package ruleset

import "github.com/spf13/cobra"

// CmdRuleset is a subcommand for ruleset
var CmdRuleset = &cobra.Command{
	Use:   "ruleset",
	Short: "Manage which rules are applied when linting",
}
