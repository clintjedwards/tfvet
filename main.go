package main

import (
	"os"

	"github.com/clintjedwards/tfvet/v3/internal/cli"
)

func main() {
	err := cli.RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
