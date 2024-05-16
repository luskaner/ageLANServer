package main

import (
	"common"
	"flag"
	"genCert/internal"
	"log"
	"os"
)

type Arguments struct {
	Force bool
}

func main() {
	log.SetFlags(0)
	args := Arguments{}
	flag.BoolVar(&args.Force, "force", false, "Whether to overwrite existing certificate pair")
	flag.Parse()
	certificatePairFolder := common.CertificatePairFolder(os.Args[0])
	if certificatePairFolder == "" {
		log.Fatal("Failed to determine certificate pair folder")
	}
	if !args.Force && common.HasCertificatePair(os.Args[0]) {
		log.Fatal("Already have certificate pair and force is false, set force to true or delete it manually.")
	}
	if !internal.GenerateCertificatePair(certificatePairFolder) {
		log.Fatal("Could not generate certificate pair.")
	} else {
		log.Println("Certificate pair generated successfully.")
	}
}
