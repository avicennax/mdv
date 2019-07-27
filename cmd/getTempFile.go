package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	badger "github.com/dgraph-io/badger"
)

// Return the path of HTML file generated via Pandoc
// for a user specified markdown file. We check a
// BadgerDB store to see if a HTML file has been generated
// for our source file by checking if an MD5 hash of our
// file is a key in the store. If the hash is unseen then
// we can Pandoc and generate a new file.
func getTempFile(mdFile string, cssFile string, mdvDir string) string {
	// Create mdv $HOME .dir if it doesn't exist.
	if _, err := os.Stat(mdvDir); os.IsNotExist(err) {
		os.Mkdir(mdvDir, 0700)
	}
	// Configure Badger DB connection
	badgerConfig := badger.DefaultOptions(mdvDir + "/badger")
	fmt.Println(badgerConfig)
	db, err := badger.Open(badgerConfig)
	handle(err)
	defer db.Close()

	// Check to see if the generated HTML file has been cached.
	fileHash, err := hashFileMd5(mdFile)
	fmt.Println(fileHash)
	var pathCopy []byte
	err = db.View(func(txn *badger.Txn) error {
		path, err := txn.Get([]byte(fileHash))
		if err == nil {
			pathCopy, err = path.ValueCopy(nil)
			return nil
		}
		return err
	})
	var tempFilePath string
	if err == nil {
		fmt.Println("Cache hit.")
		tempFilePath = string(pathCopy)
	} else {
		// Otherwise generate a new HTML file via pandoc
		// Get temporary file
		tempDir, err := ioutil.TempDir("", "")
		handle(err)
		tempFilePath = path.Join(tempDir, "temp.html")

		// Generate HTML via Pandoc
		// TODO: catch errors raised by Pandoc.
		pandocCmd := exec.Command(
			"pandoc",
			"-o", tempFilePath,
			"--css", cssFile,
			"--self-contained",
			"--quiet",
			mdFile,
		)
		pandocCmd.Run()
		fmt.Println(os.Getenv("HOME") + "/.pandoc/default.css")

		// Update cache
		fmt.Println("Updating cache ...")
		err = db.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(fileHash), []byte(tempFilePath))
			err = txn.SetEntry(e)
			handle(err)
			return err
		})
		handle(err)
	}
	return tempFilePath
}
