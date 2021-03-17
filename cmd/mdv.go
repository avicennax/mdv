package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	badger "github.com/dgraph-io/badger"
	"github.com/spf13/viper"
)

func initBadger(mdvDir string) *badger.DB {
	// Configure Badger DB connection
	badgerConfig := badger.DefaultOptions(path.Join(mdvDir, "badger"))
	// Quiet verbose Badger logging
	badgerConfig.Logger = nil
	db, err := badger.Open(badgerConfig)
	fatal(err)

	return db
}

func preflightChecks(args []string) {
	// Check to see that Pandoc is on PATH
	_, err := exec.LookPath("pandoc")
	fatal(err)

	// Grab markdown source
	if (len(args)) == 0 {
		log.Fatal("Need to specify Markdown file to preview.")
	}
}

// Check to see if the generated HTML file has been cached.
func checkCache(fileHash string, db *badger.DB) ([]byte, error) {
	var path []byte
	err := db.View(func(txn *badger.Txn) error {
		htmlFilePath, err := txn.Get([]byte(fileHash))
		if err == nil {
			path, err = htmlFilePath.ValueCopy(nil)
			return nil
		}
		return err
	})
	return path, err
}

func updateCache(fileHash string, tempFilePath string, db *badger.DB) {
	fmt.Println("Updating cache ...")
	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(fileHash), []byte(tempFilePath))
		err := txn.SetEntry(e)
		fatal(err)
		return err
	})
	fatal(err)
}

func initPandocArgs(mdFile string, tempFilePath string) ([]string, string) {
	// Otherwise generate a new HTML file via pandoc
	// Get temporary file
	tempDir, err := ioutil.TempDir("", "")
	fatal(err)
	tempFilePath = path.Join(tempDir, "temp.html")

	args := []string{
		"-o", tempFilePath,
		"--self-contained",
		"--quiet",
		mdFile,
	}

	// Use Viper to fetch CSS filepath
	cssFile := viper.GetString("css")
	// Check to see if css is defined.
	_, err = os.Stat(cssFile)
	if os.IsNotExist(err) {
		fmt.Printf("CSS file: '%s' not found\n", cssFile)
	} else {
		args = append([]string{"--css", cssFile}, args...)
		fmt.Printf("Using CSS file: %s\n", cssFile)
	}
	return args, tempFilePath
}

func runPandoc(args []string) {
	pandocCmd := exec.Command("pandoc", args...)
	err := pandocCmd.Run()
	fatal(err)
}
