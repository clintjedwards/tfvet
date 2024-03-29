package cli

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/clintjedwards/tfvet/v2/internal/cli/appcfg"
	"github.com/clintjedwards/tfvet/v2/internal/cli/rule"
	"github.com/clintjedwards/tfvet/v2/internal/cli/ruleset"
	"github.com/clintjedwards/tfvet/v2/internal/utils"
	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev_000000_33333"

// RootCmd is the base of the cli
var RootCmd = &cobra.Command{
	Use:   "tfvet",
	Short: "tfvet is a Terraform linter",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// Including these in the pre run hook instead of in the enclosing command definition
		// allows cobra to still print errors and usage for its own cli verifications, but
		// ignore our errors.
		cmd.SilenceUsage = true  // Don't print the usage if we get an upstream error
		cmd.SilenceErrors = true // Let us handle error printing ourselves

		// Make sure the configuration is present on every run
		err := setup()
		return err
	},
	Version: " ", // We leave this added but empty so that the rootcmd will supply the -v flag
}

func init() {
	RootCmd.SetVersionTemplate(humanizeVersion(appVersion))
	RootCmd.AddCommand(ruleset.CmdRuleset)
	RootCmd.AddCommand(rule.CmdRule)

	RootCmd.PersistentFlags().StringP("format", "f", "pretty",
		"output format; accepted values are 'pretty', 'json', 'silent'")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func humanizeVersion(version string) string {
	splitVersion := strings.Split(version, "_")

	semver := splitVersion[0]
	hash := splitVersion[1]
	i, _ := strconv.Atoi(splitVersion[2])
	unixTime := time.Unix(int64(i), 0)
	time := unixTime.Format("Mon Jan 2 15:04 2006")

	return fmt.Sprintf("tfvet %s [%s] %s\n", semver, hash, time)
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
