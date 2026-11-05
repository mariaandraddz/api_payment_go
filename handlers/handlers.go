package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mariaandraddz/api-payment_go/models"
	"github.com/mariaandraddz/api-payment_go/repository"
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

func ReversalHandler(c *gin.Context) {
	var request models.ReversalRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := services.ReversalAuthorization(request.AuthID, request.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func ClearingHandler(c *gin.Context) {
	var request models.ClearingBatchRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services.ProcessBatchClearing(&request))
}
func GetAllCards(c *gin.Context) {
	var cardList []models.CardResponse
	authsPorCartao := make(map[string][]models.AuthorizationResponse)

	for _, auth := range repository.Authorizations {
		authsPorCartao[auth.CardID] = append(authsPorCartao[auth.CardID], *auth)
	}

	for _, card := range repository.Cards {
		auths := authsPorCartao[card.ID]
		if auths == nil {
			auths = []models.AuthorizationResponse{}
		}

		sort.Slice(auths, func(i, j int) bool {
			id1, err1 := strconv.Atoi(auths[i].AuthID)
			id2, err2 := strconv.Atoi(auths[j].AuthID)
			if err1 == nil && err2 == nil {
				return id1 < id2
			}
			return auths[i].AuthID < auths[j].AuthID
		})

		cardList = append(cardList, models.CardResponse{
			VirtualCard:    *card,
			Authorizations: auths,
		})
	}

	sort.Slice(cardList, func(i, j int) bool {
		// Caso 'VirtualCard' não seja um struct embutido (embedded field),
		// utilize cardList[i].VirtualCard.ID
		id1, err1 := strconv.Atoi(cardList[i].VirtualCard.ID)
		id2, err2 := strconv.Atoi(cardList[j].VirtualCard.ID)

		if err1 == nil && err2 == nil {
			return id1 < id2
		}
		return cardList[i].VirtualCard.ID < cardList[j].VirtualCard.ID
	})

	c.JSON(http.StatusOK, cardList)
}

func GetCard(c *gin.Context) {
	id := c.Param("id")
	card, exists := repository.Cards[id]

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": repository.ErrCardNotFound.Error()})
		return
	}

	var auths []models.AuthorizationResponse

	for _, auth := range repository.Authorizations {
		if auth.CardID == id {
			auths = append(auths, *auth)
		}
	}

	sort.Slice(auths, func(i, j int) bool {
		id1, _ := strconv.Atoi(auths[i].AuthID)
		id2, _ := strconv.Atoi(auths[j].AuthID)
		return id1 < id2
	})

	if auths == nil {
		auths = []models.AuthorizationResponse{}
	}
	c.JSON(http.StatusOK, models.CardResponse{
		VirtualCard:    *card,
		Authorizations: auths,
	})
}

func SeedHandler(c *gin.Context) {
	totalCards := 1000
	amountPerAuth := 100.00
	totalAuthsCount := 0

	for i := 0; i < totalCards; i++ {
		cardID := fmt.Sprintf("%d", i)
		repository.Cards[cardID] = &models.VirtualCard{
			ID:             cardID,
			Name:           fmt.Sprintf("Card %d", i),
			CurrentBalance: 600.00,
			CardLimit:      500.00,
			CardActive:     true,
		}
	}

	for cardID, card := range repository.Cards {
		if !card.CardActive {
			continue
		}

		for j := 1; j <= 2; j++ {
			authResp := services.Authorize(cardID, amountPerAuth)
			if authResp.Status == models.AUTHORIZED {
				totalAuthsCount++

				clearingReq := models.ClearingBatchRequest{
					Clearings: []models.ClearingItem{
						{
							CardID:         cardID,
							AuthID:         authResp.AuthID,
							CurrentBalance: authResp.CurrentBalance,
							Amount:         authResp.Amount,
							CardLimit:      authResp.CardLimit,
						},
					},
				}
				services.ProcessBatchClearing(&clearingReq)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_cards": len(repository.Cards),
		"total_auths": totalAuthsCount,
	})
}

func GenerateSettlementPayloadHandler(c *gin.Context) {
	var settlementItems []models.SettlementItem

	for _, clearing := range repository.ClearedTransactions {
		if clearing.Status == models.CLEARED {
			settlementItems = append(settlementItems, models.SettlementItem{
				CardID:         clearing.CardID,
				AuthID:         clearing.AuthID,
				Amount:         clearing.Amount,
				CurrentBalance: clearing.CurrentBalance,
				CardLimit:      clearing.CardLimit,
				Status:         clearing.Status,
			})
		}
	}
	c.JSON(http.StatusOK, models.SettlementBatchRequest{
		Settlements: settlementItems,
	})
}

func HandleGetBatch(c *gin.Context) {
	batchID := c.Param("batch_id")
	var req *models.SettlementBatchRequest

	batch, exists := repository.BatchStore.Load(batchID)

	if exists {
		c.JSON(http.StatusOK, gin.H{
			"batch": batch,
		})
		return
	}
	err := c.ShouldBindJSON(&req)
	c.JSON(http.StatusNotFound, gin.H{
		"error": err.Error(),
	})
}

func RefundHandler(c *gin.Context) {
	var request models.RefundRequest

	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := services.Refund(request.AuthID, request.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func SettlementHandlerBatch(c *gin.Context) {
	var request models.SettlementBatchRequest

	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": models.ProcessStatusFailed,
		})
		return
	}
	batchID := services.ProcessBatchSettlement(&request)

	c.JSON(http.StatusAccepted, gin.H{
		"batch":   batchID,
		"status":  models.ProcessStatusProcessing,
		"message": "The batch settlement process has been initiated.",
	})
}
