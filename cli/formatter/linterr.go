package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/clintjedwards/tfvet/plugin/proto"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
)

// LintErrorDetails is a harness for all the details that go into a lint error
type LintErrorDetails struct {
	Filepath string
	Line     string
	Ruleset  string
	Rule     appcfg.Rule
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
  • rule: resource_should_not_contain_attr_lolwut
  • link: http://lolwut.com/
  • long: tfvet rule describe example resource_should_not_contain_attr_lolwut
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
		f.PrintStandaloneMsg(formatAdditionalInfo(details.Ruleset, details.Rule))
		f.PrintStandaloneMsg("\n")
		return
	}

	log.Error().
		Str("type", "linterror").
		Str("name", details.Rule.FileName).
		Str("short", details.Rule.Short).
		Str("link", details.Rule.Link).
		Str("line", details.Line).
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
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT})
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}

func formatAdditionalInfo(ruleset string, rule appcfg.Rule) string {
	moreInfoCmd := fmt.Sprintf("tfvet rule describe %s %s", ruleset, rule.FileName)

	data := [][]string{
		{"• rule name:", rule.FileName},
		{"• link:", rule.Link},
		{"• more info:", fmt.Sprintf("`%s`", moreInfoCmd)},
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
