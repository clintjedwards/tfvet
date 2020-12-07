package ruleset

import "github.com/spf13/cobra"

var cmdRulesetRemoveAll = &cobra.Command{
	Use:   "removeall",
	Short: "test",
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetDisable)
}
