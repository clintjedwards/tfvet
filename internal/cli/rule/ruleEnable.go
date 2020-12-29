package rule

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRuleEnable = &cobra.Command{
	Use:   "enable <ruleset> <rule>",
	Short: "Turns on a rule",
	Long:  `Turns on a particular rule, allowing it to run during linting.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runEnable,
}

func runEnable(cmd *cobra.Command, args []string) error {
	ruleset := args[0]
	rule := args[1]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Enabling rule", format)
	if err != nil {
		return err
	}

	state.fmt.PrintMsg(rule)

	err = state.cfg.EnableRule(ruleset, rule)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not enable rule: %v", err))
		return err
	}

	state.fmt.PrintFinalSuccess(fmt.Sprintf("Enabled rule %s", rule))
	return nil
}

func init() {
	CmdRule.AddCommand(cmdRuleEnable)
}