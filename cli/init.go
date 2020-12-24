package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/clintjedwards/tfvet/cli/formatter"
	"github.com/clintjedwards/tfvet/cli/tfvetcfg"
	"github.com/clintjedwards/tfvet/utils"
	"github.com/spf13/cobra"
)

// cmdInit initializes the initialing of required directories and configuration files for
// tfvet
var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Create required configuration files and ruleset directory",
	Long: `Init is used to instantiate a newly installed tfvet.

It does two important things:

1) It creates a default ~/.tfvet.d/.tfvet.hcl file (used for configuration settings)
2) It creates the ~/.tfvet.d/rulesets directory (used to store rulesets)
`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	clifmt, err := formatter.New("Initializing", formatter.Mode(format))
	if err != nil {
		clifmt.PrintFinalError(fmt.Sprintf("could not create tfvet dirs: %v", err))
		return err
	}

	//TODO(clintjedwards): omit the created if it already exists

	// Create directories
	directories := []string{tfvetcfg.ConfigDir, tfvetcfg.RulesetsDir}
	clifmt.PrintMsg(fmt.Sprintf("Creating directories: %s", strings.Join(directories, ", ")))

	err = utils.CreateDirectories(directories...)
	if err != nil {
		clifmt.PrintFinalError(fmt.Sprintf("could not create tfvet dirs: %v", err))
		return err
	}

	for _, dir := range directories {
		clifmt.PrintSuccess(fmt.Sprintf("%q created", dir))
	}

	// Create configuration files
	clifmt.PrintMsg("Creating config files")

	err = tfvetcfg.CreateNewFile()
	if err != nil {
		clifmt.PrintFinalError(fmt.Sprintf("could not write tfvet config: %v", err))
		return err
	}
	clifmt.PrintSuccess(fmt.Sprintf("%q created", tfvetcfg.ConfigFilePath))

	clifmt.PrintFinalSuccess("init finished!")
	return nil
}

func init() {
	RootCmd.AddCommand(cmdInit)
}
