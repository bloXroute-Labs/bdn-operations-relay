package main

import (
	"github.com/bloXroute-Labs/bdn-operations-relay/config"
	"github.com/bloXroute-Labs/bdn-operations-relay/log"
	"github.com/bloXroute-Labs/bdn-operations-relay/relay"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	fl.String("ws-port", "8081", "ws port")
	fl.String("bdn.ws-url", "ws://localhost:28333/ws", "BDN WebSocket URL")
	fl.String("bdn.grpc-url", "grpc://localhost:50051", "BDN gRPC URL")

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

	log.InitLogger(cfg.LogLevel)

	return relay.Run(cfg)
}
