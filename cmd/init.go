/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/alsritter/middlebaby/pkg/types/interact"
	"github.com/alsritter/middlebaby/pkg/types/mbcase"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	initConfig = false
	initCase   = false
)

func init() {
	// TODO: Remove here...
	initCmd.PersistentFlags().BoolVarP(&initConfig, "config", "f", false, "输出配置文件")
	initCmd.PersistentFlags().BoolVarP(&initCase, "case", "c", false, "输出用例文件")
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init config file",
	Long:  `init config file`,
	Run: func(cmd *cobra.Command, args []string) {
		if initConfig {
			yamlData, err := yaml.Marshal(&config)
			if err != nil {
				fmt.Printf("Error while Marshaling. %v", err)
			}

			fmt.Println(" --- YAML ---")
			fmt.Println(string(yamlData)) // yamlData will be in bytes. So converting it to string.

			fileName := "test.yaml"
			err = ioutil.WriteFile(fileName, yamlData, 0644)
			if err != nil {
				panic("Unable to write data into the file")
			}
		}

		if initCase {
			t := &mbcase.ItfTask{
				TaskInfo: &mbcase.TaskInfo{
					Protocol:           "",
					ServiceName:        "",
					ServiceMethod:      "",
					ServiceDescription: "",
					ServicePath:        "",
					ServiceProtoFile:   "",
				},
				SetUp: []*mbcase.Command{
					{
						TypeName: "",
						Commands: []string{},
					},
				},
				Mocks: []*interact.ImposterMockCase{
					{
						Request:  interact.Request{},
						Response: interact.Response{},
					},
				},
				TearDown: []*mbcase.Command{
					{
						TypeName: "",
						Commands: []string{},
					},
				},
				Cases: []*mbcase.CaseTask{
					{
						Name:        "",
						Description: "",
						SetUp:       []*mbcase.Command{},
						Mocks:       []*interact.ImposterMockCase{},
						Request:     &mbcase.CaseRequest{},
						Assert:      &mbcase.Assert{},
						TearDown:    []*mbcase.Command{},
					},
				},
			}
			caseData, err := json.Marshal(&t)
			if err != nil {
				fmt.Printf("Error while Marshaling. %v", err)
			}

			fmt.Println(" --- JSON ---")
			fmt.Println(string(caseData))
			fileName := "test.json"
			err = ioutil.WriteFile(fileName, caseData, 0644)
			if err != nil {
				panic("Unable to write data into the file")
			}
		}
	},
}
