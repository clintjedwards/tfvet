package formatter

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/theckman/yacspin"
)

// Formatter represents the data structure which controls how output is handled
type Formatter struct {
	isTTY   bool
	spinner *yacspin.Spinner
	config  yacspin.Config
}

// Init starts an output formatter that formats output based on given settings and environment.
func Init(suffix string) (*Formatter, error) {

	formatter := &Formatter{
		isTTY: isTTY(),
	}

	if formatter.isTTY {
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
	}

	return formatter, nil
}

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

func (f *Formatter) PrintMsg(msg string) {
	if f.isTTY {
		f.spinner.Message(msg)
		return
	}

	log.Println(msg)
}

func (f *Formatter) PrintFinalError(msg string) {
	if f.isTTY {
		f.spinner.StopFailMessage(msg)
		f.spinner.StopFail()
		return
	}

	log.Println(msg)
}

//TODO(clintjedwards): allow this package to print errors without stopping
func (f *Formatter) PrintError(suffix, msg string) {
	if f.isTTY {
		f.spinner.Suffix(fmt.Sprintf(" %s", suffix))
		f.spinner.StopFailMessage(msg)
		f.spinner.StopFail()
		newSpinner, err := newSpinner(f.config)
		if err != nil {
			log.Fatalf("could not init new spinner: %v", err)
		}

		f.spinner = newSpinner
		return
	}

	log.Println(msg)
}

func (f *Formatter) PrintFinalSuccess(msg string) {
	if f.isTTY {
		f.spinner.Suffix(fmt.Sprintf(" %s", msg))
		f.spinner.Stop()
		return
	}

	log.Println(msg)
}

func (f *Formatter) PrintSuccess(msg string) {
	if f.isTTY {
		f.spinner.Suffix(fmt.Sprintf(" %s", msg))
		f.spinner.Stop()
		newSpinner, err := newSpinner(f.config)
		if err != nil {
			log.Fatalf("could not init new spinner: %v", err)
		}

		f.spinner = newSpinner
		return
	}

	log.Println(msg)
}
