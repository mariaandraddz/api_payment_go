package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mariaandraddz/api-payment_go/models"
	"github.com/mariaandraddz/api-payment_go/services"
)

func AuthorizationHandler(c *gin.Context) {
	n := (rand.Intn(100) + 100)
	time.Sleep(time.Duration(n) * time.Millisecond)

	var request models.AuthorizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	path := c.FullPath()
	if path == "" {
		path = "/authorize"
	}

	response := services.Authorize(request.CardID, request.Amount)

	switch response.Status {
	case observability.StatusAuthSuccess:
		observability.RecordAuthSuccess(c.Request.Method, path, "200")
		c.JSON(http.StatusOK, response)

	case observability.StatusDeclineInsufficientFunds:
		observability.RecordDeclineInsufficientCards(c.Request.Method, path)
		c.JSON(http.StatusBadRequest, response)

	case observability.StatusDeclineInactiveCard:
		observability.RecordDeclineInactiveCard(c.Request.Method, path)
		c.JSON(http.StatusBadRequest, response)

	default:
		c.JSON(http.StatusBadRequest, response)
	}

}
