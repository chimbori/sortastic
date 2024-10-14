package home

import (
	"log"
	"net/http"

	"go.chimbori.app/sortastic/conf"
	"go.chimbori.app/sortastic/web/media"
)

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		log.Println("Only GET allowed for", req.URL.Path)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	var directories []string
	for _, item := range conf.Config.Directories {
		directories = append(directories, item.Slug)
	}

	HomeTempl(HomePage{
		Title:           conf.AppName,
		MediaPathPrefix: media.MediaPathPrefix,
		Directories:     directories,
	}).Render(req.Context(), w)
}
