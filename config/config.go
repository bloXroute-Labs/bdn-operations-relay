package config

import (
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const envPrefix = "BDN_OPS_RELAY"

type Config struct {
	LogLevel string `mapstructure:"log-level"`
	HTTPPort int    `mapstructure:"http-port"`
	WSPort   string `mapstructure:"ws-port"`
	BDN      struct {
		WSURL   string `mapstructure:"ws-url"`
		GRPCURL string `mapstructure:"grpc-url"`
	} `mapstructure:"bdn"`
}

func Read(vip *viper.Viper) (*Config, error) {
	var cfg Config

	err := vip.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func Load(cmd *cobra.Command, _ []string) error {
	vip := viper.GetViper()

	if cmd != nil {
		err := vip.BindPFlags(cmd.Flags())
		if err != nil {
			return fmt.Errorf("failed to bind flags: %w", err)
		}
	}

	replacer := strings.NewReplacer("-", "_", ".", "_")
	vip.SetEnvKeyReplacer(replacer)
	vip.SetEnvPrefix(envPrefix)
	vip.AutomaticEnv() // read in environment variables that match

	err := readConfigFile(vip)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return nil
}

func readConfigFile(vip *viper.Viper) error {
	configFile := vip.GetString("config")
	if configFile == "" {
		slog.Warn("no config file set, set one with `--config`, " +
			"or ensure the necessary configuration are provided via ENV " +
			"variables or command-line flags. See --help")
		return nil
	} else {
		vip.SetConfigFile(configFile)
	}

	configExt := strings.TrimPrefix(path.Ext(configFile), ".")
	if !hasSupportedConfigExtension(configExt) {
		vip.SetConfigType("yaml")
	} else {
		vip.SetConfigType(configExt)
	}

	return vip.ReadInConfig()
}

func hasSupportedConfigExtension(configExt string) bool {
	for _, ext := range viper.SupportedExts {
		if configExt == ext {
			return true
		}
	}

	return false
}
