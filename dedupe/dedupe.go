package dedupe

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"go.chimbori.app/sortastic/conf"
)

// A map of [hash] -> [array of files with that hash]
var filesByHash map[string]*[]conf.ExifToolInfo

// Dedupe: dedupes media files, ignoring differences in metadata.
// When given a set of files, Dedupe computes a hash for each file’s image content (ignoring
// metadata), and outputs a list of files that can be deleted. The output file is in
// Shell Script syntax, so it can be reviewed before executing.
func Dedupe(args []string) {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Printf("Usage: %s <file> <file> ...\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	computeHashes(args)
	pickDupes()
}

func computeHashes(fileNames []string) {
	filesByHash = make(map[string]*[]conf.ExifToolInfo)
	for _, fileName := range fileNames {
		x, err := getExif(fileName)
		if err != nil {
			log.Fatal(err)
		}

		hash := x.ImageDataHash
		if filesByHash[hash] == nil {
			filesByHash[hash] = &[]conf.ExifToolInfo{}
		}
		*filesByHash[hash] = append(*filesByHash[hash], *x)
	}
}

func pickDupes() {
	for _, files := range filesByHash {
		var keepFiles []conf.ExifToolInfo
		var discardFiles []conf.ExifToolInfo
		for _, file := range *files {
			if file.Title != "" { // If an image has an EXIF Title, add it to the keep list.
				keepFiles = append(keepFiles, file)
			} else {
				discardFiles = append(discardFiles, file)
			}
		}
		// Only delete discarded files if a kept file also exists for the same hash.
		if len(keepFiles) > 0 && len(discardFiles) > 0 {
			for _, keepFile := range keepFiles {
				fmt.Printf("# KEEP \"%s\" (%s)\n", keepFile.SourceFile, keepFile.Title)
			}
			for _, discardFile := range discardFiles {
				fmt.Printf("    rm \"%s\"\n", discardFile.SourceFile)
			}
		}
	}
}

func getExif(fileName string) (*conf.ExifToolInfo, error) {
	// Among many ways to get EXIF data from a file, `exiftool` is by far the best
	// battle-tested parser out there. Even though it involves shelling out to a
	// PERL executable, it is still preferred over native Go implementations.
	exiftoolOutput, err := exec.Command("exiftool", "-json",
		"-ImageDataHash",
		"-Title",
		"-SubLocation",
		"-LocationName",
		"-ObjectName",
		"-ContentLocationName",
		"-IPTC:Caption-Abstract",
		"-EXIF:ImageDescription",
		"-XMP-dc:Description",
		"-City",
		"-Province-State",
		"-State",
		"-Country",
		"-CountryCode",
		fileName).Output()
	if err != nil {
		return nil, err
	}

	exifToolInfo := conf.ExifToolInfos{}
	err = json.Unmarshal(exiftoolOutput, &exifToolInfo)
	if err != nil {
		return nil, err
	}

	return &exifToolInfo[0], nil
}
