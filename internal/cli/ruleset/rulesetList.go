package ruleset

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRulesetList = &cobra.Command{
	Use:   "list [ruleset]",
	Short: "Lists a ruleset and its rules",
	Long: `Allows the listing of a ruleset and its rules.

If no argument is provided, list will display all possible rulesets and relevant details.
If a ruleset is provided, list will display the ruleset's details and rules.
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("", format)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		state.fmt.PrintStandaloneMsg("== Summary ==\n\n")
		state.fmt.PrintAllRulesets(state.cfg.Rulesets)
		state.fmt.PrintStandaloneMsg("\n== Rulesets ==\n\n")
		for index, ruleset := range state.cfg.Rulesets {
			state.fmt.PrintRuleset(ruleset)
			if index != len(state.cfg.Rulesets)-1 {
				state.fmt.PrintStandaloneMsg("\n\n")
			}
		}
		return nil
	}

	rulesetName := args[0]
	ruleset, err := state.cfg.GetRuleset(rulesetName)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not find ruleset %s", rulesetName))
		return err
	}
	state.fmt.PrintRuleset(ruleset)
	return nil
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetList)
}
