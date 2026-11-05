package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mariaandraddz/api-payment_go/handlers"
	"github.com/mariaandraddz/api-payment_go/healthcheck"
	"github.com/mariaandraddz/api-payment_go/observability"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	router := gin.Default()

	router.Use(observability.Metric())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", healthcheck.Health)
	router.GET("/card/:id", handlers.GetCard)
	router.GET("/card", handlers.GetAllCards)
	router.GET("generate-settlement-payload", handlers.GenerateSettlementPayloadHandler)
	router.GET("/batch/:batch_id", handlers.HandleGetBatch)
	router.POST("/authorize", handlers.AuthorizationHandler)
	router.POST("/reversal", handlers.ReversalHandler)
	router.POST("/clearing-batch", handlers.ClearingHandler)
	router.POST("/refund", handlers.RefundHandler)
	router.POST("/seed", handlers.SeedHandler)
	router.POST("/settlements-batch", handlers.SettlementHandlerBatch)

	router.Run(":8080")
}
