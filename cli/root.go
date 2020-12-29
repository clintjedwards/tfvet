package cli

import (
	"log"
	"os"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/clintjedwards/tfvet/cli/rule"
	"github.com/clintjedwards/tfvet/cli/ruleset"
	"github.com/clintjedwards/tfvet/utils"
	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev_000000_33333"

// RootCmd is the base of the cli
var RootCmd = &cobra.Command{
	Use:   "tfvet",
	Short: "tfvet is a Terraform linter",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Including these in the pre run hook instead of in the enclosing command definition
		// allows cobra to still print errors and usage for its own cli verifications, but
		// ignore our errors.
		cmd.SilenceUsage = true  // Don't print the usage if we get an upstream error
		cmd.SilenceErrors = true // Let us handle error printing ourselves

		// Make sure the configuration is present on every run
		err := setup()
		return err
	},
	Version: appVersion,
}

func init() {
	RootCmd.AddCommand(ruleset.CmdRuleset)
	RootCmd.AddCommand(rule.CmdRule)

	RootCmd.PersistentFlags().StringP("format", "f", "pretty",
		"output format; accepted values are 'pretty', 'json', and, 'plain'")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func setup() error {

	err := utils.CreateDir(appcfg.ConfigPath())
	if err != nil {
		log.Println(err)
		return err
	}
	err = utils.CreateDir(appcfg.RulesetsPath())
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = os.Stat(appcfg.ConfigFilePath())
	if os.IsNotExist(err) {
		err = appcfg.CreateNewFile()
		if err != nil {
			log.Println(err)
			return err
		}
	} else if os.IsExist(err) {
	} else if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
