package cmd

/*
Copyright © 2019 Simon Haxby

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

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var cssFile string
var mdvDir string

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mdv",
	Short: "Creates Browser viewable HTML files from markdown sources.",
	Run: func(cmd *cobra.Command, args []string) {
		// Grab markdown source
		if (len(args)) == 0 {
			log.Fatal("Need to specify MD file to preview.")
			return
		}
		mdFile := args[0]

		// Check to see that Pandoc is on PATH
		_, err := exec.LookPath("pandoc")
		handle(err)

		// Get HTML file path generated from markdown source
		tempFilePath := getTempFile(mdFile, cssFile, mdvDir)

		// Open HTML file via default browser
		openCmd := exec.Command("open", tempFilePath)
		openCmd.Run()

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set mdv directory var
	mdvDir = os.Getenv("HOME") + "/.mdv"

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default "+mdvDir+"/config.yaml)",
	)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// CSS for HTML formatting
	rootCmd.PersistentFlags().StringVar(
		&cssFile,
		"css",
		os.Getenv("HOME")+"/.pandoc/default.css",
		"CSS file used to format output HTML",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".mdv" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".mdv/config.yaml")
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
