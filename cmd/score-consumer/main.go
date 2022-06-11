package main

import (
	"github.com/lonng/tidb-demo-leaderboard/config"
	"github.com/lonng/tidb-demo-leaderboard/internal/consumer"
	"github.com/spf13/cobra"
)

func main() {
	opt := &config.ConsumerOptions{}
	cmd := cobra.Command{
		Use:          "score-consumer",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			svr := consumer.NewService(opt)
			return svr.Serve()
		},
	}

	opt.AddFlags(cmd.Flags())
	cobra.CheckErr(cmd.Execute())
}
