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
	"alsritter.icu/middlebaby/internal/log"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"

	config "alsritter.icu/middlebaby/internal/config"
)

var (
	logLevel string
	cfgFile  string
	flagApp  string
)

var rootCmd = &cobra.Command{
	Use:   "middlebaby",
	Short: "仿照 middlewomen 编写的 Mock 工具",
	Long:  `仿照 middlewomen 编写的 Mock 工具`,
}

func init() {
	// 设置传递的函数，以便在每个命令的调用执行方法。
	cobra.OnInitialize(initConfig)
	// 指定配置文件
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.middlebaby.yaml)")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "INFO", "Log level")
	rootCmd.PersistentFlags().StringVarP(&flagApp, "app", "", "", "启动的app路径")

	// 设置日志级别
	log.SetLevel(logLevel)
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func initConfig() {
	if cfgFile != "" {
		// 使用 --config 指定的配置文件路径
		viper.SetConfigFile(cfgFile)
	} else {
		// 找到 home 目录
		// home, err := os.UserHomeDir()
		// cobra.CheckErr(err)

		// 在 home 目录查询 .middlebaby.yaml 文件
		// viper.AddConfigPath(home)
		viper.AddConfigPath("./")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".middlebaby")
	}

	if err := viper.Unmarshal(&config.GlobalConfigVar); err != nil {
		log.Errorf("配置文件序列化成结构体失败: %s", err.Error())
	} else {
		log.Debugf("读取配置文件数据: %+v", config.GlobalConfigVar)
	}
}
