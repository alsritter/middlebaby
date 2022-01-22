/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"

	config "alsritter.icu/middlebaby/internal/config"
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

	imposters = make(map[string][]proxy.Imposter)
)

func init() {
	// Set up the function passed so that the method is executed on each command invocation.
	cobra.OnInitialize(initConfig)
	// Specifying a configuration file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $WORKSPACE/.middlebaby.yaml)")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "DEBUG", "Log level")
	rootCmd.PersistentFlags().StringVarP(&flagApp, "app", "", "", "Startup app path")

	// set log level.
	log.SetLevel(logLevel)
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func initConfig() {
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

	if err := viper.Unmarshal(&config.GlobalConfigVar); err != nil {
		log.Fatalf("failed to serialize configuration file to structure: %s", err.Error())
	} else {
		log.Debugf("Read configuration file data: %+v", config.GlobalConfigVar)
	}

	for _, filePath := range config.GlobalConfigVar.HttpFiles {
		loadImposter(filePath)
	}

	if config.GlobalConfigVar.Watcher {
		runWatcher(true, config.GlobalConfigVar.HttpFiles...)
	}
}

func runWatcher(canWatch bool, pathToWatch ...string) *watcher.Watcher {
	if !canWatch {
		return nil
	}

	w, err := config.InitializeWatcher(pathToWatch...)
	if err != nil {
		log.Fatal(err)
	}

	config.AttachWatcher(w, func(evn watcher.Event) {
		loadImposter(evn.Path)
	})
	return w
}

func loadImposter(filePath string) {
	if !filepath.IsAbs(filePath) {
		if fp, err := filepath.Abs(filePath); err != nil {
			log.Errorf("to absolute representation path err: %s", err)
			return
		} else {
			filePath = fp
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("%w: error trying to read config file: %s", err, filePath)
	}
	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)

	var imposter []proxy.Imposter
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		log.Errorf("%w: error while unmarshal configFile file %s", err, filePath)
	}

	imposters[filePath] = imposter
}
