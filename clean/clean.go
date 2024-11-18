package clean

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var dryRun = true

// Clean: deletes empty directories and other unwanted files from the specified paths.
func Clean(args []string) {

	cleanFlags := flag.NewFlagSet("clean", flag.ExitOnError)
	cleanFlags.BoolVar(&dryRun, "dryrun", true, "when true, does not delete anything")
	cleanFlags.Parse(args)

	if len(cleanFlags.Args()) == 0 {
		cleanFlags.Usage()
		fmt.Fprintf(os.Stderr, "\n%s clean [options] <path> <path> …\n\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	if dryRun {
		log.Println("DRY RUN: Nothing will be deleted.")
	}

	rootDirs := cleanFlags.Args()
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
	if !dryRun {
		log.Printf("Deleting %s\n", path)
		err := os.RemoveAll(path)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Printf("Would delete %s\n", path)
	}
}
