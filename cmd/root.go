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
	"fmt"
	"github.com/alsritter/middlebaby/pkg/startup"
	"github.com/alsritter/middlebaby/pkg/startup/generic"
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
	rootCmd.AddCommand(CommandServe(startup.Startup, generic.NewConfig()))
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
