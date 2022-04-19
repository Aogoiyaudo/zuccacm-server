/*
Copyright © 2021 ZAwei <awei.zbw@qq.com>

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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"zuccacm-server/config"
	"zuccacm-server/handler"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zuccacm-server",
	Short: "Server of zuccacm.top",
	Long:  `Server of zuccacm.top`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Instance.Port), handler.Router); err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(Init)

	rootCmd.PersistentFlags().IntP("port", "p", 80, "serve port")

	bindFlag("ServerConfig.Port", "port")
}

func Init() {
	config.Init(config.DefaultConfigFile)
}

func bindFlag(configName, flagName string) {
	if err := viper.BindPFlag(configName, rootCmd.PersistentFlags().Lookup(flagName)); err != nil {
		log.WithFields(log.Fields{
			"Config name": configName,
			"Flag name":   flagName,
			"Error":       err,
		}).Fatal("Bind flag failed!")
	}
}
