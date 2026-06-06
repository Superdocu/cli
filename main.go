package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/Superdocu/cli/internal/cli"
)

//go:embed openapi/api.yaml
var specData []byte

var version = "dev"

func main() {
	if err := cli.Execute(specData, version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
