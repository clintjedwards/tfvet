package rule

import (
	"fmt"
	"log"
	"strconv"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/spf13/cobra"
)

var cmdRuleDescribe = &cobra.Command{
	Use:   "describe <ruleset> <rule>",
	Short: "Prints details about a rule",
	Long:  `Prints extended information about a particular rule.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runDescribe,
}

func runDescribe(cmd *cobra.Command, args []string) error {
	ruleset := args[0]
	ruleName := args[1]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("", format)
	if err != nil {
		log.Println(err)
		return err
	}

	rule, err := state.cfg.GetRule(ruleset, ruleName)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not describe rule %v", err))
		return err
	}

	state.fmt.PrintStandaloneMsg(fmt.Sprintf("%s(%s)", rule.Name, rule.FileName))
	state.fmt.PrintStandaloneMsg("")
	state.fmt.PrintStandaloneMsg(rule.Short)
	state.fmt.PrintStandaloneMsg("")
	state.fmt.PrintStandaloneMsg(rule.Long)
	state.fmt.PrintStandaloneMsg("")
	state.fmt.PrintStandaloneMsg(fmt.Sprintf("Severity: %s | Enabled: %s | Link: %s",
		appcfg.SeverityToString(rule.Severity), strconv.FormatBool(rule.Enabled), rule.Link))

	return nil
}

func init() {
	CmdRule.AddCommand(cmdRuleDescribe)
}
