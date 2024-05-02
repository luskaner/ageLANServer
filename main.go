package main

import (
	"aoe2DELanServer/asset"
	"aoe2DELanServer/asset/cloud"
	"aoe2DELanServer/routes"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(session.Middleware())
	err := r.SetTrustedProxies(nil)
	if err != nil {
		log.Println(err)
		return
	}
	cloud.Initialize()
	asset.Initialize()
	routes.Initialize(r)
	log.Fatal(r.RunTLS(":443", "certificates/cert.pem", "certificates/key.pem"))
}
