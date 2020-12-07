package ruleset

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cmdRulesetEnable = &cobra.Command{
	Use:   "enable <ruleset>",
	Short: "test",
	Run:   runTest,
}

func runTest(cmd *cobra.Command, args []string) {
	fmt.Println("lol")
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetEnable)
}
