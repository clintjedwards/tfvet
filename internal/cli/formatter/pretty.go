package formatter

import (
	"log"
	"os"
	"time"

	"github.com/theckman/yacspin"
)

type prettyPrinter struct {
	spinner *yacspin.Spinner
	config  yacspin.Config
}

func newPrettyPrinter(suffix string) (prettyPrinter, error) {
	cfg := yacspin.Config{
		Writer:            os.Stdout,
		Frequency:         100 * time.Millisecond,
		CharSet:           yacspin.CharSets[14],
		Suffix:            " " + suffix,
		SuffixAutoColon:   true,
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "x",
		StopFailColors:    []string{"fgRed"},
	}

	pp := prettyPrinter{
		config: cfg,
	}
	pp.newSpinner()

	return pp, nil
}

// newSpinner creates a brand new pretty mode spinner from config and starts it.
func (pp *prettyPrinter) newSpinner() {
	spinner, err := yacspin.New(pp.config)
	if err != nil {
		log.Println(err)
	}
	_ = spinner.Start()
	pp.spinner = spinner
}
