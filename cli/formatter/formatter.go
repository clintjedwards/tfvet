package formatter

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theckman/yacspin"
)

// Formatter represents the data structure which controls how output is handled
type Formatter struct {
	mode    Mode
	spinner *yacspin.Spinner
	config  yacspin.Config
}

// Mode is the type of formatting to output. Plain logging, pretty print, or json provided.
type Mode string

const (
	// Plain outputs normal golang plaintext logging.
	Plain Mode = "plain"
	// Pretty outputs text in a more humanized fashion and provides spinners for longer actions.
	Pretty Mode = "pretty"
	// JSON outputs json formatted text, mainly suitable to be read by computers.
	JSON Mode = "json"
)

// New provides a new formatter with output format determined by the mode.
//
// If the formatter detects that it is within a TTY in pretty mode it will switch to plain mode.
// This avoids any mistaken garbaled output for none terminal destinations.
//
// The suffix parameter is only applicable to pretty mode; suffix is the string of text printed
// right after the spinner. Hence the name, despite there being other text after it.
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
		newPlainLogger()
	case JSON:
		newJSONLogger()
	case Pretty:
		cfg := yacspin.Config{
			Writer:            os.Stderr,
			Frequency:         100 * time.Millisecond,
			CharSet:           yacspin.CharSets[14],
			Suffix:            " " + suffix,
			SuffixAutoColon:   true,
			StopCharacter:     "✓",
			StopColors:        []string{"fgGreen"},
			StopFailCharacter: "x",
			StopFailColors:    []string{"fgRed"},
		}

		spinner, err := newSpinner(cfg)
		if err != nil {
			return nil, err
		}

		formatter.spinner = spinner
		formatter.config = cfg

	default:
		return nil, errors.New("could not establish formatter; mode not found")
	}

	return formatter, nil
}

func newPlainLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func newJSONLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// newSpinner creates a brand new pretty mode spinner from config and starts it.
func newSpinner(cfg yacspin.Config) (*yacspin.Spinner, error) {
	spinner, err := yacspin.New(cfg)
	if err != nil {
		return nil, err
	}
	spinner.Start()
	return spinner, nil
}

// isTTY determines if program is being run from terminal
func isTTY() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}

// PrintMsg displays a simple message. In pretty mode it will display it after the suffix text.
func (f *Formatter) PrintMsg(msg string) {
	if f.mode == Pretty {
		f.spinner.Message(msg)
		return
	}

	log.Info().Msg(msg)
}

// PrintStandaloneMsg displays a message unattached to the spinner.
// In pretty mode this causes the spinner to first stop, print the message, and then immediately
// start a new spinner.
func (f *Formatter) PrintStandaloneMsg(msg string) {
	if f.mode == Pretty {
		f.spinner.Suffix("")
		f.spinner.StopCharacter("")
		f.spinner.Stop()

		if !strings.HasSuffix(msg, "\n") {
			msg = msg + "\n"
		}

		fmt.Print(msg)
		newSpinner, err := newSpinner(f.config)
		if err != nil {
			log.Fatal().Err(err).Msg("could not init new spinner")
		}

		f.spinner = newSpinner
		return
	}

	log.Info().Msg(msg)
}

// PrintError displays an error. Pretty mode it will cause it to print the message replacing the
// current suffix and immediately start a new spinner.
func (f *Formatter) PrintError(suffix, msg string) {
	if f.mode == Pretty {
		f.spinner.Suffix(fmt.Sprintf(" %s", suffix))
		f.spinner.StopFailMessage(msg)
		f.spinner.StopFail()
		newSpinner, err := newSpinner(f.config)
		if err != nil {
			log.Fatal().Err(err).Msg("could not init new spinner")
		}

		f.spinner = newSpinner
		return
	}

	log.Error().Msg(msg)
}

// PrintFinalError prints an error; usually to end the program on.
// In pretty mode this will stop the spinner, display a red x and not start a new one.
// In all other modes this will simply print an error message.
func (f *Formatter) PrintFinalError(msg string) {
	if f.mode == Pretty {
		f.spinner.StopFailMessage(msg)
		f.spinner.StopFail()
		return
	}

	log.Error().Msg(msg)
}

// PrintSuccess prints a success message. Pretty mode will cause it to print the message
// replacing the current suffix and immediately start a new spinner.
func (f *Formatter) PrintSuccess(msg string) {
	if f.mode == Pretty {
		f.spinner.Suffix(fmt.Sprintf(" %s", msg))
		f.spinner.Stop()
		newSpinner, err := newSpinner(f.config)
		if err != nil {
			log.Fatal().Err(err).Msg("could not init new spinner")
		}

		f.spinner = newSpinner
		return
	}

	log.Info().Msg(msg)
}

// PrintFinalSuccess prints a final success message; usually to end the program on.
// In pretty mode this will stop the spinner, display a green checkmark and not start a new one.
// In all other modes this will simply print an info message.
func (f *Formatter) PrintFinalSuccess(msg string) {
	if f.mode == Pretty {
		f.spinner.Suffix(fmt.Sprintf(" %s", msg))
		f.spinner.Stop()
		return
	}

	log.Info().Msg(msg)
}
