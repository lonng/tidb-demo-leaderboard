package main

import (
	"github.com/lonng/tidb-demo-trending/config"
	"github.com/lonng/tidb-demo-trending/internal/trending"
	"github.com/spf13/cobra"
)

func main() {
	opt := &config.ServiceOptions{}
	cmd := cobra.Command{
		Use:          "trending-service",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			svr := trending.NewService(opt)
			return svr.Serve()
		},
	}

	opt.AddFlags(cmd.Flags())
	cobra.CheckErr(cmd.Execute())
}
