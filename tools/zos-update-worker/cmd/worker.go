/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/zos4/tools/zos-update-version/internal"
)

var rootCmd = &cobra.Command{
	Use:   "zos-update-version",
	Short: "A worker to update the version of zos",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ok, _ := cmd.Flags().GetBool("debug"); ok {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		src, err := cmd.Flags().GetString("src")
		if err != nil {
			return err
		}

		dst, err := cmd.Flags().GetString("dst")
		if err != nil {
			return err
		}

		params := internal.Params{}
		interval, err := cmd.Flags().GetInt("interval")
		if err != nil {
			return err
		}
		params.Interval = time.Duration(interval) * time.Minute

		production, err := cmd.Flags().GetString("main-url")
		if err != nil {
			return err
		}
		params.MainUrl = production

		test, err := cmd.Flags().GetString("test-url")
		if err != nil {
			return err
		}
		params.TestUrl = test

		qa, err := cmd.Flags().GetString("qa-url")
		if err != nil {
			return err
		}
		params.QAUrl = qa

		worker, err := internal.NewWorker(src, dst, params)
		if err != nil {
			return err
		}
		worker.UpdateWithInterval(cmd.Context())
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	log.Logger = log.Output(zerolog.NewConsoleWriter())

	cobra.OnInitialize()

	rootCmd.Flags().StringP("src", "s", "tf-autobuilder", "Enter your source directory")
	rootCmd.Flags().StringP("dst", "d", "tf-zos", "Enter your destination directory")
	rootCmd.Flags().IntP("interval", "i", 10, "Enter the interval between each update")
	rootCmd.Flags().Bool("debug", false, "enable debug logging")
	rootCmd.Flags().StringP("main-url", "m", "https://registrar.prod4.grid.tf", "Enter your mainnet registrar urls")
	rootCmd.Flags().StringP("test-url", "t", "https://registrar.test4.grid.tf", "Enter your testnet registrar urls")
	rootCmd.Flags().StringP("qa-url", "q", "https://registrar.qa4.grid.tf", "Enter your qanet registrar urls")
}
