package services

import (
	"errors"
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

func ReversalAuthorization(authID string, amount float64) (*models.ReversalResponse, error) {
	auth, exists := repository.Authorizations[authID]

	if !exists || auth == nil {
		return nil, errors.New("Authorization not found")
	} else if auth.Status == models.CLEARED || auth.Status == models.SETTLED {
		return nil, errors.New("The transaction has already gone through clearing/settlement")
	} else if auth.Status != models.AUTHORIZED {
		return nil, errors.New("Authorization not found")
	} else if amount < 0 {
		return nil, errors.New("Negative amount")
	} else if auth.Amount != amount {
		return nil, errors.New("Check the refund amount again")
	}

	card, exists := repository.Cards[auth.CardID]

	if !exists || card == nil {
		return nil, repository.ErrCardNotFound
	} else if auth.Amount < amount {
		return nil, errors.New("Amount greater than authorized")
	}

	card.CardLimit += amount
	auth.Status = models.REVERSED

	return &models.ReversalResponse{
		AuthID:         auth.AuthID,
		CardID:         card.ID,
		Amount:         amount,
		CurrentBalance: card.CurrentBalance,
		CardLimit:      card.CardLimit,
		Status:         models.REVERSED,
	}, nil
}

func Refund(authID string, amount float64) (*models.RefundResponse, error) {
	clearedItem, exists := repository.ClearedTransactions[authID]

	if !exists || clearedItem == nil {
		return nil, errors.New("Clearing was not performed")
	} else if clearedItem.Status != models.SETTLED {
		return nil, errors.New("Settlement was not performed")
	}

	card, cardExists := repository.Cards[clearedItem.CardID]

	if !cardExists || card == nil {
		return nil, repository.ErrCardNotFound
	} else if clearedItem.Amount < amount {
		return nil, errors.New("Reimbursement amount greater settled amount")
	}

	card.CurrentBalance += amount
	clearedItem.Status = models.REFUNDED

	if auth, ok := repository.Authorizations[authID]; ok {
		auth.Status = models.REVERSED
	}

	return &models.RefundResponse{
		AuthID:         authID,
		CardID:         card.ID,
		Amount:         amount,
		CurrentBalance: card.CurrentBalance,
		CardLimit:      card.CardLimit,
		Status:         models.REFUNDED,
	}, nil
}
