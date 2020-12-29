package ruleset

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"hash/fnv"

	"github.com/clintjedwards/tfvet/internal/cli/appcfg"
	"github.com/clintjedwards/tfvet/internal/cli/models"
	tfvetPlugin "github.com/clintjedwards/tfvet/internal/plugin"
	"github.com/clintjedwards/tfvet/internal/plugin/proto"
	"github.com/clintjedwards/tfvet/internal/utils"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
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

func generateHash(filename string) string {
	digest := fnv.New32()
	_, _ = digest.Write([]byte(filename))
	hash := hex.EncodeToString(digest.Sum(nil))
	return fmt.Sprintf(hash[0:5])
}

//TODO(clintjedwards): Break down this function, it should not work this way
// it needs to be a lot smaller
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

	err = s.cfg.UpsertRule(ruleset, models.Rule{
		ID:      rule,
		Name:    response.RuleInfo.Name,
		Short:   response.RuleInfo.Short,
		Long:    response.RuleInfo.Long,
		Link:    response.RuleInfo.Link,
		Enabled: response.RuleInfo.Enabled,
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

	// Download remote repository
	state.fmt.PrintMsg(fmt.Sprintf("Retrieving %s", repoLocation))
	tmpDownloadPath := fmt.Sprintf("%s/tfvet_%s", os.TempDir(), generateHash(repoLocation))
	err = getRemoteRuleset(repoLocation, tmpDownloadPath)
	if err != nil {
		errText := fmt.Sprintf("could not download ruleset: %v", err)
		state.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}
	defer os.RemoveAll(tmpDownloadPath) // Remove tmp dir in case we end early
	state.fmt.PrintSuccess(fmt.Sprintf("Retrieved %s", repoLocation))

	// Retrieve repo into in hcl file
	info, err := getRemoteRulesetInfo(tmpDownloadPath)
	if err != nil {
		errText := fmt.Sprintf("could not get ruleset info: %v", err)
		state.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	// Ruleset verification
	state.fmt.PrintMsg("Verifying ruleset")
	err = verifyRuleset(tmpDownloadPath, info)
	if err != nil {
		errText := fmt.Sprintf("could not verify ruleset: %v", err)
		state.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}
	state.fmt.PrintSuccess("Verified ruleset")

	state.fmt.PrintMsg("Adding ruleset to config")
	err = state.cfg.AddRuleset(models.Ruleset{
		Name:       info.Name,
		Version:    info.Version,
		Repository: repoLocation,
		Enabled:    true,
	})
	if err != nil {
		errText := fmt.Sprintf("could not add ruleset: %v", err)
		state.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	err = state.moveRepo(info.Name, tmpDownloadPath)
	if err != nil {
		errText := fmt.Sprintf("could not move ruleset repository: %v", err)
		state.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	err = buildRulesetRules(state, info.Name)
	if err != nil {
		errText := fmt.Sprintf("could not build ruleset rules: %v", err)
		state.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	state.fmt.PrintFinalSuccess(fmt.Sprintf("Successfully added ruleset: %s v%s", info.Name, info.Version))
	return nil
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetAdd)
}
