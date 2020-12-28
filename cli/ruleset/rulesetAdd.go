package ruleset

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"hash/fnv"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	tfvetPlugin "github.com/clintjedwards/tfvet/plugin"
	"github.com/clintjedwards/tfvet/plugin/proto"
	"github.com/clintjedwards/tfvet/utils"
	getter "github.com/hashicorp/go-getter/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

const (
	golangBinaryName = "go"
)

var cmdRulesetAdd = &cobra.Command{
	Use:   "add <repository>",
	Short: "Downloads, adds, and enables a new ruleset",
	Long: `The add command downloads, adds, and enables a new tfvet ruleset.

Supports a wide range of sources including github url, fileserver path, or even just
the path to a local directory.

See https://github.com/hashicorp/go-getter#url-format for more information on how to form input.

Arguments:

• <repository> is the location of the ruleset repository. Ruleset repositories must adhere to the
following rules:

  • Repository must contain a ruleset.hcl file containing name and version.
  • Repository must contain a rules folder with rules plugins built with tfvet sdk.

For more information on tfvet ruleset repository requirements see: TODO(clintjedwards):
`,
	Example: `$ tfvet add github.com/example/tfvet-ruleset-aws
$ tfvet add ~/tmp/tfvet-ruleset-example`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

type rulesetInfo struct {
	Name    string `hcl:"name"`
	Version string `hcl:"version"`
}

// getRuleset is used to retrieve a ruleset from any path given
// (supports a wide range of remote and local)
//
// Returns temporary ruleset download path
func (s *state) getRuleset(location string) (string, error) {
	s.fmt.PrintMsg(fmt.Sprintf("Retrieving %s", location))

	tmpDownloadPath := fmt.Sprintf("%s/tfvet_%s", os.TempDir(), generateHash(location))
	_, err := getter.Get(context.Background(), tmpDownloadPath, location)
	if err != nil {
		errText := fmt.Sprintf("could not download ruleset: %v", err)
		s.fmt.PrintFinalError(errText)
		return "", errors.New(errText)
	}

	s.fmt.PrintSuccess(fmt.Sprintf("Retrieved %s", location))
	return tmpDownloadPath, err
}

// getRulesetInfo makes sure a downloaded ruleset has the correct structure and collects its
// information.
//
// specifically it:
//
//	* Makes sure the ruleset has a parseable ruleset.hcl file.
//	* Makes sure the ruleset has a version and a name.
//	* Makes sure the ruleset has a rules folder.
func (s *state) getRulesetInfo(repositoryPath string) (rulesetInfo, error) {

	s.fmt.PrintMsg("Verifying ruleset")

	var info rulesetInfo

	// Check for parseable ruleset file and get info
	rulesetFilePath := fmt.Sprintf("%s/%s", repositoryPath, "ruleset.hcl")

	err := hclsimple.DecodeFile(rulesetFilePath, nil, &info)
	if err != nil {
		errText := fmt.Sprintf("could not verify ruleset: %v", err)
		s.fmt.PrintFinalError(errText)
		return rulesetInfo{}, errors.New(errText)
	}

	//TODO(clintjedwards): Better validation of these
	if len(info.Name) < 3 {
		errText := "ruleset name cannot be less than 3 characters"
		s.fmt.PrintFinalError(errText)
		return rulesetInfo{}, errors.New(errText)
	}

	if len(info.Version) < 5 {
		errText := "ruleset version text malformed; should be in semvar notation"
		s.fmt.PrintFinalError(errText)
		return rulesetInfo{}, errors.New(errText)
	}

	rulesDirPath := fmt.Sprintf("%s/%s", repositoryPath, "rules")

	if _, err := os.Stat(rulesDirPath); os.IsNotExist(err) {
		errText := "no rules directory found; all rulesets must have a rules directory"
		s.fmt.PrintFinalError(errText)
		return rulesetInfo{}, errors.New(errText)
	}

	s.fmt.PrintSuccess("Verified ruleset")

	return info, nil
}

// move a downloaded repo from the temporary path to the well known repo path for the ruleset.
func (s *state) moveRepo(ruleset, tmpPath string) error {

	s.fmt.PrintMsg("Moving ruleset to permanent config location")

	err := utils.CreateDir(appcfg.RulesetPath(ruleset))
	if err != nil {
		errText := fmt.Sprintf("could not create parent directory: %v", err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	err = copy.Copy(tmpPath, appcfg.RepoPath(ruleset))
	if err != nil {
		errText := fmt.Sprintf("could not copy ruleset to config directory: %v", err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	s.fmt.PrintSuccess("New ruleset added")

	return nil
}

//TODO(clintjedwards): When you hit a rule you can't build continue to other rules
// if 3 built rules fail in a row, there is a bigger problem and we should return immediately
//
// buildRulesetRules builds the rules plugins and places the binary underneath the correct ruleset.
func (s *state) buildRulesetRules(ruleset string) error {
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

func generateHash(filename string) string {
	digest := fnv.New32()
	_, _ = digest.Write([]byte(filename))
	hash := hex.EncodeToString(digest.Sum(nil))
	return fmt.Sprintf(hash[0:5])
}

func (s *state) getRuleInfo(ruleset, rule string) error {
	tmpPluginName := "tfvetPlugin"

	s.fmt.PrintMsg(fmt.Sprintf("Collecting rule info for: %s", rule))

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: tfvetPlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			tmpPluginName: &tfvetPlugin.TfvetRulePlugin{},
		},
		Cmd: exec.Command(appcfg.RulePath(ruleset, rule)),
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: ioutil.Discard,
			Level:  0,
			Name:   "plugin",
		}),
		Stderr:           nil,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		errText := fmt.Sprintf("could not connect to rule plugin %s: %v", rule, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	raw, err := rpcClient.Dispense(tmpPluginName)
	if err != nil {
		errText := fmt.Sprintf("could not connect to rule plugin %s: %v", rule, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	plugin, ok := raw.(tfvetPlugin.RuleDefinition)
	if !ok {
		errText := fmt.Sprintf("could not convert rule plugin %s: %v", rule, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	response, err := plugin.GetRuleInfo(&proto.GetRuleInfoRequest{})
	if err != nil {
		errText := fmt.Sprintf("could not get rule info for %s: %v", rule, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	err = s.cfg.AddRule(ruleset, appcfg.Rule{
		ID:       rule,
		Name:     response.RuleInfo.Name,
		Short:    response.RuleInfo.Short,
		Long:     response.RuleInfo.Long,
		Link:     response.RuleInfo.Link,
		Severity: int(response.RuleInfo.Severity),
		Enabled:  response.RuleInfo.Default,
	})
	if err != nil {
		errText := fmt.Sprintf("could not add rule %s to config file: %v", rule, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	return nil
}

// buildRule builds the rule/plugin from srcPath and stores it in dstPath
// with the provided name.
func buildRule(srcPath, dstPath string) ([]byte, error) {
	buildArgs := []string{"build", "-o", dstPath}

	golangBinaryPath, err := exec.LookPath(golangBinaryName)
	if err != nil {
		return nil, err
	}

	// go build <args> <path_to_plugin_src_files>
	output, err := utils.ExecuteCmd(golangBinaryPath, buildArgs, nil, srcPath)
	if err != nil {
		return output, err
	}

	return output, nil
}

func runAdd(cmd *cobra.Command, args []string) error {
	repoLocation := args[0]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Adding ruleset", format)
	if err != nil {
		log.Println(err)
		return err
	}

	state.fmt.PrintMsg("Retrieving config")

	// Check that repository does not yet exist
	if state.cfg.RepositoryExists(repoLocation) {
		errText := "repository already exists; use `tfvet ruleset delete` or" +
			"`tfvet ruleset update` to manipulate already added rulesets"
		state.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	tmpDownloadPath, err := state.getRuleset(repoLocation)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDownloadPath) // Remove tmp dir in case we end early

	info, err := state.getRulesetInfo(tmpDownloadPath)
	if err != nil {
		return err
	}

	state.fmt.PrintMsg("Adding ruleset to config")
	err = state.cfg.AddRuleset(appcfg.Ruleset{
		Name:       info.Name,
		Version:    info.Version,
		Repository: repoLocation,
		Enabled:    true,
	})
	if err != nil {
		return err
	}

	err = state.moveRepo(info.Name, tmpDownloadPath)
	if err != nil {
		return err
	}

	err = state.buildRulesetRules(info.Name)
	if err != nil {
		return err
	}

	state.fmt.PrintFinalSuccess(fmt.Sprintf("Successfully added ruleset: %s v%s", info.Name, info.Version))

	return nil
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetAdd)
}
