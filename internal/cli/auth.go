package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Superdocu/cli/internal/client"
	"github.com/Superdocu/cli/internal/config"
	"github.com/Superdocu/cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate and store an API token",
		Long: "Validates the token against GET /api/v2/me and stores it in the OS keychain\n" +
			"(falling back to a 0600 config file). Provide it via --token, SUPERDOCU_TOKEN,\n" +
			"a pipe, or the interactive prompt. Tokens are issued from the Superdocu admin UI.",
		RunE: func(_ *cobra.Command, _ []string) error {
			host := api.Host
			token := firstNonEmpty(g.token, os.Getenv("SUPERDOCU_TOKEN"))
			if token == "" {
				t, err := promptToken()
				if err != nil {
					return err
				}
				token = t
			}
			if token == "" {
				return fmt.Errorf("no token provided")
			}

			resp, err := client.New(host, token).Do(client.Request{Method: "GET", Path: "/api/v2/me"})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode >= 400 {
				return client.ParseError(resp, body)
			}

			if err := config.SaveToken(host, token); err != nil {
				return err
			}
			if cfg, _ := config.Load(); cfg != nil {
				cfg.Host = host
				_ = cfg.Save()
			}
			fmt.Fprintln(os.Stderr, "Logged in to "+host)
			return output.PrintJSON(os.Stdout, body, g.raw)
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored API token",
		RunE: func(_ *cobra.Command, _ []string) error {
			config.DeleteToken(api.Host)
			fmt.Fprintln(os.Stderr, "Logged out of "+api.Host)
			return nil
		},
	}
}

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the current identity (GET /api/v2/me)",
		RunE: func(_ *cobra.Command, _ []string) error {
			if api.Token == "" {
				return fmt.Errorf("not authenticated — run `superdocu login`")
			}
			resp, err := api.Do(client.Request{Method: "GET", Path: "/api/v2/me"})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode >= 400 {
				return client.ParseError(resp, body)
			}
			return output.PrintJSON(os.Stdout, body, g.raw)
		},
	}
}

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show the resolved configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("host:  " + api.Host)
			fmt.Println("token: " + maskToken(api.Token))
			fmt.Println("file:  " + config.Path())
			return nil
		},
	}
}

func promptToken() (string, error) {
	fmt.Fprint(os.Stderr, "API token: ")
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		return strings.TrimSpace(string(b)), err
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func maskToken(t string) string {
	if t == "" {
		return "(none)"
	}
	if len(t) <= 8 {
		return "********"
	}
	return t[:4] + "…" + t[len(t)-4:]
}
