package main

import (
	"github.com/lonng/tidb-demo-leaderboard/config"
	"github.com/lonng/tidb-demo-leaderboard/internal/leaderboard"
	"github.com/spf13/cobra"
)

func main() {
	opt := &config.ServiceOptions{}
	cmd := cobra.Command{
		Use:          "leaderboard-service",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			svr := leaderboard.NewService(opt)
			return svr.Serve()
		},
	}

	opt.AddFlags(cmd.Flags())
	cobra.CheckErr(cmd.Execute())
}
