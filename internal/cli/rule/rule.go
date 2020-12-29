package rule

import (
	"errors"
	"fmt"

	"github.com/clintjedwards/tfvet/internal/cli/appcfg"
	"github.com/clintjedwards/tfvet/internal/cli/formatter"
	"github.com/spf13/cobra"
)

// CmdRule is a subcommand for rule
var CmdRule = &cobra.Command{
	Use:   "rule",
	Short: "Manage linting rules",
	Long: `Manage linting rules.

Rules are the constraints on which tfvet lints documents against.

The rule subcommand allows you to describe, enable, and otherwise manipulate particular rules.`,
}

// state contains a bunch of useful state information for cli functions.
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
