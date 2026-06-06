package main

import (
	"fmt"
	"os"

	superdocu "github.com/Superdocu/cli"
	"github.com/Superdocu/cli/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(superdocu.SpecData, version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
