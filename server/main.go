package main

import (
	"common"
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
	certificatePairFolder := common.CertificatePairFolder(os.Args[0])
	log.Fatal(server.ListenAndServeTLS(certificatePairFolder+common.Cert, certificatePairFolder+common.Key))
}
