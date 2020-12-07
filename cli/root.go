package cli

import (
	ruleset "github.com/clintjedwards/tfvet/cli/ruleset"
	"github.com/spf13/cobra"
)

// RootCmd is the base of the cli
var RootCmd = &cobra.Command{
	Use:           "tfvet",
	Short:         "tfvet is a Terraform linter",
	SilenceUsage:  true, // Don't print the usage if we get an upstream error
	SilenceErrors: true, // Let us handle error printing ourselves
	Version:       "0.0.0",
}

func init() {
	RootCmd.AddCommand(ruleset.CmdRuleset)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}
