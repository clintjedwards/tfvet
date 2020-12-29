package rule

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRuleDisable = &cobra.Command{
	Use:   "disable <ruleset> <rule>",
	Short: "Turns off a rule",
	Long:  `Turns off a particular rule, skipping it when linting is run.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runDisable,
}

func runDisable(cmd *cobra.Command, args []string) error {
	ruleset := args[0]
	rule := args[1]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Disabling rule", format)
	if err != nil {
		return err
	}

	state.fmt.PrintMsg(rule)

	err = state.cfg.DisableRule(ruleset, rule)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not disable rule: %v", err))
		return err
	}

	state.fmt.PrintFinalSuccess(fmt.Sprintf("Disabled rule %s", rule))
	return nil
}

func init() {
	CmdRule.AddCommand(cmdRuleDisable)
}
