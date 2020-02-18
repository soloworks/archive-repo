package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const ignorefile string = ".archiveignore"

func main() {

	ignore := make(map[string]bool)          // Setup files to ignore, set defaults
	ignore[ignorefile] = true                // Ignore the config file
	ignore[filepath.Base(os.Args[0])] = true // Ignore this executable

	// Read in .archiveignore file if present
	f, err := os.Open(ignorefile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error with %s: %s", ignorefile, err)
	}
	defer f.Close()

	if f != nil {
		// Store all Globs in a map from the file
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			// // Get all files that match the ignore pattern
			filepath.Walk(strings.TrimSpace(scanner.Text()), func(path string, info os.FileInfo, err error) error {
				if path != "" {
					ignore[path] = true
				}
				return nil
			})
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	buf := new(bytes.Buffer) // Create Byte buffer to hold archive
	w := zip.NewWriter(buf)  // Create Zip Writer on Buffer
	defer w.Close()          // Defer close

	// Walk all files in this folder and sub-folders
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if ignore[path] != true && !info.IsDir() {

			// Open the file
			fileToZip, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fileToZip.Close()

			// Get the file information
			info, err := fileToZip.Stat()
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Using FileInfoHeader() above only uses the basename of the file. If we want
			// to preserve the folder structure we can overwrite this with the full path.
			header.Name = filepath.ToSlash(path)

			// Change to deflate to gain better compression
			// see http://golang.org/pkg/archive/zip/#pkg-constants
			header.Method = zip.Deflate

			writer, err := w.CreateHeader(header)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, fileToZip)

		}
		return nil
	})

	// Push archive to Server

}
