package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

func createZipfile(filename string, messageList []MessageTuple) {

	// create the unzipped file
	unzippedName := strings.TrimSuffix(filename, ".zip")
	baseUnzipped := path.Base(unzippedName)
	log.Printf("INFO: creating %s...", unzippedName)
	unzipped, err := os.Create(unzippedName)
	fatalIfError(err)
	w := bufio.NewWriter(unzipped)
	for ix := range messageList {
		ts := time.Unix(int64(messageList[ix].FirstSent/1000), 0) // cos our format is epoch plus milliseconds
		s := fmt.Sprintf("Id: %s (first sent: %s)\n", messageList[ix].id, ts)
		_, err = w.WriteString(s)
		fatalIfError(err)
	}
	w.Flush()
	unzipped.Close()

	// create the zipfile
	log.Printf("INFO: creating %s...", filename)
	archive, err := os.Create(filename)
	fatalIfError(err)
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	z, err := zipWriter.Create(baseUnzipped)
	fatalIfError(err)
	f, err := os.Open(unzippedName)
	fatalIfError(err)
	_, err = io.Copy(z, f)
	fatalIfError(err)
	f.Close()
	zipWriter.Close()
}

//
// end of file
//
