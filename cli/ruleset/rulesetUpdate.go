package ruleset

import (
	"fmt"
	"log"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/spf13/cobra"
)

var cmdRulesetUpdate = &cobra.Command{
	Use:   "update [ruleset]",
	Short: "Retrieves most recent rules for a ruleset",
	Long: `Update will attempt to download rule updates for a specific ruleset.

Running without arguments will update all rulesets.

It works by comparing the version of the ruleset currently installed with the remote ruleset.
If the version downloaded is different than the local version this will trigger a
recompilation of all rules.

The resolution process is very basic and does not perform any more than a rudimentary check for diffs
and as such, for sufficiently large repositories this might be a heavy operation.
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetUpdate)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Updating ruleset", format)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(args) == 0 {
		for _, ruleset := range state.cfg.Rulesets {
			state.fmt.UpdateSuffix(fmt.Sprintf("Updating ruleset %s", ruleset.Name))
			err := updateRuleset(state, ruleset)
			if err != nil {
				state.fmt.PrintFinalError(fmt.Sprintf("could not update ruleset %s", ruleset.Name))
				return err
			}
		}
		state.fmt.PrintFinalSuccess("Updated all rulesets")
		return nil
	}

	rulesetName := args[0]
	state.fmt.UpdateSuffix(fmt.Sprintf("Updating ruleset %s", rulesetName))
	ruleset, err := state.cfg.GetRuleset(rulesetName)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not find ruleset %s", rulesetName))
		return err
	}
	err = updateRuleset(state, ruleset)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not update ruleset %s", ruleset.Name))
		return err
	}
	state.fmt.PrintFinalSuccess("Updated all rulesets")

	return nil
}

func updateRuleset(s *state, ruleset appcfg.Ruleset) error {

	s.fmt.PrintMsg("Retrieveing ruleset")
	err := getRemoteRuleset(ruleset.Repository, appcfg.RepoPath(ruleset.Name))
	if err != nil {
		return err
	}

	s.fmt.PrintMsg("Parsing remote info")
	info, err := getRemoteRulesetInfo(appcfg.RepoPath(ruleset.Name))
	if err != nil {
		return err
	}

	s.fmt.PrintMsg("Verifying ruleset")
	err = verifyRuleset(appcfg.RepoPath(ruleset.Name), info)
	if err != nil {
		return err
	}

	// TODO(clintjedwards): We should probably parse this into semver first
	if info.Version == ruleset.Version {
		s.fmt.PrintSuccess(fmt.Sprintf("Ruleset %s at newest version (%s)", ruleset.Name, ruleset.Version))
		return nil
	}

	s.fmt.PrintSuccess(fmt.Sprintf("Found newer ruleset for %s (current: %s, remote: %s)",
		ruleset.Name, ruleset.Version, info.Version))

	s.fmt.PrintMsg("Updating ruleset")
	err = s.cfg.UpdateRuleset(appcfg.Ruleset{
		Name:       info.Name,
		Version:    info.Version,
		Repository: ruleset.Repository,
		Enabled:    ruleset.Enabled,
		Rules:      ruleset.Rules,
	})
	if err != nil {
		return err
	}

	err = buildRulesetRules(s, info.Name)
	if err != nil {
		return err
	}

	return nil
}
