package ruleset

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRulesetDisable = &cobra.Command{
	Use:   "disable <ruleset>",
	Short: "Turns off a ruleset",
	Long:  `Turns off a particular ruleset, skipping it when linting is run.`,
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

	err = state.cfg.DisableRuleset(ruleset)
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
