package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bingo-project/bingo/internal/pkg/log"
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
func LoadConfig(cfg string, defaultName string, data interface{}, onChange func()) {
	if cfg != "" {
		viper.SetConfigFile(cfg)
	} else {
		// Get User home dir
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Add `$HOME/<RecommendedHomeDir>` & `.`
		viper.AddConfigPath(filepath.Join("/etc", RecommendedName))
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
		os.Exit(1)
	}

	if err := viper.Unmarshal(data); err != nil {
		log.Errorw("config unmarshal err", "err", err)
		os.Exit(1)
	}

	// Print using config file.
	log.Debugw("Using config file", "file", viper.ConfigFileUsed())

	// Watch config file
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infow("config file changed")
		if err := viper.Unmarshal(data); err != nil {
			log.Errorw("config unmarshal err", "err", err)

			return
		}

		onChange()
	})
}
