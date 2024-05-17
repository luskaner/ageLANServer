package main

import (
	"common"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"server/files"
	"server/ip"
	"server/middleware"
	"server/routes"
)

func main() {
	mux := http.NewServeMux()
	files.Initialize()
	routes.Initialize(mux)
	sessionMux := middleware.SessionMiddleware(mux)
	addr := ip.ResolveHost(files.Config.Host)
	if addr == nil {
		log.Fatal("Failed to resolve host")
	}
	server := &http.Server{
		Addr:    addr.String() + ":443",
		Handler: handlers.LoggingHandler(os.Stdout, sessionMux),
	}
	if files.Config.Announce {
		go func() {
			ip.Announce(addr)
		}()
	}
	certificatePairFolder := common.CertificatePairFolder(os.Args[0])
	log.Fatal(server.ListenAndServeTLS(filepath.Join(certificatePairFolder, common.Cert), filepath.Join(certificatePairFolder, common.Key)))
}
