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
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cssFile string
	mdvDir  string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "mdv",
		Short: "Creates Browser viewable HTML files from markdown sources.",

		// Return the path of HTML file generated via Pandoc
		// for a user specified markdown file. We check a
		// BadgerDB store to see if a HTML file has been generated
		// for our source file by checking if an MD5 hash of our
		// file is a key in the store. If the hash is unseen then
		// we can Pandoc and generate a new file.
		Run: func(cmd *cobra.Command, args []string) {
			// Sanity checks
			preflightChecks(args)

			mdFile := args[0]
			force, _ := cmd.Flags().GetBool("force")

			// Get HTML file path generated from markdown source
			// Create mdv $HOME .dir if it doesn't exist.
			if _, err := os.Stat(mdvDir); os.IsNotExist(err) {
				os.Mkdir(mdvDir, 0700)
			}

			db := initBadger(mdvDir)
			defer db.Close()

			// Check if we get a cache hit
			fileHash, err := hashFileMd5(mdFile)
			fatal(err)
			path, err := checkCache(fileHash, db)
			fmt.Println(fileHash)

			// We render the HTML if the cache misses or the user passes
			// the 'force' flag.
			var tempFilePath string
			if err == nil && !force {
				fmt.Println("Cache hit.")
				tempFilePath = string(path)
			} else {
				pandocArgs, tempFilePath := initPandocArgs(mdFile, tempFilePath)
				runPandoc(pandocArgs)
				updateCache(fileHash, tempFilePath, db)
			}

			// Open HTML file via default browser
			openCmd := exec.Command("open", tempFilePath)
			err = openCmd.Run()
			fatal(err)

		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Here you will define your flags and configuration settings.
// Cobra supports persistent flags, which, if defined here,
// will be global for your application.
func init() {
	cobra.OnInitialize(initConfig)

	// Set mdv directory var
	mdvDir = path.Join(os.Getenv("HOME"), ".mdv")

	// Specifies path to optional parameter cfg file.
	// We can bind our parameters to config key-value pairs
	// loaded by Viper.
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default "+mdvDir+"/config.yaml)",
	)
	// CSS for HTML formatting
	rootCmd.PersistentFlags().StringVar(
		&cssFile,
		"css",
		path.Join(os.Getenv("HOME"), ".pandoc", "default.css"),
		"CSS file used to format output HTML",
	)
	viper.BindPFlag("css", rootCmd.PersistentFlags().Lookup("css"))

	// Force render of Markdown flag
	rootCmd.PersistentFlags().BoolP("force", "f", false, "Force Pandoc render - even source hasn't changed")
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
		viper.AddConfigPath(path.Join(home, ".mdv"))
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv() // Read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
