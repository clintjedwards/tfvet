// Package appcfg controls actions that can be performed around the app's configuration file and
// config directory.
package appcfg

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

//TODO(clintjedwards): Create and return custom errors

// Appcfg represents the parsed hcl config of the main app configuration
type Appcfg struct {
	Rulesets []Ruleset         `hcl:"ruleset,block"`
	RepoMap  map[string]string `hcl:"repo_map,optional"` // RepoMap is a mapping of repository to ruleset
}

// Ruleset represents a packaged set of rules that govern what tfvet checks for
type Ruleset struct {
	Name       string `hcl:"name,label"`
	Version    string `hcl:"version"`
	Repository string `hcl:"repository"`
	Enabled    bool   `hcl:"enabled"`
	Rules      []Rule `hcl:"rule,block"`
}

// RuleSeverity is used to convey how serious the offending error is. This is passed in the output
// of tfvet linting errors so that downstream tools can choose whether to surface them or not.
type RuleSeverity int

const (
	// Unknown is used when the severity is not supplied.
	Unknown RuleSeverity = iota
	// Info is used to convey general information about something which might not be immediately fixable.
	Info
	// Warning is used to convey things to watch out for that can potentially turn bad.
	Warning
	// Error is used to convey things that should change immediately.
	Error
)

// SeverityToString converts a severity into a string
func SeverityToString(severity int) string {
	switch severity {
	case int(Unknown):
		return "Unknown"
	case int(Info):
		return "Info"
	case int(Warning):
		return "Warning"
	case int(Error):
		return "Error"
	default:
		return "Unknown"
	}
}

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

// CreateNewFile creates a new empty config file
func CreateNewFile() error {
	cfgFile := hclwrite.NewEmptyFile()

	f, err := os.Create(ConfigFilePath())
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
func GetConfig() (*Appcfg, error) {
	hclFile := &Appcfg{}

	err := hclsimple.DecodeFile(ConfigFilePath(), nil, hclFile)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	// We soft init the repo map to prevent nil entry panics
	if hclFile.RepoMap == nil {
		hclFile.RepoMap = make(map[string]string)
	}

	return hclFile, nil
}

// RepositoryExists checks to see if the config already has an entry for the repository in the config.
func (appcfg *Appcfg) RepositoryExists(name string) bool {
	if _, ok := appcfg.RepoMap[name]; ok {
		return true
	}

	return false
}

// AddRuleset adds a new ruleset if it doesn't exist.
func (appcfg *Appcfg) AddRuleset(rs Ruleset) error {
	if appcfg.rulesetExists(rs.Name) {
		return errors.New("ruleset already exists")
	}

	appcfg.Rulesets = append(appcfg.Rulesets, rs)
	appcfg.RepoMap[rs.Repository] = rs.Name
	err := appcfg.writeConfig()
	if err != nil {
		return err
	}

	return nil
}

// rulesetExists determines if a ruleset has already been added.
func (appcfg *Appcfg) rulesetExists(name string) bool {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name == name {
			return true
		}
	}

	return false
}

// AddRule adds a new rule to an already established ruleset.
func (appcfg *Appcfg) AddRule(rulesetName string, newRule Rule) error {
	for index, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rulesetName {
			continue
		}

		for _, rule := range ruleset.Rules {
			if rule.FileName == newRule.FileName {
				return errors.New("rule already exists")
			}
		}

		ruleset.Rules = append(ruleset.Rules, newRule)
		appcfg.Rulesets[index] = ruleset
		err := appcfg.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("ruleset not found")
}

// DisableRuleset changes the enabled attribute on a ruleset to false
func (appcfg *Appcfg) DisableRuleset(name string) error {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != name {
			continue
		}

		ruleset.Enabled = false
		err := appcfg.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("ruleset not found")
}

//EnableRuleset changes the enabled attribute on a ruleset to true
func (appcfg *Appcfg) EnableRuleset(name string) error {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != name {
			continue
		}

		ruleset.Enabled = true
		err := appcfg.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("ruleset not found")
}

// DisableRule changes the enabled attribute on a rule to false
func (appcfg *Appcfg) DisableRule(ruleset, rule string) error {
	for _, rs := range appcfg.Rulesets {
		if rs.Name != ruleset {
			continue
		}

		for index, r := range rs.Rules {
			if r.FileName != rule {
				continue
			}

			rs.Rules[index].Enabled = false
			err := appcfg.writeConfig()
			if err != nil {
				return err
			}

			return nil
		}
		return errors.New("rule not found")
	}

	return errors.New("ruleset not found")
}

// EnableRule changes the enabled attribute on a rule to true
func (appcfg *Appcfg) EnableRule(ruleset, rule string) error {
	for _, rs := range appcfg.Rulesets {
		if rs.Name != ruleset {
			continue
		}

		for index, r := range rs.Rules {
			if r.FileName != rule {
				continue
			}

			rs.Rules[index].Enabled = true
			err := appcfg.writeConfig()
			if err != nil {
				return err
			}

			return nil
		}
		return errors.New("rule not found")
	}

	return errors.New("ruleset not found")
}

// GetRuleset is a convenience function that returns the ruleset object of a given name
func (appcfg *Appcfg) GetRuleset(name string) (Ruleset, error) {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != name {
			continue
		}

		return ruleset, nil
	}

	return Ruleset{}, errors.New("ruleset not found")
}

// GetRule is a convenience function that returns the rule object of a given name
func (appcfg *Appcfg) GetRule(rulesetName, ruleName string) (Rule, error) {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rulesetName {
			continue
		}

		for _, rule := range ruleset.Rules {
			if rule.FileName != ruleName {
				continue
			}

			return rule, nil
		}

		return Rule{}, errors.New("rule not found")
	}

	return Rule{}, errors.New("ruleset not found")
}

// writeConfig takes the current representation of config and writes it to the file.
func (appcfg *Appcfg) writeConfig() error {
	f := hclwrite.NewEmptyFile()

	gohcl.EncodeIntoBody(appcfg, f.Body())

	err := ioutil.WriteFile(ConfigFilePath(), f.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
