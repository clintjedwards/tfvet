package ruleset

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/clintjedwards/tfvet/cli/formatter"
	getter "github.com/hashicorp/go-getter/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
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

// state contains a bunch of useful state information for cli functions.
type state struct {
	fmt *formatter.Formatter
	cfg *appcfg.Appcfg
}

type rulesetInfo struct {
	Name    string `hcl:"name"`
	Version string `hcl:"version"`
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

// getRemoteRuleset is used to retrieve a ruleset from any path given
// (supports a wide range of remote and local)
//
// See  https://github.com/hashicorp/go-getter#url-format formats accepted
func getRemoteRuleset(srcPath, dstPath string) error {
	_, err := getter.Get(context.Background(), dstPath, srcPath)
	if err != nil {
		return err
	}

	return err
}

func getRemoteRulesetInfo(repoPath string) (rulesetInfo, error) {
	var info rulesetInfo

	rulesetFilePath := fmt.Sprintf("%s/%s", repoPath, "ruleset.hcl")
	err := hclsimple.DecodeFile(rulesetFilePath, nil, &info)
	if err != nil {
		return rulesetInfo{}, err
	}

	return info, nil
}

// buildRulesetRules builds the rules plugins and places the binary underneath the correct ruleset.
// TODO(clintjedwards): Remove the addRule bool
func buildRulesetRules(s *state, ruleset string) error {
	s.fmt.PrintMsg("Opening rules directory")

	file, err := os.Open(appcfg.RepoRulesPath(ruleset))
	if err != nil {
		errText := fmt.Sprintf("could not open rules folder: %v", err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}
	defer file.Close()

	fileList, err := file.Readdir(0)
	if err != nil {
		errText := fmt.Sprintf("could not read rules folder: %v", err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	startTime := time.Now()
	count := 0

	// Rules are separated into directories. We iterate through directories and build whats inside
	// them.
	for _, file := range fileList {
		if !file.IsDir() {
			continue
		}

		// Get the filename and not the full path
		// Sometimes file.Name will return the full path based on what is passed to file.Open
		fileName := filepath.Base(file.Name())

		s.fmt.PrintMsg(fmt.Sprintf("Compiling %s", fileName))

		rawRulePath := fmt.Sprintf("%s/%s", appcfg.RepoRulesPath(ruleset), fileName)

		// we take the hash of the filename(aka the rule folder name) and make it the rule ID
		ruleID := generateHash(fileName)

		_, err := buildRule(rawRulePath, appcfg.RulePath(ruleset, ruleID))
		if err != nil {
			errText := fmt.Sprintf("could not build rule %s: %v", fileName, err)
			s.fmt.PrintFinalError(errText)
			return errors.New(errText)
		}

		err = s.getRuleInfo(ruleset, ruleID)
		if err != nil {
			return err
		}
		count++
	}

	duration := time.Since(startTime)
	durationSeconds := float64(duration) / float64(time.Second)
	timePerRule := float64(duration) / float64(count)

	s.fmt.PrintSuccess(fmt.Sprintf("Compiled %d rules in %.2fs (average %.2fms/rule)",
		count, durationSeconds, timePerRule/float64(time.Millisecond)))

	return nil
}

// verifyRuleset makes sure a downloaded ruleset has the correct structure.
// specifically it:
//
//	* Makes sure the ruleset has a parseable ruleset.hcl file.
//	* Makes sure the ruleset has a version and a name.
//	* Makes sure the ruleset has a rules folder.
func verifyRuleset(path string, info rulesetInfo) error {

	//TODO(clintjedwards): Better validation of these
	if len(info.Name) < 3 {
		return errors.New("ruleset name cannot be less than 3 characters")
	}

	if len(info.Version) < 5 {
		return errors.New("ruleset version text malformed; should be in semvar notation")
	}

	rulesDirPath := fmt.Sprintf("%s/%s", path, "rules")
	if _, err := os.Stat(rulesDirPath); os.IsNotExist(err) {
		return errors.New("no rules directory found; all rulesets must have a rules directory")
	}

	return nil
}
