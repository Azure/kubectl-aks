// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const currentInstanceKey = "current-instance"

var config = viper.New()

func init() {
	config.SetConfigFile(path.Join(configDir(), "config.yaml"))
}

func BindEnvAndFlags(cmd *cobra.Command) error {
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return config.BindPFlags(cmd.PersistentFlags())
}

func LoadConfig() error {
	if err := config.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to read config: %w", err)
	}
	return nil
}

func LoadCurrentInstanceConfig() error {
	if err := LoadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	currentInstance := config.GetString(currentInstanceKey)
	if currentInstance != "" {
		config = config.Sub("instances." + currentInstance)
	}
	return nil
}

func GetConfig(key string) string {
	return config.GetString(key)
}

func IsConfigSet(key string) bool {
	return config.IsSet(key)
}

// configDir returns the directory where the cache file is stored.
// It will use the home directory if possible otherwise the current directory.
func configDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warn: getting home directory: %s\n", err)
		fmt.Fprintf(os.Stderr, "Warn: using current directory for config\n")
	}
	return path.Join(dir, ".kubectl-az")
}
