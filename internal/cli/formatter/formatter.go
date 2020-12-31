package formatter

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	golog "log"

	"github.com/clintjedwards/tfvet/internal/config"
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

// New provides a formatter with output format determined by the mode.
//
// If the formatter detects that it is within a TTY and in pretty mode, it will switch to plain mode.
// This avoids any mistaken garbaled output for none terminal destinations.
//
// The suffix parameter is only applicable to pretty mode.
// Suffix is the string of text printed right after the spinner.
// (Hence the name, despite there being other text after it.)
func New(suffix string, mode Mode) (*Formatter, error) {

	config, err := config.FromEnv()
	if err != nil {
		return nil, fmt.Errorf("could not access config: %w", err)
	}

	// If we can't pretty print into it just fallback to normal logging
	if mode == Pretty && !isTTY() {
		mode = Plain
	}

	formatter := &Formatter{
		mode: mode,
	}

	switch mode {
	case Plain:
		newPlainLogger(parseLogLevel(config.LogLevel))
	case JSON:
		newJSONLogger(parseLogLevel(config.LogLevel))
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

func parseLogLevel(loglevel string) zerolog.Level {
	switch loglevel {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		golog.Printf("loglevel %s not recognized; defaulting to debug", loglevel)
		return zerolog.DebugLevel
	}
}

func newPlainLogger(level zerolog.Level) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func newJSONLogger(level zerolog.Level) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(level)
}

// newSpinner creates a brand new pretty mode spinner from config and starts it.
func newSpinner(cfg yacspin.Config) (*yacspin.Spinner, error) {
	spinner, err := yacspin.New(cfg)
	if err != nil {
		return nil, err
	}
	_ = spinner.Start()
	return spinner, nil
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
	if f.mode == Pretty {
		f.spinner.Message(msg)
		return
	}

	log.Info().Msg(msg)
}

// PrintStandaloneMsg outputs a message unattached to the spinner or suffix text.
// In pretty mode this causes the spinner to first stop, print the message, and then immediately
// start a new spinner as to not cause the spinner suffix text to be printed.
func (f *Formatter) PrintStandaloneMsg(msg string) {
	if f.mode == Pretty {
		f.spinner.Suffix("")
		f.spinner.StopCharacter("")
		_ = f.spinner.Stop()

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

// PrintError outputs an error.
// In pretty mode it will cause it to print the message replacing the
// current suffix and immediately start a new spinner.
func (f *Formatter) PrintError(suffix, msg string) {
	if f.mode == Pretty {
		f.spinner.Suffix(fmt.Sprintf(" %s", suffix))
		f.spinner.StopFailMessage(msg)
		_ = f.spinner.StopFail()
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
		_ = f.spinner.StopFail()
		return
	}

	log.Error().Msg(msg)
}

// PrintSuccess outputs a success message.
// Pretty mode will cause it to print the message with a checkmark, replacing the current suffix,
// and immediately start a new spinner.
func (f *Formatter) PrintSuccess(msg string) {
	if f.mode == Pretty {
		f.spinner.Suffix(fmt.Sprintf(" %s", msg))
		_ = f.spinner.Stop()
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
		_ = f.spinner.Stop()
		return
	}

	log.Info().Msg(msg)
}

// UpdateSuffix updates the text that comes right after the spinner.
func (f *Formatter) UpdateSuffix(text string) {
	if f.mode == Pretty {
		f.spinner.Suffix(fmt.Sprintf(" %s", text))
		return
	}

	log.Info().Msg(text)
}
