package web

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"syscall"

	"go.chimbori.app/sortastic/conf"
	"go.chimbori.app/sortastic/web/home"
	"go.chimbori.app/sortastic/web/media"
	"golang.org/x/term"
)

//go:embed static
var staticFiles embed.FS

func Web(args []string) {
	if conf.Config == nil {
		log.Fatal("Missing config file: sortastic.yml")
	}

	if conf.Config.Web.Username != "" && conf.Config.Web.Password == "" {
		// Request a password without echoing it to the console or in shell history.
		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
		}
		conf.Config.Web.Password = string(bytePassword)
		fmt.Println()
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServer(http.FS(staticFiles)))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./static/favicon.svg")
	})
	mux.HandleFunc("GET /", home.HomeHandler)

	mux.HandleFunc("GET /media/{slug}", media.MediaHandler)
	mux.HandleFunc("GET /media/{slug}/{path...}", media.MediaHandler)
	mux.HandleFunc("POST /media/{slug}/{path...}", media.MediaHandler)

	log.Printf("Listening at <http://%s:%d/>", conf.Config.Web.Host, conf.Config.Web.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", conf.Config.Web.Host, conf.Config.Web.Port), mux))
}
