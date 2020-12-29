package ruleset

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRulesetEnable = &cobra.Command{
	Use:   "enable <ruleset>",
	Short: "Turns on a ruleset",
	Long:  `Turns on a particular ruleset, allowing it to run during linting`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEnable,
}

func runEnable(cmd *cobra.Command, args []string) error {
	ruleset := args[0]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Enabling ruleset", format)
	if err != nil {
		return err
	}

	state.fmt.PrintMsg(ruleset)

	err = state.cfg.EnableRuleset(ruleset)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not enable ruleset: %v", err))
		return err
	}

	state.fmt.PrintFinalSuccess(fmt.Sprintf("Enabled ruleset %s", ruleset))
	return nil
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetEnable)
}
