package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/clintjedwards/tfvet/internal/cli/models"
	"github.com/clintjedwards/tfvet/internal/plugin/proto"
	"github.com/olekukonko/tablewriter"
)

// LintErrorDetails is a harness for all the details that go into a lint error
type LintErrorDetails struct {
	Filepath string
	Line     string
	Ruleset  string
	Rule     models.Rule
	LintErr  *proto.RuleError
}

// PrintLintError formats and prints details from a lint error.
//
// It borrows(blatantly copies) from rust style errors:
// https://doc.rust-lang.org/edition-guide/rust-2018/the-compiler/improved-error-messages.html
//
// Example format:
//
/*
x Error: lolwut is inherently unsafe; see link for more details
  --> ./testdata/test1.tf:17:1
    |
 17 | lolwut = "weow"
    |
	|
 = additional information:
  • rule name: resource_should_not_contain_attr_lolwut
  • link: http://lolwut.com/
  • more info: tfvet rule describe example resource_should_not_contain_attr_lolwut
  • remediation: some text here about how to fix the problem.
  • remediation: `some code to fix the issue`
*/
func (f *Formatter) PrintLintError(details LintErrorDetails) {
	switch f.mode {
	case Pretty:
		// First notify the user of an error and print the short error description.
		f.PrintError(fmt.Sprintf("Error[%s]", details.Rule.ID), strings.ToLower(details.Rule.Short))
		// Next print the filename along with the starting line and column
		f.PrintStandaloneMsg(fmt.Sprintf("  --> %s:%d:%d\n",
			details.Filepath, details.LintErr.Location.Start.Line, details.LintErr.Location.Start.Column))
		// Next pretty print the error line
		f.PrintStandaloneMsg(formatLineTable(details.Line, int(details.LintErr.Location.Start.Line)))
		f.PrintStandaloneMsg("  = additional information:\n")
		f.PrintStandaloneMsg(formatAdditionalInfo(details))
		f.PrintStandaloneMsg("\n")
		f.PrintStandaloneMsg(
			fmt.Sprintf("For more information about this error, try `tfvet rule describe %s %s`.",
				details.Ruleset, details.Rule.ID))
		f.PrintStandaloneMsg("\n")
	case JSON:
		metadata, _ := json.Marshal(details.LintErr.Metadata)

		f.json.log.Error().
			Str("type", "linterror").
			Str("id", details.Rule.ID).
			Str("name", details.Rule.Name).
			Str("short", details.Rule.Short).
			Str("link", details.Rule.Link).
			Str("line", details.Line).
			RawJSON("metadata", metadata).
			Str("suggestion", details.LintErr.Suggestion).
			Str("remediation", details.LintErr.Remediation).
			Int("start_line", int(details.LintErr.Location.Start.Line)).
			Int("start_col", int(details.LintErr.Location.Start.Column)).
			Int("end_line", int(details.LintErr.Location.End.Line)).
			Int("end_col", int(details.LintErr.Location.End.Column)).
			Msg("")
	case Plain:
		f.plain.print(details)
	}
}

// formatLineTable returns a pretty printed string of an error line
func formatLineTable(line string, lineNum int) string {
	data := [][]string{
		{"", "|", ""},
		{" " + strconv.Itoa(lineNum), "|", line},
		{"", "|", ""},
		{"", "|", ""},
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetColMinWidth(0, 3)
	table.SetRowSeparator("")
	table.SetBorder(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT})
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}

func formatAdditionalInfo(details LintErrorDetails) string {
	data := [][]string{
		{" ", "• name:", details.Rule.Name},
		{" ", "• link:", details.Rule.Link},
	}

	if details.LintErr.Suggestion != "" {
		data = append(data, []string{" ", "• suggestion:", details.LintErr.Suggestion})
	}
	if details.LintErr.Remediation != "" {
		data = append(data,
			[]string{" ", "• remediation:", fmt.Sprintf("`%s`", details.LintErr.Remediation)})
	}

	if len(details.LintErr.Metadata) != 0 {
		for key, value := range details.LintErr.Metadata {
			data = append(data,
				[]string{" ", fmt.Sprintf("• %s", key), value},
			)
		}
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}
