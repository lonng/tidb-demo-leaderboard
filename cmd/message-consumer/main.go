package main

import (
	"github.com/lonng/tidb-demo-trending/config"
	"github.com/lonng/tidb-demo-trending/internal/consumer"
	"github.com/spf13/cobra"
)

func main() {
	opt := &config.ConsumerOptions{}
	cmd := cobra.Command{
		Use:          "message-consumer",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			svr := consumer.NewService(opt)
			return svr.Serve()
		},
	}

	opt.AddFlags(cmd.Flags())
	cobra.CheckErr(cmd.Execute())
}
