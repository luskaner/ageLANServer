package main

import (
	"aoe2DELanServer/files"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(models.SessionMiddleware())
	err := r.SetTrustedProxies(nil)
	if err != nil {
		log.Println(err)
		return
	}
	files.Initialize()
	routes.Initialize(r)
	log.Fatal(r.RunTLS(":443", "static/certificates/cert.pem", "static/certificates/key.pem"))
}
