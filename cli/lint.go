package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/clintjedwards/tfvet/cli/formatter"
	tfvetPlugin "github.com/clintjedwards/tfvet/plugin"
	"github.com/clintjedwards/tfvet/plugin/proto"
	"github.com/clintjedwards/tfvet/utils"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"
)

// cmdLint is a subcommand that controls the actual act of running the linter
var cmdLint = &cobra.Command{
	Use:   "lint",
	Short: "Runs the terraform linter",
	Long: `Runs the terraform linter for all enabled rules, grabbing all terraform files in current
directory by default.
`,
	RunE: runLint,
	Args: cobra.MaximumNArgs(0),
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

// getTerraformFiles returns the paths of all terraform files within the path given.
// If the given path is not a directory, but a terraform file instead, it will return a list with
// only that file included.
func (s *state) getTerraformFiles(paths []string) ([]string, error) {

	tfFiles := []string{}

	for _, path := range paths {

		//TODO(clintjedwards): You don't have to always glob with a *
		// instead we should check for at least two slashes here
		// and take the last slash off and check the contents of the first slash
		// we trim the star if included so that we can check if the path exists.
		baseDir := strings.TrimSuffix(path, "*")
		_, err := os.Stat(baseDir)
		if err != nil {
			errText := fmt.Sprintf("could not open path: %v", err)
			s.fmt.PrintFinalError(errText)
			return nil, errors.New(errText)
		}

		// Get full path for file
		path, err := filepath.Abs(path)
		if err != nil {
			errText := fmt.Sprintf("could not parse path %s", path)
			s.fmt.PrintFinalError(errText)
			return nil, errors.New(errText)
		}

		globFiles, err := filepath.Glob(path)
		if err != nil {
			errText := fmt.Sprintf("could match on glob pattern %s", path)
			s.fmt.PrintFinalError(errText)
			return nil, errors.New(errText)
		}

		for _, file := range globFiles {
			if strings.HasSuffix(file, ".tf") {
				tfFiles = append(tfFiles, file)
			}
		}
	}

	return tfFiles, nil
}

func runLint(cmd *cobra.Command, args []string) error {

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Running Linter", format)
	if err != nil {
		return err
	}

	paths, err := cmd.Flags().GetStringSlice("path")
	if err != nil {
		return err
	}

	files, err := state.getTerraformFiles(paths)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		state.fmt.PrintFinalError("No terraform files found")
		return errors.New("No terraform files found")
	}

	startTime := time.Now()
	numFiles := 0

	for _, file := range files {
		file, err := os.Open(file)
		if err != nil {
			return err
		}
		defer file.Close()

		//TODO(clintjedwards): check that the file is valid hcl and throw an error if its not
		// we can probably just skip that file in normal circumstances but make sure to return
		// a bad error code.
		// parser := hclparse.NewParser()
		// _, diags := parser.ParseHCL(contents, file)
		// if diags.HasErrors() {
		// 	state.fmt.PrintFinalError(fmt.Sprintf("%v", diags.Errs()))
		// 	return err
		// }

		err = state.lintFile(file)
		if err != nil {
			return err
		}
		numFiles++
	}

	duration := time.Since(startTime)
	durationSeconds := float64(duration) / float64(time.Second)
	timePerFile := float64(duration) / float64(numFiles)

	state.fmt.PrintSuccess(fmt.Sprintf("Linted %d file(s) in %.2fs (average %.2fms/file)",
		numFiles, durationSeconds, timePerFile/float64(time.Millisecond)))

	state.fmt.PrintSuccess("Finished lint")

	return nil
}

func (s *state) lintFile(file *os.File) error {
	s.fmt.PrintMsg(fmt.Sprintf("Linting %s", file.Name()))

	// TODO(clintjedwards): Replace with a function that doesn't suck the entire file into memory
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	rulesets := s.cfg.Rulesets

	for _, ruleset := range rulesets {
		s.fmt.PrintMsg(fmt.Sprintf("Linting %s: Ruleset %s", file.Name(), ruleset.Name))

		for _, rule := range ruleset.Rules {
			s.fmt.PrintMsg(fmt.Sprintf("Linting %s: Ruleset %s - Rule %s",
				file.Name(), ruleset.Name, strings.ToLower(rule.Name)))

			err := s.runRule(ruleset.Name, rule, file.Name(), contents)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *state) runRule(ruleset string, rule appcfg.Rule, filepath string, rawHCLFile []byte) error {
	tmpPluginName := "tfvetPlugin"

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: tfvetPlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			tmpPluginName: &tfvetPlugin.TfvetRulePlugin{},
		},
		Cmd: exec.Command(appcfg.RulePath(ruleset, rule.FileName)),
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
		errText := fmt.Sprintf("could not connect to rule plugin %s: %v", rule.FileName, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	raw, err := rpcClient.Dispense(tmpPluginName)
	if err != nil {
		errText := fmt.Sprintf("could not connect to rule plugin %s: %v", rule.FileName, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	plugin, ok := raw.(tfvetPlugin.RuleDefinition)
	if !ok {
		errText := fmt.Sprintf("could not convert rule plugin %s: %v", rule.FileName, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	//TODO(clintjedwards): If rule has a proper suggestion it should supply that as well
	response, err := plugin.ExecuteRule(&proto.ExecuteRuleRequest{
		HclFile: rawHCLFile,
	})
	if err != nil {
		errText := fmt.Sprintf("could not run lint rule %s: %v", rule.FileName, err)
		s.fmt.PrintFinalError(errText)
		return errors.New(errText)
	}

	lintErrs := response.Errors
	for _, lintError := range lintErrs {
		line, _, err := utils.ReadLine(bytes.NewBuffer(rawHCLFile), int(lintError.Location.Start.Line))
		if err != nil {
			return errors.New("could not get line from file")
		}

		s.fmt.PrintLintError(formatter.LintErrorDetails{
			Filepath: filepath,
			Line:     line,
			Ruleset:  ruleset,
			Rule:     rule,
			LintErr:  lintError,
		})
	}

	return err
}

func init() {
	defaultPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}

	defaultPath = defaultPath + "/*"

	cmdLint.Flags().StringSliceP("path", "p", []string{defaultPath},
		"Path to terraform files to lint; can point to a directory or a single file; can be used multiple times")

	RootCmd.AddCommand(cmdLint)
}
