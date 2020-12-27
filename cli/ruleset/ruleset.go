package ruleset

import (
	"errors"
	"fmt"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/clintjedwards/tfvet/cli/formatter"
	"github.com/spf13/cobra"
)

// CmdRuleset is a subcommand for ruleset
var CmdRuleset = &cobra.Command{
	Use:   "ruleset",
	Short: "Manage linting rulesets",
	Long: `Manage linting rulesets.

Rulesets are a grouping of rules that are used to lint documents.

The ruleset subcommand allows you to retrieve, remove, and otherwise manipulate particular rulesets.`,
}

// state contains a bunch of useful state information for the add cli function. This is mostly
// just for convenience.
type state struct {
	fmt *formatter.Formatter
	cfg *appcfg.Appcfg
}

// newState returns a new state object with the fmt initialized
func newState(initialFmtMsg, format string) (*state, error) {

	clifmt, err := formatter.New(initialFmtMsg, formatter.Mode(format))
	if err != nil {
		return nil, err
	}

	cfg, err := appcfg.GetConfig()
	if err != nil {
		errText := fmt.Sprintf("config file `%s` does not exist."+
			" Run `tfvet init` to create.", appcfg.ConfigFilePath())
		clifmt.PrintFinalError(errText)
		return nil, errors.New(errText)
	}

	return &state{
		fmt: clifmt,
		cfg: cfg,
	}, nil
}
