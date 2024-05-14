package main

import (
	"common"
	"genCert/internal"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	certificatePairFolder := common.CertificatePairFolder(os.Args[0])
	if certificatePairFolder == "" {
		log.Fatal("Failed to determine certificate pair folder")
	}
	if common.HasCertificatePair(os.Args[0]) {
		log.Fatal("Already have certificate pair, delete it manually and re-run.")
	}
	if !internal.GenerateCertificatePair(certificatePairFolder) {
		log.Fatal("Could not generate certificate pair.")
	} else {
		log.Println("Certificate pair generated successfully.")
	}
}
