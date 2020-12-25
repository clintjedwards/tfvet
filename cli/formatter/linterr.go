package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/clintjedwards/tfvet/plugin/proto"
	"github.com/olekukonko/tablewriter"
)

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
//TODO(clintjedwards): Remove all params and include them into a harness instead
func (f *Formatter) PrintLintError(filepath, line, ruleset string, rule appcfg.Rule, lintErr *proto.RuleError) {
	// First notify the user of an error and print the short error description.
	f.PrintError("Error", strings.ToLower(rule.Short))
	// Next print the filename along with the starting line and column
	f.PrintStandaloneMsg(fmt.Sprintf("  --> %s:%d:%d\n",
		filepath, lintErr.Location.Start.Line, lintErr.Location.Start.Column))
	// Next pretty print the error line
	f.PrintStandaloneMsg(formatLineTable(line, int(lintErr.Location.Start.Line)))
	f.PrintStandaloneMsg("  = additional information:\n")
	f.PrintStandaloneMsg(formatAdditionalInfo(ruleset, rule))
	f.PrintStandaloneMsg("\n")
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
