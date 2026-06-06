// Package cli wires the parsed spec into a Cobra command tree.
package cli

import (
	"fmt"
	"os"

	"github.com/Superdocu/cli/internal/client"
	"github.com/Superdocu/cli/internal/config"
	"github.com/Superdocu/cli/internal/spec"
	"github.com/spf13/cobra"
)

type globals struct {
	host        string
	token       string
	raw         bool
	output      string
	idempotency string
	verbose     bool
	queryKV     []string
}

var (
	g       = &globals{}
	apiSpec *spec.Spec
	api     *client.Client
)

// Execute parses the embedded spec and runs the root command.
func Execute(specData []byte, version string) error {
	s, err := spec.Load(specData)
	if err != nil {
		return fmt.Errorf("load embedded spec: %w", err)
	}
	apiSpec = s

	root := &cobra.Command{
		Use:           "superdocu",
		Short:         "Superdocu API v2 command-line client",
		Long:          fmt.Sprintf("%s — commands are generated from the embedded OpenAPI spec (v%s).", s.Title, s.Version),
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := root.PersistentFlags()
	pf.StringVar(&g.host, "host", "", "API host (env SUPERDOCU_HOST; default "+s.DefaultHost+")")
	pf.StringVar(&g.token, "token", "", "Bearer token (env SUPERDOCU_TOKEN)")
	pf.BoolVar(&g.raw, "raw", false, "Output raw JSON without pretty-printing")
	pf.StringVarP(&g.output, "output", "o", "", "Write a binary/CSV response body to this file")
	pf.StringVar(&g.idempotency, "idempotency-key", "", "Idempotency-Key header (UUID v4) for mutations")
	pf.BoolVarP(&g.verbose, "verbose", "v", false, "Print response headers to stderr")
	pf.StringArrayVar(&g.queryKV, "query", nil, "Extra query parameter key=value (repeatable)")

	root.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		host := firstNonEmpty(g.host, os.Getenv("SUPERDOCU_HOST"))
		if host == "" {
			if cfg, _ := config.Load(); cfg != nil {
				host = cfg.Host
			}
		}
		if host == "" {
			host = s.DefaultHost
		}
		token := firstNonEmpty(g.token, os.Getenv("SUPERDOCU_TOKEN"))
		if token == "" {
			token = config.LoadToken(host)
		}
		api = client.New(host, token)
		return nil
	}

	root.AddCommand(newLoginCmd(), newLogoutCmd(), newWhoamiCmd(), newConfigCmd())
	addAPICommands(root, s)
	return root.Execute()
}
