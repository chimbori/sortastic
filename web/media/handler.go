package media

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/facette/natsort"

	"go.chimbori.app/sortastic/conf"
)

var MediaPathPrefix = "/media/"

func MediaHandler(w http.ResponseWriter, req *http.Request) {
	slug := req.PathValue("slug")
	relLocalPath := req.PathValue("path")

	var dir conf.AppConfigDirectory
	for _, d := range conf.Config.Directories {
		if d.Slug == slug {
			dir = d
		}
	}
	absLocalPath := filepath.Clean(filepath.Join(dir.Source, relLocalPath))

	switch req.Method {
	case http.MethodGet:
		fi, err := os.Stat(absLocalPath)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if filepath.Base(absLocalPath)[0:1] == "." && len(absLocalPath) > 1 {
			log.Printf("Refusing to serve hidden file: %s\n", absLocalPath)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if fi.IsDir() {
			serveIndex(w, req, dir, relLocalPath)
		} else {
			http.ServeFile(w, req, absLocalPath)
		}

	case http.MethodPost:
		if err := req.ParseForm(); err != nil {
			log.Println("ParseForm() err:", err)
			return
		}

		mediaFile, err := createMediaFile(absLocalPath, req.URL.Path, dir)
		if err != nil {
			log.Println(err)
			return
		}

		switch req.FormValue("action") {
		case "approve":
			log.Printf("Approving %s", absLocalPath)
			moveFile(mediaFile.AbsPath, filepath.Clean(filepath.Join(dir.Destination, relLocalPath)), "moveToDestination")
			MovedToDestinationTempl(mediaFile).Render(req.Context(), w)

		case "delete":
			log.Printf("Deleting %s", absLocalPath)
			// Delete the main file, then delete all related sidecar files.
			absTrashPath := filepath.Clean(filepath.Join(dir.Trash, relLocalPath))
			moveFile(mediaFile.AbsPath, absTrashPath, "moveToTrash")
			for _, sidecarRelLocalPath := range getSidecarFiles(relLocalPath) {
				sidecarAbsLocalPath := filepath.Clean(filepath.Join(dir.Source, sidecarRelLocalPath))
				sidecarAbsTrashPath := filepath.Clean(filepath.Join(dir.Trash, sidecarRelLocalPath))
				moveFile(sidecarAbsLocalPath, sidecarAbsTrashPath, "moveToTrash")
			}
			RestoreFromTrashTempl(mediaFile).Render(req.Context(), w)

		case "restore":
			log.Printf("Restoring from Trash: %s", absLocalPath)
			// Restore the main file, then restore all related sidecar files.
			absTrashPath := filepath.Clean(filepath.Join(dir.Trash, relLocalPath))
			moveFile(absTrashPath, mediaFile.AbsPath, "restoreFromTrash")
			for _, sidecarRelLocalPath := range getSidecarFiles(relLocalPath) {
				sidecarAbsLocalPath := filepath.Clean(filepath.Join(dir.Source, sidecarRelLocalPath))
				sidecarAbsTrashPath := filepath.Clean(filepath.Join(dir.Trash, sidecarRelLocalPath))
				moveFile(sidecarAbsTrashPath, sidecarAbsLocalPath, "restoreFromTrash")
			}
			MediaFileTempl(mediaFile).Render(req.Context(), w)

		case "rename-start":
			log.Printf("Initiating Rename %s", absLocalPath)
			RenameStartedTempl(mediaFile).Render(req.Context(), w)

		case "rename-save":
			log.Printf("Saving Renamed File %s", absLocalPath)
			renameToFileName := req.FormValue("rename-to")
			absRenamedPath := filepath.Join(filepath.Dir(mediaFile.AbsPath), renameToFileName)
			renamedUrlPath, err := url.JoinPath(path.Dir(req.URL.Path), renameToFileName)
			if err != nil {
				log.Println(err)
				return
			}
			moveFile(mediaFile.AbsPath, absRenamedPath, "renameSave")
			renamedMediaFile, err := createMediaFile(absRenamedPath, renamedUrlPath, dir)
			if err != nil {
				log.Println(err)
				return
			}
			MediaFileTempl(renamedMediaFile).Render(req.Context(), w)

		case "rename-cancel":
			log.Printf("Cancel Rename %s", absLocalPath)
			MediaFileTempl(mediaFile).Render(req.Context(), w)
		}
	}
}

func createMediaFile(relLocalPath string, urlPath string, dir conf.AppConfigDirectory) (mediaFile MediaFile, err error) {
	absPath, err := filepath.Abs(relLocalPath)
	if err != nil {
		log.Println(err)
		return
	}

	mediaFile = MediaFile{
		FileName:         filepath.Base(relLocalPath),
		UrlPath:          urlPath,
		AbsPath:          absPath,
		MediaType:        getMediaType(absPath),
		EditMode:         dir.Mode == "edit",
		DestinationAvail: dir.Destination != "",
		TrashAvail:       dir.Trash != "",
	}
	return
}

func serveIndex(w http.ResponseWriter, req *http.Request, dir conf.AppConfigDirectory, relLocalPath string) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	absLocalPath := filepath.Clean(filepath.Join(dir.Source, relLocalPath))
	files, err := os.ReadDir(absLocalPath)
	if err != nil {
		log.Println(err)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return natsort.Compare(strings.ToLower(files[i].Name()), strings.ToLower(files[j].Name()))
	})

	mediaFiles := make([]MediaFile, 0)
	for _, file := range files {
		if file.Name()[0:1] == "." {
			continue
		}

		urlPath, err := url.JoinPath(MediaPathPrefix, dir.Slug, relLocalPath, file.Name())
		if err != nil {
			log.Println(err)
			continue
		}

		mediaFiles = append(mediaFiles, MediaFile{
			FileName:         file.Name(),
			UrlPath:          urlPath,
			MediaType:        getMediaType(file.Name()),
			EditMode:         dir.Mode == "edit",
			DestinationAvail: dir.Destination != "",
			TrashAvail:       dir.Trash != "",
		})
	}

	IndexPageTempl(IndexPage{
		Title:    conf.AppName,
		Slug:     dir.Slug,
		UrlPath:  relLocalPath,
		NumFiles: len(mediaFiles),
		Files:    mediaFiles,
	}).Render(req.Context(), w)
}

func moveFile(srcFile string, destFile string, reason string) {
	log.Println(reason, srcFile, destFile)
	os.MkdirAll(filepath.Dir(destFile), os.ModePerm)

	fi, err := os.Stat(srcFile)
	if err != nil {
		log.Println(err)
	} else {
		if fi.IsDir() {
			// If this is a directory, then recurse & move its contents, otherwise moving the
			// directory itself will fail if the same directory already exists at the destination.
			files, err := os.ReadDir(srcFile)
			if err != nil {
				log.Println(err)
			} else {
				for _, subFile := range files {
					moveFile(
						filepath.Join(srcFile, subFile.Name()),
						filepath.Join(destFile, subFile.Name()), reason)
				}
			}
		} else {
			err := os.Rename(srcFile, destFile)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
