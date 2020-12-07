package ruleset

import "github.com/spf13/cobra"

var cmdRulesetDisable = &cobra.Command{
	Use:   "disable <ruleset>",
	Short: "test",
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetDisable)
}
