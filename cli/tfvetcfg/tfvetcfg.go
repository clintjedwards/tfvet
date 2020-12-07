// Package tfvetcfg manipulates the app config ".tfvet.hcl"
package tfvetcfg

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/clintjedwards/tfvet/config"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/go-homedir"
)

var (
	// ConfigDir is the path to the main app config. All configuration and plugins are stored here.
	ConfigDir string = getConfigPath()
	// RulesetsDir is the directory that rulesets are stored.
	RulesetsDir string = ConfigDir + "/rulesets.d"
	// ConfigFilePath is the path to the main configuration file for the app
	// it stores information about overall app state and ruleset plugins.
	ConfigFilePath string = ConfigDir + "/.tfvet.hcl"
	// RepoDirName is the name of the directory that contains the ruleset's repo.
	//
	// We keep a copy of the repo on disk so that we don't have to redownload it when a user
	// decides to update.
	RepoDirName string = "repo"
	// RulesDirName is the name of the directory that contains the rules plugins within a ruleset repo.
	RulesDirName string = "rules"
)

// TfvetConfig represents the wrapping data structure for manipulation of the underlying app
// config file.
type TfvetConfig struct {
	config hclConfig
}

// hclConfig represents the parsed hcl config of the main app configuration
type hclConfig struct {
	Rulesets []Ruleset         `hcl:"ruleset,block"`
	RepoMap  map[string]string `hcl:"repo_map,optional"` // RepoMap is a mapping of repository to ruleset
}

// Ruleset represents a packaged set of rules that govern what tfvet checks for
type Ruleset struct {
	// TODO(clintjedwards): Add the ability to lock versions via some sort of syntax
	Name       string `hcl:"name,label"`
	Version    string `hcl:"version"`
	Repository string `hcl:"repository"`
	Rules      []Rule `hcl:"rule,block"`
}

type RuleSeverity int

const (
	Unknown RuleSeverity = iota
	Info
	Warning
	Error
)

// Rule represents a single lint check within a ruleset.
type Rule struct {
	FileName string `hcl:"filename,label"`
	Name     string `hcl:"name"`
	Short    string `hcl:"short"`
	Long     string `hcl:"long"`
	Severity int    `hcl:"severity"`
	Link     string `hcl:"link"`
	Enabled  bool   `hcl:"enabled"`
}

// CreateNewFile creates a new empty hcl file
func CreateNewFile() error {
	cfgFile := hclwrite.NewEmptyFile()

	f, err := os.Create(ConfigFilePath)
	if err != nil {
		return err
	}

	_, err = f.Write(cfgFile.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// GetConfig parses the on disk config file and returns its representation in golang
func GetConfig() (*TfvetConfig, error) {
	hclFile := hclConfig{}

	err := hclsimple.DecodeFile(ConfigFilePath, nil, &hclFile)
	if err != nil {
		//TODO(clintjedwards): Until we return sane errors keep this so we know why it failed
		log.Print(err)
		return nil, err
	}

	// We soft init the repo map to prevent nil entry panics
	if hclFile.RepoMap == nil {
		hclFile.RepoMap = make(map[string]string)
	}

	return &TfvetConfig{
		config: hclFile,
	}, nil
}

// RepositoryExists checks to see if a repository has already been added by a user
func (c *TfvetConfig) RepositoryExists(name string) bool {
	if _, ok := c.config.RepoMap[name]; ok {
		return true
	}

	return false
}

// AddRuleset adds a new ruleset if it doesn't exist
func (c *TfvetConfig) AddRuleset(rs Ruleset) error {
	if c.rulesetExists(rs.Name) {
		//TODO(clintjedwards): create proper custom error here
		return errors.New("ruleset already exists")
	}

	c.config.Rulesets = append(c.config.Rulesets, rs)
	c.config.RepoMap[rs.Repository] = rs.Name
	err := c.writeConfig()
	if err != nil {
		return err
	}

	return nil
}

// rulesetExists determines if a ruleset has already been added
func (c *TfvetConfig) rulesetExists(name string) bool {
	for _, ruleset := range c.config.Rulesets {
		if ruleset.Name == name {
			return true
		}
	}

	return false
}

// AddRule adds a new rule to an already established ruleset
func (c *TfvetConfig) AddRule(rulesetName string, newRule Rule) error {
	for index, ruleset := range c.config.Rulesets {
		if ruleset.Name != rulesetName {
			continue
		}

		for _, rule := range ruleset.Rules {
			if rule.FileName == newRule.FileName {
				return errors.New("rule already exists")
			}
		}

		ruleset.Rules = append(ruleset.Rules, newRule)
		c.config.Rulesets[index] = ruleset
		err := c.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("ruleset not found")
}

// GetRulesets returns the rulesets as a datastructure
func (c *TfvetConfig) GetRulesets() []Ruleset {
	return c.config.Rulesets
}

// writeConfig takes the current golang representation of the hclfile and writes it to the file
func (c *TfvetConfig) writeConfig() error {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(c.config, f.Body())
	err := ioutil.WriteFile(ConfigFilePath, f.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func getConfigPath() string {
	config, err := config.FromEnv()
	if err != nil {
		log.Fatalf("could not access config: %v", err)
	}

	absConfigPath, err := homedir.Expand(config.ConfigPath)
	if err != nil {
		log.Fatalf("could not access config: %v", err)
	}

	return absConfigPath
}
