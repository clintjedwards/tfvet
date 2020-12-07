package ruleset

import "github.com/spf13/cobra"

// Check config file for rulesets, check rulesets for

var cmdRulesetUpdate = &cobra.Command{
	Use:   "update",
	Short: "Retrieves most recent rulesets",
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetUpdate)
}
