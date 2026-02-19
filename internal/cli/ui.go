package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/ui"
)

func newUICmd() *cobra.Command {
	var (
		port   int
		host   string
		noOpen bool
	)

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Start the web UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			if port == 0 {
				port = cfg.UI.Port
			}
			if host == "" {
				host = cfg.UI.Host
			}

			addr := fmt.Sprintf("%s:%d", host, port)
			fmt.Printf("Starting drift UI at http://%s\n", addr)

			srv := ui.NewServer(cfg, addr)
			return srv.Start(!noOpen)
		},
	}

	cmd.Flags().IntVar(&port, "port", 0, "server port (default: 4077)")
	cmd.Flags().StringVar(&host, "host", "", "server host (default: 127.0.0.1)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "don't open browser automatically")

	return cmd
}
