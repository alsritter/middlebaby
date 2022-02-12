/*
Copyright © 2021 NAME HERE alsritter@outlook.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/alsritter/middlebaby/internal/log"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"

	config "github.com/alsritter/middlebaby/internal/file/config"
)

var (
	logLevel string
	cfgFile  string
	flagApp  string
)

var (
	rootCmd = &cobra.Command{
		Use:   "middlebaby",
		Short: "a Mock server tool.",
		Long:  `a Mock server tool.`,
	}

	GlobalConfigVar config.Config
)

func init() {
	// Set up the function passed so that the method is executed on each command invocation.
	cobra.OnInitialize(initConfig)
	// Specifying a configuration file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $WORKSPACE/.middlebaby.yaml)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", "INFO", "Log level")
	rootCmd.PersistentFlags().StringVarP(&flagApp, "app", "", "", "Startup app path")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func initConfig() {
	// set log level.
	log.SetLevel(logLevel)

	if cfgFile != "" {
		// use --config specifies the path to the configuration file.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./tests")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".middlebaby")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Failed to read the configuration file.")
	} else {
		log.Debugf("Configuration file to use: %s", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&GlobalConfigVar); err != nil {
		log.Fatalf("failed to serialize configuration file to structure: %s", err.Error())
	}
}
