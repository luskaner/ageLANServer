package main

import (
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
	"server/announce"
	"server/files"
	"server/middleware"
	"server/routes"
)

func main() {
	mux := http.NewServeMux()
	files.Initialize()
	routes.Initialize(mux)
	sessionMux := middleware.SessionMiddleware(mux)
	server := &http.Server{
		Addr:    files.Config.Host + ":443",
		Handler: handlers.LoggingHandler(os.Stdout, sessionMux),
	}
	if files.Config.Announce {
		go func() {
			announce.Announce(files.Config.Host)
		}()
	}
	log.Fatal(server.ListenAndServeTLS("resources/certificates/cert.pem", "resources/certificates/key.pem"))
}
