package home

import (
	"log"
	"net/http"

	"go.chimbori.app/sortastic/conf"
	"go.chimbori.app/sortastic/web/error"
	"go.chimbori.app/sortastic/web/media"
)

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	if conf.Config.Web.Username != "" && conf.Config.Web.Password != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok || !isAuthorized(reqUsername, reqPassword) {
			w.Header().Add("WWW-Authenticate", `Basic realm="Cenote Auth"`)
			w.WriteHeader(http.StatusUnauthorized)
			error.ErrorTempl("Please enter valid credentials to access this app.", "Unauthorized").Render(req.Context(), w)
			return
		}
	}

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

// Trivial check for now
func isAuthorized(reqUsername, reqPassword string) bool {
	return reqUsername == conf.Config.Web.Username && reqPassword == conf.Config.Web.Password
}
