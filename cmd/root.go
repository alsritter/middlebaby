/*
Copyright © 2021 alsritter@outlook.com

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
	"context"
	"fmt"
	"github.com/alsritter/middlebaby/pkg/startup"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/spf13/cobra"
)

const (
	asciiImage = `
-----------------------------------------------------------
              _     __    ____     __          __         
   ____ ___  (_)___/ /___/ / /__  / /_  ____ _/ /_  __  __
  / __ '__ \/ / __  / __  / / _ \/ __ \/ __ '/ __ \/ / / /
 / / / / / / / /_/ / /_/ / /  __/ /_/ / /_/ / /_/ / /_/ / 
/_/ /_/ /_/_/\__,_/\__,_/_/\___/_.___/\__,_/_.___/\__, /  
                                                 /____/   
-----------------------------------------------------------
Powered by: alsritter
	`
)

var (
	rootCmd = &cobra.Command{
		Use:     "middlebaby",
		Short:   "middlebaby",
		Long:    `a auto mock tool.`,
		Version: "",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(asciiImage)
			_ = cmd.Help()
		},
	}
)

func init() {
	rootCmd.AddCommand(CommandServe(Setup, config))
}

var config = startup.NewConfig()

func Setup(ctx context.Context) {
	log, err := logger.New(config.Log, "main")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := util.RegisterExitHandlers(log, cancel)
	defer cancel()

	if err := config.Validate(); err != nil {
		log.Fatal(nil, "failed to validate config: %s", err)
	}

	if err := startup.Startup(ctx, cancel, config, log); err != nil {
		log.Fatal(nil, "serve startup fail: %s", err)
	}

	<-stop
	log.Info(nil, "Goodbye")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
