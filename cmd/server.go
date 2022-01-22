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
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run Mock serve",
	Long:  `run Mock serve`,
	Run: func(cmd *cobra.Command, args []string) {
		if flagApp != "" {
			if _, err := os.Stat(flagApp); err != nil {
				log.Fatal("target app err: ", err)
			}

			group := new(errgroup.Group)

			sigs := make(chan os.Signal, 1)
			done := make(chan bool, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			runMockServe(group, done)

			command := exec.Command(flagApp)

			go func() {
				switch s := <-sigs; s {
				case os.Interrupt:
					log.Info("Received: Interrupt signal.")
				case os.Kill:
					log.Info("Received: Kill signal.")
				default:
					log.Infof("Received other signal: %+v", s)
				}

				done <- true
			}()

			port := config.GlobalConfigVar.Port
			parentEnv := os.Environ()
			// set proxy path
			parentEnv = append(parentEnv, fmt.Sprintf("HTTP_PROXY=http://127.0.0.1:%d", port))
			parentEnv = append(parentEnv, fmt.Sprintf("http_proxy=http://127.0.0.1:%d", port))
			parentEnv = append(parentEnv, fmt.Sprintf("HTTPS_PROXY=http://127.0.0.1:%d", port))
			parentEnv = append(parentEnv, fmt.Sprintf("https_proxy=http://127.0.0.1:%d", port))

			command.Env = parentEnv
			// TODO: add filter support
			command.Stdout = os.Stdout
			command.Stderr = os.Stdout

			if err := command.Run(); err != nil {
				if _, isExist := err.(*exec.ExitError); !isExist {
					log.Fatal("Failed to start the program to be tested, err:", err)
				}
			}

			if err := group.Wait(); err != nil {
				fmt.Println("Get errors: ", err)
			} else {
				fmt.Println("Get all num successfully!")
			}

			os.Exit(0)
		}
	},
}

func runMockServe(group *errgroup.Group, done chan bool) {
	group.Go(func() error {

		switch {
		case <-done:
		}

		close(done)
		return nil
	})
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
