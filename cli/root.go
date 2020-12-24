package cli

import (
	ruleset "github.com/clintjedwards/tfvet/cli/ruleset"
	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev_000000_33333"

// RootCmd is the base of the cli
var RootCmd = &cobra.Command{
	Use:           "tfvet",
	Short:         "tfvet is a Terraform linter",
	SilenceUsage:  true, // Don't print the usage if we get an upstream error
	SilenceErrors: true, // Let us handle error printing ourselves
	Version:       appVersion,
}

func init() {
	RootCmd.AddCommand(ruleset.CmdRuleset)

	RootCmd.PersistentFlags().StringP("format", "f", "pretty",
		"output format; accepted values are 'pretty', 'json', and, 'plain'")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}
