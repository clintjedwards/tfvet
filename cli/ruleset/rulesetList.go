package ruleset

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var cmdRulesetList = &cobra.Command{
	Use:   "list [ruleset]",
	Short: "Lists a ruleset and its rules",
	Long: `Allows the listing of a ruleset and its rules.

If no argument is provided, list will display all possible rulesets and relevant details.
If a ruleset is provided, list will display the ruleset's details and rules.
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("", format)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(args) == 0 {
		state.fmt.PrintStandaloneMsg(formatAllRulesets(state.cfg.Rulesets))
		return nil
	}

	rulesetName := args[0]
	ruleset, err := state.cfg.GetRuleset(rulesetName)
	if err != nil {
		state.fmt.PrintFinalError(fmt.Sprintf("could not find ruleset %s", rulesetName))
		return err
	}
	state.fmt.PrintStandaloneMsg(formatRuleset(ruleset))
	return nil
}

func formatAllRulesets(rulesets []appcfg.Ruleset) string {
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

func formatRuleset(ruleset appcfg.Ruleset) string {

	enabledStr := ""
	if ruleset.Enabled {
		enabledStr = "enabled"
	} else {
		enabledStr = "disabled"
	}

	// example v1.0.0 :: 2 rule(s) :: enabled
	title := fmt.Sprintf("Ruleset: %s %s :: %d rule(s) :: %s\n\n",
		ruleset.Name, ruleset.Version, len(ruleset.Rules), enabledStr)

	headers := []string{"Rule", "Name", "Description", "Severity", "Enabled"}
	data := [][]string{}

	for _, rule := range ruleset.Rules {
		data = append(data, []string{
			rule.FileName,
			rule.Name,
			rule.Short,
			appcfg.SeverityToString(rule.Severity),
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

func init() {
	CmdRuleset.AddCommand(cmdRulesetList)
}
