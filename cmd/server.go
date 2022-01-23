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
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

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
					log.Debug("Received: Interrupt signal.")
				case os.Kill:
					log.Debug("Received: Kill signal.")
				default:
					log.Debugf("Received other signal: %+v", s)
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

			done <- true

			if err := group.Wait(); err != nil {
				log.Error("Get errors: ", err)
			}
			os.Exit(0)
		}
	},
}

func runMockServe(group *errgroup.Group, done chan bool) {
	group.Go(func() error {
		port := config.GlobalConfigVar.Port
		s := newServer(port)

		go func() {
			for {
				<-FilechangeEvent
				if err := s.Shutdown(); nil != err {
					log.Fatalf("server shutdown failed, err: %v\n", err)
				}

				s = newServer(port)
				s.Run()
			}
		}()

		// make sure idle connections returned
		processed := make(chan struct{})
		go func() {
			switch {
			case <-done:
				close(done)
			}

			if err := s.Shutdown(); nil != err {
				log.Fatalf("server shutdown failed, err: %v\n", err)
			}
			close(FilechangeEvent)
			close(processed)
		}()

		// serve
		s.Run()
		// waiting for goroutine above processed
		<-processed

		return nil
	})
}

func mapToSlice(m map[string][]proxy.Imposter) []proxy.Imposter {
	s := make([]proxy.Imposter, 0, len(m))
	for _, v := range m {
		s = append(s, v...)
	}
	return s
}

func newServer(port int) proxy.Server {
	router := mux.NewRouter()
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: time.Second * 3,
		Handler:      handlers.CORS(proxy.PrepareAccessControl(config.GlobalConfigVar.CORS)...)(router),
	}

	return proxy.NewServer(router, httpServer, mapToSlice(Imposters))
}
