package rule

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
	ruleID := args[1]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("", format)
	if err != nil {
		return err
	}

	rule, err := state.cfg.GetRule(ruleset, ruleID)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not describe rule %v", err))
		return err
	}

	// Example of output
	//
	// 	[d2d21] No resource with the name 'example'
	//
	// Example is a poor name for a resource and might lead to naming collisions.
	//
	// This is simply a test description of a resource that effectively alerts on nothingness.
	// In turn this is essentially a really long description so we can test that our descriptions
	// work properly and are displayed properly in the terminal.
	//
	// Enabled: true | Link: https://google.com
	state.fmt.PrintStandaloneMsg(fmt.Sprintf("[%s] %s\n\n", rule.ID, rule.Name))
	state.fmt.PrintStandaloneMsg(rule.Short + "\n\n")
	state.fmt.PrintStandaloneMsg(strings.TrimPrefix(rule.Long, "\n") + "\n")
	state.fmt.PrintStandaloneMsg(fmt.Sprintf("Enabled: %s | Link: %s",
		strconv.FormatBool(rule.Enabled), rule.Link))

	return nil
}

func init() {
	CmdRule.AddCommand(cmdRuleDescribe)
}
