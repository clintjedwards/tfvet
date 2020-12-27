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
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// cmdLint is a subcommand that controls the actual act of running the linter
var cmdLint = &cobra.Command{
	Use:   "lint [paths...]",
	Short: "Runs the terraform linter",
	Long: `Runs the terraform linter for all enabled rules, grabbing all terraform files in current
directory by default.

Accepts multiple paths delimited by a space.
`,
	RunE: runLint,
	Example: `$ tfvet lint
$ tfvet lint myfile.tf
$ tfvet line somefile.tf manyfilesfolder/*`,
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

		// Resolve home directory
		path, err := homedir.Expand(path)
		if err != nil {
			errText := fmt.Sprintf("could not parse path %s", path)
			s.fmt.PrintFinalError(errText)
			return nil, errors.New(errText)
		}

		// Get full path for file
		path, err = filepath.Abs(path)
		if err != nil {
			errText := fmt.Sprintf("could not parse path %s", path)
			s.fmt.PrintFinalError(errText)
			return nil, errors.New(errText)
		}

		// Check that the path exists
		_, err = os.Stat(filepath.Dir(path))
		if err != nil {
			errText := fmt.Sprintf("could not open path: %v", err)
			s.fmt.PrintFinalError(errText)
			return nil, errors.New(errText)
		}

		// Return all terraform files
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
		log.Print(err)
		return err
	}

	state, err := newState("Running Linter", format)
	if err != nil {
		log.Print(err)
		return err
	}

	// Get paths from arguments, if no arguments were given attempt to get files from current dir.
	paths := []string{}
	if len(args) == 0 {
		defaultPath, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
			return err
		}
		defaultPath = defaultPath + "/*"

		paths = []string{defaultPath}
	} else {
		paths = args
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
			state.fmt.PrintError("Skipped file", fmt.Sprintf("%s; could not open: %v", filepath.Base(file.Name()), err))
			state.fmt.PrintStandaloneMsg("")
			continue
		}
		defer file.Close()

		err = state.lintFile(file)
		if err != nil {
			state.fmt.PrintError("Skipped file", fmt.Sprintf("%s; could not lint: %v", filepath.Base(file.Name()), err))
			state.fmt.PrintStandaloneMsg("")
			continue
		}
		numFiles++
	}

	duration := time.Since(startTime)
	durationSeconds := float64(duration) / float64(time.Second)
	timePerFile := float64(duration) / float64(numFiles)

	state.fmt.PrintFinalSuccess(fmt.Sprintf("Linted %d file(s) in %.2fs (avg %.2fms/file)",
		numFiles, durationSeconds, timePerFile/float64(time.Millisecond)))

	return nil
}

// lintFile orchestrates the process of linting the given file.
func (s *state) lintFile(file *os.File) error {
	// TODO(clintjedwards): Replace with a function that doesn't suck the entire file into memory
	// or make sure file isn't too large before we absorb it.
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	_, diags := hclparse.NewParser().ParseHCL(contents, file.Name())
	if diags.HasErrors() {
		return diags
	}

	rulesets := s.cfg.Rulesets

	// For each ruleset we need to run each one of the enabled rules against the given file.
	for _, ruleset := range rulesets {
		for _, rule := range ruleset.Rules {
			if !rule.Enabled {
				continue
			}

			//<filename>::<ruleset>::<rule>
			s.fmt.PrintMsg(fmt.Sprintf("%s::%s::%s",
				filepath.Base(file.Name()), strings.ToLower(ruleset.Name), strings.ToLower(rule.Name)))

			err := s.runRule(ruleset.Name, rule, file.Name(), contents)
			if err != nil {
				s.fmt.PrintError("Rule failed", fmt.Sprintf("%s; encountered an error while running: %v",
					rule.Name, err))
				continue
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
		return fmt.Errorf("could not create rpc client: %w", err)
	}

	raw, err := rpcClient.Dispense(tmpPluginName)
	if err != nil {
		return fmt.Errorf("could not connect to rule plugin: %w", err)
	}

	plugin, ok := raw.(tfvetPlugin.RuleDefinition)
	if !ok {
		return fmt.Errorf("could not convert rule interface: %w", err)
	}

	response, err := plugin.ExecuteRule(&proto.ExecuteRuleRequest{
		HclFile: rawHCLFile,
	})
	if err != nil {
		return fmt.Errorf("could not execute linting rule: %w", err)
	}

	lintErrs := response.Errors
	for _, lintError := range lintErrs {
		line, _, err := utils.ReadLine(bytes.NewBuffer(rawHCLFile), int(lintError.Location.Start.Line))
		if err != nil {
			return fmt.Errorf("could not get line from file: %w", err)
		}

		s.fmt.PrintLintError(formatter.LintErrorDetails{
			Filepath: filepath,
			Line:     line,
			Ruleset:  ruleset,
			Rule:     rule,
			LintErr:  lintError,
		})
	}

	return nil
}

func init() {
	RootCmd.AddCommand(cmdLint)
}
