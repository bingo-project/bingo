package server

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"bingo/internal/pkg/log"
)

const (
	// RecommendedName defines the default project name.
	RecommendedName = "bingo"
)

var (
	// RecommendedHomeDir defines the default directory used to place all service configurations.
	RecommendedHomeDir = "." + RecommendedName

	// RecommendedEnvPrefix defines the ENV prefix used by all service.
	RecommendedEnvPrefix = strings.ToUpper(RecommendedName)
)

// LoadConfig reads in config file and ENV variables if set.
func LoadConfig(cfg string, defaultName string, data interface{}) {
	if cfg != "" {
		viper.SetConfigFile(cfg)
	} else {
		// Get User home dir
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Add `$HOME/<RecommendedHomeDir>` & `.`
		viper.AddConfigPath(filepath.Join(home, RecommendedHomeDir))
		viper.AddConfigPath(".")

		viper.SetConfigType("yaml")
		viper.SetConfigName(defaultName)
	}

	// Use config file from the flag.
	viper.AutomaticEnv()                     // read in environment variables that match.
	viper.SetEnvPrefix(RecommendedEnvPrefix) // set ENVIRONMENT variables prefix.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Errorw("Failed to read viper configuration file", "err", err)
	}

	if err := viper.Unmarshal(data); err != nil {
		log.Errorw("config unmarshal err", "err", err)
	}

	// Print using config file.
	log.Debugw("Using config file", "file", viper.ConfigFileUsed())
}
