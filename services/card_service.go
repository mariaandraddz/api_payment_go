package services

import (
	"fmt"
	"sync/atomic"

	"github.com/mariaandraddz/api-payment_go/models"
	"github.com/mariaandraddz/api-payment_go/repository"
)

func Authorize(cardID string, amount float64) *models.AuthorizationResponse {
	card, exists := repository.Cards[cardID]

	if !exists || card == nil {
		return &models.AuthorizationResponse{
			CardID: cardID,
			Amount: amount,
			Status: repository.ErrCardNotFound.Error() + models.NotAuthorize,
		}
	}

	if amount < 0 {
		return &models.AuthorizationResponse{
			CardID:         cardID,
			Amount:         amount,
			CurrentBalance: card.CurrentBalance,
			CardLimit:      card.CardLimit,
			Status:         repository.ErrNegativeValue.Error() + models.NotAuthorize,
		}
	}
	if !card.CardActive {
		return &models.AuthorizationResponse{
			CardID:         cardID,
			Amount:         amount,
			CurrentBalance: card.CurrentBalance,
			CardLimit:      card.CardLimit,
			Status:         repository.ErrInactiveCard.Error() + models.NotAuthorize,
		}
	}

	if card.CardLimit < amount {
		return &models.AuthorizationResponse{
			CardID:         cardID,
			Amount:         amount,
			CurrentBalance: card.CurrentBalance,
			CardLimit:      card.CardLimit,
			Status:         repository.ErrInsufficientFunds.Error() + models.NotAuthorize,
		}
	}
	card.CardLimit -= amount

	newAuthID := atomic.AddUint64(&repository.AuthCounter, 1)
	authID := fmt.Sprintf("%d", newAuthID)

	auth := &models.AuthorizationResponse{
		AuthID:         authID,
		CardID:         cardID,
		Amount:         amount,
		CurrentBalance: card.CurrentBalance,
		CardLimit:      card.CardLimit,
		Status:         models.AUTHORIZED,
	}
	repository.Authorizations[cardID] = auth
	return auth
}
