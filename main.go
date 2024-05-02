package main

import (
	"aoe2DELanServer/files"
	"aoe2DELanServer/middleware"
	"aoe2DELanServer/routes"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	routes.Initialize(mux)
	sessionMux := middleware.SessionMiddleware(mux)
	server := &http.Server{
		Addr: ":443",
		//Handler: handlers.RecoveryHandler()(handlers.LoggingHandler(os.Stdout, mux)),
		Handler: handlers.LoggingHandler(os.Stdout, sessionMux),
	}
	files.Initialize()
	log.Fatal(server.ListenAndServeTLS("static/certificates/cert.pem", "static/certificates/key.pem"))
}
