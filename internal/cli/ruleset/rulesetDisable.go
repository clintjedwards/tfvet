package ruleset

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRulesetDisable = &cobra.Command{
	Use:   "disable <ruleset>",
	Short: "Turns off a ruleset",
	Long:  `Turns off a particular ruleset, causing it to be skipped on all future linting runs.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDisable,
}

func runDisable(cmd *cobra.Command, args []string) error {
	ruleset := args[0]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Disabling ruleset", format)
	if err != nil {
		return err
	}

	state.fmt.PrintMsg(ruleset)

	err = state.cfg.SetRulesetEnabled(ruleset, false)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not disable ruleset: %v", err))
		return err
	}

	state.fmt.PrintFinalSuccess(fmt.Sprintf("Disabled ruleset %s", ruleset))
	return nil
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetDisable)
}
