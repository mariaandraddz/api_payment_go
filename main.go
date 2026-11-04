package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mariaandraddz/api-payment_go/healthcheck"
)

func main() {
	router := gin.Default()
	router.GET("/health", healthcheck.Health)

	router.Run(":8080")
}
