package clean

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Clean: deletes empty directories and other unwanted files from the specified paths.
func Clean(args []string) {
	flag.Parse()
	if len(flag.Args()) == 1 {
		fmt.Printf("Usage: %s <path> <path> ...\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	rootDirs := args
	for _, rootDir := range rootDirs {
		numPaths := 0
		err := filepath.Walk(rootDir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				numPaths++
				if info.IsDir() {
					// Directory cleanup actions
					deleteIfEmpty(path)
				} else {
					// File cleanup actions
					deleteDSStore(path)
				}
				return nil
			})
		if err != nil {
			log.Println(err)
		}
		log.Printf("Examined \"%s\" (%d files/dirs)", rootDir, numPaths)
	}
}

func deleteIfEmpty(path string) {
	dirContents, err := os.ReadDir(path)
	if err != nil {
		log.Println(err)
	}
	if len(dirContents) == 0 {
		logAndDelete(path)
	}
}

func deleteDSStore(path string) {
	basename := filepath.Base(path)
	if basename == ".DS_Store" {
		logAndDelete(path)
	}
}

func logAndDelete(path string) {
	log.Printf("Deleting %s\n", path)
	err := os.RemoveAll(path)
	if err != nil {
		log.Println(err)
	}
}
