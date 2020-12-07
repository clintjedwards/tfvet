package ruleset

import "github.com/spf13/cobra"

var cmdRulesetList = &cobra.Command{
	Use:   "list",
	Short: "test",
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetList)
}
