package formatter

import (
	"fmt"
	"os"
	"strings"
)

// Formatter represents the data structure which controls how output is handled
type Formatter struct {
	mode   Mode
	plain  *plainPrinter
	pretty *prettyPrinter
	json   *jsonPrinter
}

// Mode is the type of formatting to output. Plain logging, pretty print, or json provided.
type Mode string

const (
	// Plain pretty prints json but indented and colorized.
	Plain Mode = "plain"
	// Pretty outputs text in a more humanized fashion and provides spinners for longer actions.
	Pretty Mode = "pretty"
	// JSON outputs json formatted text, mainly suitable to be read by computers.
	JSON Mode = "json"
)

// New provides a formatter with output format determined by the mode.
//
// If the formatter detects that it is within a TTY and in pretty mode, it will switch to plain mode.
// This avoids any mistaken garbaled output for none terminal destinations.
//
// The suffix parameter is only applicable to pretty mode.
// Suffix is the string of text printed right after the spinner.
// (Hence the name, despite there being other text after it.)
func New(suffix string, mode Mode) (*Formatter, error) {
	// If we can't pretty print into it just fallback to normal logging
	if mode == Pretty && !isTTY() {
		mode = Plain
	}

	formatter := &Formatter{
		mode: mode,
	}

	switch mode {
	case Plain:
		printer := newPlainPrinter()
		formatter.plain = &printer
	case JSON:
		printer, err := newJSONPrinter()
		if err != nil {
			return nil, err
		}
		formatter.json = &printer
	case Pretty:
		printer, err := newPrettyPrinter(suffix)
		if err != nil {
			return nil, err
		}
		formatter.pretty = &printer
	default:
		return nil, fmt.Errorf("invalid mode %q;"+
			" please choose an accepted format mode: [pretty, json, plain]", mode)
	}

	return formatter, nil
}

// isTTY determines if program is being run from terminal
func isTTY() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}

// PrintMsg outputs a simple message.
// In pretty mode it will display it after the suffix text.
func (f *Formatter) PrintMsg(msg string) {
	switch f.mode {
	case Pretty:
		f.pretty.spinner.Message(msg)
	case JSON:
		f.json.log.Info().Msg(msg)
	case Plain:
		f.plain.printMsg(msg)
	}
}

// PrintStandaloneMsg outputs a message unattached to the spinner or suffix text.
// In pretty mode this causes the spinner to first stop, print the message, and then immediately
// start a new spinner as to not cause the spinner suffix text to be printed.
func (f *Formatter) PrintStandaloneMsg(msg string) {
	switch f.mode {
	case Pretty:
		f.pretty.spinner.Suffix("")
		f.pretty.spinner.StopCharacter("")
		_ = f.pretty.spinner.Stop()

		if !strings.HasSuffix(msg, "\n") {
			msg = msg + "\n"
		}

		fmt.Print(msg)
		f.pretty.newSpinner()
	case JSON:
		if msg != "" {
			f.json.log.Info().Msg(msg)
		}
	case Plain:
		if msg != "" {
			f.plain.printMsg(msg)
		}
	}
}

// PrintError outputs an error.
// In pretty mode it will cause it to print the message replacing the
// current suffix and immediately start a new spinner.
func (f *Formatter) PrintError(suffix, msg string) {
	switch f.mode {
	case Pretty:
		f.pretty.spinner.Suffix(fmt.Sprintf(" %s", suffix))
		f.pretty.spinner.StopFailMessage(msg)
		_ = f.pretty.spinner.StopFail()
		f.pretty.newSpinner()
	case JSON:
		f.json.log.Error().Msg(msg)
	case Plain:
		f.plain.printErr(msg)
	}
}

// PrintFinalError prints an error; usually to end the program on.
// In pretty mode this will stop the spinner, display a red x and not start a new one.
// In all other modes this will simply print an error message.
func (f *Formatter) PrintFinalError(msg string) {
	switch f.mode {
	case Pretty:
		f.pretty.spinner.StopFailMessage(msg)
		_ = f.pretty.spinner.StopFail()
	case JSON:
		f.json.log.Error().Msg(msg)
	case Plain:
		f.plain.printErr(msg)
	}
}

// PrintSuccess outputs a success message.
// Pretty mode will cause it to print the message with a checkmark, replacing the current suffix,
// and immediately start a new spinner.
func (f *Formatter) PrintSuccess(msg string) {
	switch f.mode {
	case Pretty:
		f.pretty.spinner.Suffix(fmt.Sprintf(" %s", msg))
		_ = f.pretty.spinner.Stop()
		f.pretty.newSpinner()
	case JSON:
		f.json.log.Info().Msg(msg)
	case Plain:
		f.plain.printMsg(msg)
	}
}

// PrintFinalSuccess prints a final success message; usually to end the program on.
// In pretty mode this will stop the spinner, display a green checkmark and not start a new one.
// In all other modes this will simply print an info message.
func (f *Formatter) PrintFinalSuccess(msg string) {
	switch f.mode {
	case Pretty:
		f.pretty.spinner.Suffix(fmt.Sprintf(" %s", msg))
		_ = f.pretty.spinner.Stop()
	case JSON:
		f.json.log.Info().Msg(msg)
	case Plain:
		f.plain.printMsg(msg)
	}
}

// UpdateSuffix updates the text that comes right after the spinner.
func (f *Formatter) UpdateSuffix(text string) {
	switch f.mode {
	case Pretty:
		f.pretty.spinner.Suffix(fmt.Sprintf(" %s", text))
	case JSON:
		f.json.log.Info().Msg(text)
	case Plain:
		f.plain.printMsg(text)
	}
}
