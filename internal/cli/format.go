package cli

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/clintjedwards/tfvet/internal/cli/models"
	"github.com/clintjedwards/tfvet/internal/plugin/proto"
	"github.com/olekukonko/tablewriter"
)

// LintErrorDetails is a harness for all the details that go into a lint error
type LintErrorDetails struct {
	Filepath string           `json:"filepath"`
	Line     string           `json:"line"`
	LintErr  *proto.RuleError `json:"lint_error"`
	Rule     models.Rule      `json:"rule"`
	Ruleset  string           `json:"ruleset"`
}

// PrintLintError formats and prints details from a lint error.
//
// It borrows(blatantly copies) from rust style errors:
// https://doc.rust-lang.org/edition-guide/rust-2018/the-compiler/improved-error-messages.html
func formatLintError(details LintErrorDetails) string {
	const lintErrorTmpl = `Error[{{.ID}}]: {{.Short}}
  --> {{.Filepath}}:{{.StartLine}}:{{.StartColumn}}
{{.LineText}}
  = additional information:
{{.Metadata}}
For more information about this error, try running ` + "`tfvet rule describe {{.Ruleset}} {{.ID}}`."

	var tpl bytes.Buffer
	t := template.Must(template.New("tmp").Parse(lintErrorTmpl))
	_ = t.Execute(&tpl, struct {
		ID          string
		Short       string
		Filepath    string
		StartLine   int
		StartColumn int
		LineText    string
		Metadata    string
		Ruleset     string
	}{
		ID:          details.Rule.ID,
		Short:       details.Rule.Short,
		Filepath:    details.Filepath,
		StartLine:   int(details.LintErr.Location.Start.Line),
		StartColumn: int(details.LintErr.Location.Start.Column),
		LineText:    formatLineTable(details.Line, int(details.LintErr.Location.Start.Line)),
		Metadata:    formatAdditionalInfo(details),
		Ruleset:     details.Ruleset,
	})

	return tpl.String()
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
				[]string{" ", fmt.Sprintf("• %s:", key), value},
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
