package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/clintjedwards/tfvet/internal/cli/models"
	"github.com/clintjedwards/tfvet/internal/plugin/proto"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
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
// It borrows(see blatantly copies) a lot from rust style errors:
// https://blog.rust-lang.org/2016/08/10/Shape-of-errors-to-come.html
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
	if f.mode == Pretty {
		// First notify the user of an error and print the short error description.
		f.PrintError("Error", strings.ToLower(details.Rule.Short))
		// Next print the filename along with the starting line and column
		f.PrintStandaloneMsg(fmt.Sprintf("  --> %s:%d:%d\n",
			details.Filepath, details.LintErr.Location.Start.Line, details.LintErr.Location.Start.Column))
		// Next pretty print the error line
		f.PrintStandaloneMsg(formatLineTable(details.Line, int(details.LintErr.Location.Start.Line)))
		f.PrintStandaloneMsg("  = additional information:\n")
		f.PrintStandaloneMsg(formatAdditionalInfo(details))
		f.PrintStandaloneMsg("\n")
		return
	}

	metadata, _ := json.Marshal(details.LintErr.Metadata)

	log.Error().
		Str("type", "linterror").
		Str("name", details.Rule.ID).
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
		{"• id:", details.Rule.ID},
		{"• name:", details.Rule.Name},
		{"• link:", details.Rule.Link},
		{"• documentation:", fmt.Sprintf("$ tfvet rule describe %s %s", details.Ruleset, details.Rule.ID)},
	}

	if details.LintErr.Suggestion != "" {
		data = append(data, []string{"• suggestion:", details.LintErr.Suggestion})
	}
	if details.LintErr.Remediation != "" {
		data = append(data,
			[]string{"• remediation:", fmt.Sprintf("`%s`", details.LintErr.Remediation)})
	}
	if len(details.LintErr.Metadata) != 0 {
		metadata, _ := json.Marshal(details.LintErr.Metadata)
		data = append(data,
			[]string{"• metadata:", fmt.Sprintf("`%s`", metadata)})
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
