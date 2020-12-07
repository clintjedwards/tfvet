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
	"strconv"
	"strings"
	"time"

	"github.com/clintjedwards/tfvet/cli/formatter"
	"github.com/clintjedwards/tfvet/cli/tfvetcfg"
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
}

// state contains a bunch of useful state information for the add cli function. This is mostly
// just for convenience.
type state struct {
	fmt *formatter.Formatter
	cfg *tfvetcfg.TfvetConfig
}

// newState returns a new state object with the fmt initialized
func newState(initialFmtMsg string) (*state, error) {
	clifmt, err := formatter.Init(initialFmtMsg)
	if err != nil {
		return nil, err
	}

	cfg, err := tfvetcfg.GetConfig()
	if err != nil {
		//TODO(clintjedwards): Make sure to pass an actual error here we can check against
		// to explicitly say that config file does not exist
		errText := fmt.Sprintf("config file `%s` does not exist."+
			" Run `tfvet init` to create.", tfvetcfg.ConfigFilePath)
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
		// we trim the star if included so that we can check if the path exists.
		baseDir := strings.TrimSuffix(path, "*")
		_, err := os.Stat(baseDir)
		if err != nil {
			errText := fmt.Sprintf("could not open path: %v", err)
			s.fmt.PrintFinalError(errText)
			return nil, errors.New(errText)
		}

		globFiles, err := filepath.Glob(path)
		if err != nil {
			fmt.Println(globFiles)
			errText := fmt.Sprintf("could not parse path %s", path)
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

	//TODO(clintjedwards): Make sure all of these report errors correctly
	state, err := newState("Running Linter")
	if err != nil {
		return err
	}

	//state.fmt.PrintMsg("testing")
	paths, err := cmd.Flags().GetStringSlice("path")
	if err != nil {
		return err
	}

	files, err := state.getTerraformFiles(paths)
	if err != nil {
		return err
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

	rulesets := s.cfg.GetRulesets()

	for _, ruleset := range rulesets {
		s.fmt.PrintMsg(fmt.Sprintf("Linting %s: Ruleset %s", file.Name(), ruleset.Name))

		for _, rule := range ruleset.Rules {
			s.fmt.PrintMsg(fmt.Sprintf("Linting %s: Ruleset %s - Rule %s",
				file.Name(), ruleset.Name, rule.Name))

			err := s.runRule(ruleset.Name, rule, file.Name(), contents)
			if err != nil {
				//TODO(clintjedwards): proper errors
				return err
			}
		}
	}

	return nil
}

func (s *state) runRule(ruleset string, rule tfvetcfg.Rule, filepath string, rawHCLFile []byte) error {
	rulesPath := tfvetcfg.RulesetsDir + "/" + ruleset
	finalPath := rulesPath + "/" + rule.FileName
	tmpPluginName := "tfvetPlugin"

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: tfvetPlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			tmpPluginName: &tfvetPlugin.TfvetRulePlugin{},
		},
		Cmd: exec.Command(finalPath),
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
		_ = lintError
		s.fmt.PrintError("Error", strings.ToLower(rule.Short))
		//TODO(clintjedwards): Wrap these with the formatter
		///https://blog.rust-lang.org/2016/08/10/Shape-of-errors-to-come.html
		fmt.Printf(" --> %s:%d:%d\n", filepath, lintError.Location.Start.Line, lintError.Location.Start.Column)
		line, lineNum, err := utils.ReadLine(bytes.NewBuffer(rawHCLFile), int(lintError.Location.Start.Line))
		if err != nil {
			panic(err)
		}
		//TODO(clintjedwards): format this as a table
		fmt.Printf("%s|\n", strings.Join(spacer(lineNum), ""))
		fmt.Printf("%d | %s\n", lineNum, line)
		fmt.Printf("%s|\n", strings.Join(spacer(lineNum), ""))
		fmt.Printf("%s|\n", strings.Join(spacer(lineNum), ""))
		fmt.Printf(" = additional information:\n")
		fmt.Printf("   • rule: %s\n", rule.FileName)
		fmt.Printf("   • link: %s\n", rule.Link)
		fmt.Printf("   • long: tfvet rule describe %s %s\n", ruleset, rule.FileName)
		fmt.Println()
	}

	return err
}

//TODO(clintjedwards): get rid of this in favor if just drawing it as a simple table
func spacer(digit int) []string {
	spaces := []string{}

	length := len(strconv.Itoa(digit))

	for i := 1; i < length+2; i++ {
		spaces = append(spaces, " ")
	}

	return spaces
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
