package rule

import "github.com/spf13/cobra"

// cmdRuleCreate creates a skeleton ruleset
var cmdRuleCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new rule",
	Long: `Creates the files and folders needed to create a new rule.

Navigate to the ruleset folder in which you mean to create the rule. From there, simply run this command
to create all files and folders required for a tfvet rule.
`,
	Example: `$ tfvet rule create example_rule`,
	RunE:    runCreate,
	Args:    cobra.ExactArgs(1),
}

func init() {
	CmdRule.AddCommand(cmdRuleCreate)
}

func runCreate(cmd *cobra.Command, args []string) error {
	return nil
}
