package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bloXroute-Labs/bdn-operations-relay/config"
	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
	"github.com/bloXroute-Labs/bdn-operations-relay/relay"
)

var relayCmd = &cobra.Command{
	Use:     "relay",
	Short:   "Relay operations between BDN and Atlas",
	PreRunE: config.Load,
	RunE:    runRelay,
}

func init() {
	fl := relayCmd.PersistentFlags()

	fl.String("config", "", "path to config file")
	fl.String("log-level", "info", "log level")
	fl.Int("http-port", 8080, "http port")
	fl.String("bdn.ws-url", "ws://localhost:28333/ws", "BDN WebSocket URL")
	fl.String("bdn.grpc-url", "grpc://localhost:50051", "BDN gRPC URL")
	fl.String("bdn.auth-header", "", "BDN auth header")
	fl.String("dapp-private-key", "", "DApp private key")
	fl.String("solver-private-key", "", "Solver private key")
	fl.String("dapp-address", "", "DApp address")

	err := viper.BindPFlags(fl)
	if err != nil {
		panic(err)
	}
}

func runRelay(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Read(viper.GetViper())
	if err != nil {
		return err
	}

	logger.InitLogger(cfg.LogLevel)

	return relay.Run(cfg)
}
