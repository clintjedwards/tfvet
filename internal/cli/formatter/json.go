package formatter

import (
	"os"

	"github.com/rs/zerolog"
)

type jsonPrinter struct {
	log *zerolog.Logger
}

func newJSONPrinter() (jsonPrinter, error) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Level(zerolog.InfoLevel)
	return jsonPrinter{
		log: &logger,
	}, nil
}
