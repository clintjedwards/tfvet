package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/clintjedwards/tfvet/internal/cli/models"
	"github.com/olekukonko/tablewriter"
)

// PrintAllRulesets formats and outputs all currently installed rulesets
func (f *Formatter) PrintAllRulesets(rulesets []models.Ruleset) {
	switch f.mode {
	case Pretty:
		f.PrintStandaloneMsg(formatAllRulesets(rulesets))
	case JSON:
		for _, ruleset := range rulesets {

			// We don't want to print the rules yet just the ruleset info
			// so we create a new object without the rules and then
			// json marshal that instead.
			strippedRuleset := models.Ruleset{
				Name:       ruleset.Name,
				Version:    ruleset.Version,
				Repository: ruleset.Repository,
				Enabled:    ruleset.Enabled,
			}

			jsonRuleset, _ := json.Marshal(strippedRuleset)

			f.json.log.Info().RawJSON(ruleset.Name, jsonRuleset).Msg("")
		}
	case Plain:
		for _, ruleset := range rulesets {

			// We don't want to print the rules yet just the ruleset info
			// so we create a new object without the rules and then
			// json marshal that instead.
			strippedRuleset := models.Ruleset{
				Name:       ruleset.Name,
				Version:    ruleset.Version,
				Repository: ruleset.Repository,
				Enabled:    ruleset.Enabled,
			}

			f.plain.print(strippedRuleset)
		}
	}
}

// PrintRuleset formats and outputs the information for a single ruleset
func (f *Formatter) PrintRuleset(ruleset models.Ruleset) {
	switch f.mode {
	case Pretty:
		f.PrintStandaloneMsg(formatRuleset(ruleset))
	case JSON:
		jsonRuleset, _ := json.Marshal(ruleset)

		f.json.log.Info().RawJSON(ruleset.Name, jsonRuleset).Msg("")
	case Plain:
		f.plain.print(ruleset)
	}
}

func formatAllRulesets(rulesets []models.Ruleset) string {
	headers := []string{"Name", "Version", "Repository", "Enabled", "Rules"}
	data := [][]string{}

	for _, ruleset := range rulesets {
		data = append(data, []string{
			ruleset.Name,
			ruleset.Version,
			ruleset.Repository,
			strconv.FormatBool(ruleset.Enabled),
			strconv.Itoa(len(ruleset.Rules)),
		})
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetBorder(false)
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader(headers)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}

func formatRuleset(ruleset models.Ruleset) string {

	enabledStr := ""
	if ruleset.Enabled {
		enabledStr = "enabled"
	} else {
		enabledStr = "disabled"
	}

	// example v1.0.0 :: 2 rule(s) :: enabled
	title := fmt.Sprintf("Ruleset: %s %s :: %d rule(s) :: %s\n\n",
		ruleset.Name, ruleset.Version, len(ruleset.Rules), enabledStr)

	headers := []string{"Rule", "Name", "Description", "Enabled"}
	data := [][]string{}

	for _, rule := range ruleset.Rules {
		data = append(data, []string{
			rule.ID,
			rule.Name,
			rule.Short,
			strconv.FormatBool(rule.Enabled),
		})
	}

	tableString := &strings.Builder{}
	tableString.WriteString(title)
	table := tablewriter.NewWriter(tableString)

	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader(headers)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}
