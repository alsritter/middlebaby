/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init config file",
	Long:  `init config file`,
	Run: func(cmd *cobra.Command, args []string) {
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

	},
}
