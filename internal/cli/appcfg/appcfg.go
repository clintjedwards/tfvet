// Package appcfg controls actions that can be performed around the app's configuration file and
// config directory.
package appcfg

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/clintjedwards/tfvet/internal/cli/models"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Appcfg represents the parsed hcl config of the main app configuration
type Appcfg struct {
	Rulesets []models.Ruleset  `hcl:"ruleset,block"`
	RepoMap  map[string]string `hcl:"repo_map,optional"` // RepoMap is a mapping of repository to ruleset.
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
func (appcfg *Appcfg) AddRuleset(rs models.Ruleset) error {
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

// UpdateRuleset updates and existing ruleset
func (appcfg *Appcfg) UpdateRuleset(rs models.Ruleset) error {
	for index, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rs.Name {
			continue
		}

		appcfg.Rulesets[index] = rs
		appcfg.RepoMap[rs.Repository] = rs.Name
		err := appcfg.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("could not find ruleset")
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

// UpsertRule adds a new rule to an already established ruleset if it does not exist. If the rule
// already exists it simply updates the rule with the newer information.
func (appcfg *Appcfg) UpsertRule(rulesetName string, newRule models.Rule) error {
	for index, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rulesetName {
			continue
		}

		for rindex, rule := range ruleset.Rules {
			if rule.ID == newRule.ID {
				// Keep user settings for updated rule
				newRule.Enabled = rule.Enabled
				appcfg.Rulesets[index].Rules[rindex] = newRule
				err := appcfg.writeConfig()
				if err != nil {
					return err
				}
				return nil
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
			if r.ID != rule {
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
			if r.ID != rule {
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
func (appcfg *Appcfg) GetRuleset(name string) (models.Ruleset, error) {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != name {
			continue
		}

		return ruleset, nil
	}

	return models.Ruleset{}, errors.New("ruleset not found")
}

// GetRule is a convenience function that returns the rule object of a given name
func (appcfg *Appcfg) GetRule(rulesetName, ruleName string) (models.Rule, error) {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rulesetName {
			continue
		}

		for _, rule := range ruleset.Rules {
			if rule.ID != ruleName {
				continue
			}

			return rule, nil
		}

		return models.Rule{}, errors.New("rule not found")
	}

	return models.Rule{}, errors.New("ruleset not found")
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
