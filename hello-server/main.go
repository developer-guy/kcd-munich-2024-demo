package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var Version = "0.1.0"

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World %s!", Version)
	})
	r.Run() // listen on 0.0.0.0:8080
}
