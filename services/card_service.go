package services

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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

func ProcessBatchClearing(request *models.ClearingBatchRequest) models.ClearingBatchResponse {
	var results []models.ClearingResults

	successCount, failureCount := 0, 0
	for _, req := range request.Clearings {
		auth := repository.Authorizations[req.AuthID]

		if auth == nil {
			failureCount++
			results = append(results, models.ClearingResults{
				AuthID: req.AuthID,
				CardID: req.CardID,
				Status: "authID not found",
			})
		} else if auth.Status == models.CLEARED {
			failureCount++
			results = append(results, models.ClearingResults{
				AuthID: req.AuthID,
				CardID: req.CardID,
				Status: "Authorization has already been cleared",
			})
		} else if auth.CardID != req.CardID {
			failureCount++
			results = append(results, models.ClearingResults{
				AuthID: req.AuthID,
				CardID: req.CardID,
				Status: "This card was not authorized",
			})
		} else if auth.Status != models.AUTHORIZED {
			failureCount++
			results = append(results, models.ClearingResults{
				AuthID: req.AuthID,
				CardID: req.CardID,
				Status: "The authorization was not approved",
			})
		} else if auth.Amount < req.Amount {
			failureCount++
			results = append(results, models.ClearingResults{
				AuthID: req.AuthID,
				CardID: req.CardID,
				Status: "Amount higher than authorized",
			})
		} else {
			auth.Status = models.CLEARED
			repository.ClearedTransactions[req.AuthID] = &models.ClearingItem{
				AuthID:         req.AuthID,
				CardID:         req.CardID,
				CurrentBalance: req.CurrentBalance,
				Amount:         req.Amount,
				CardLimit:      req.CardLimit,
				Status:         models.CLEARED,
			}
			successCount++
			results = append(results, models.ClearingResults{
				AuthID: req.AuthID,
				CardID: req.CardID,
				Status: models.CLEARED,
			})
		}
	}
	return models.ClearingBatchResponse{
		Total_processed: len(request.Clearings),
		Authorized:      successCount,
		Failure:         failureCount,
		Results:         results,
	}
}

func ProcessBatchSettlement(request *models.SettlementBatchRequest) string {
	newBatchID := atomic.AddUint64(&repository.BatchCounter, 1)
	batchID := fmt.Sprintf("%d", newBatchID)
	now := time.Now()

	batch := &models.BatchStatus{
		BatchID:   batchID,
		Status:    models.ProcessStatusProcessing,
		CreatedAt: now.Format("2006-01-02 15:04:05"),
	}

	repository.BatchStore.Store(batchID, batch)

	go func() {
		defer func() {
			repository.BatchStore.Store(batchID, batch)
		}()

		if request == nil || request.Settlements == nil {
			batch.BatchID = batchID
			batch.Status = models.ProcessStatusFailed
			batch.Result = &models.SettlementBatchResponse{}
			return
		}

		total := len(request.Settlements)
		if total == 0 {
			batch.BatchID = batchID
			batch.Status = models.ProcessStatusCompleted
			batch.Result = &models.SettlementBatchResponse{}
			return
		}

		resultChan := make(chan models.SettlementResults, total)
		var wg sync.WaitGroup

		// Itera sobre a lista de liquidações enviadas na requisição
		for _, item := range request.Settlements {
			wg.Add(1)

			newTransaction := atomic.AddUint64(&repository.TransactionCounter, 1)
			transactionID := fmt.Sprintf("%d", newTransaction)

			go func(item models.SettlementItem, txID string) {
				defer wg.Done()

				repository.Mu.Lock()
				cleared, exists := repository.ClearedTransactions[item.AuthID]

				if cleared == nil || !exists || cleared.Status != models.CLEARED {
					repository.Mu.Unlock()
					resultChan <- models.SettlementResults{
						TransactionID: txID,
						AuthID:        item.AuthID,
						CardID:        item.CardID,
						Status:        models.NOT_CLEARED,
					}
					return
				}

				if cleared.CardID != item.CardID {
					repository.Mu.Unlock()
					resultChan <- models.SettlementResults{
						TransactionID: txID,
						AuthID:        item.AuthID,
						CardID:        item.CardID,
						Status:        repository.ErrCardNotFound.Error(),
					}
					return
				}

				card, exists := repository.Cards[cleared.CardID]
				if card == nil || !exists {
					repository.Mu.Unlock()
					resultChan <- models.SettlementResults{
						TransactionID: txID,
						AuthID:        item.AuthID,
						CardID:        item.CardID,
						Status:        repository.ErrCardNotFound.Error(),
					}
					return
				}

				// Atualizações de saldo e limite protegidas por mutex
				card.CurrentBalance -= cleared.Amount
				card.CardLimit += cleared.Amount
				cleared.Status = models.SETTLED

				for i := range repository.Authorizations {
					if repository.Authorizations[i].AuthID == item.AuthID {
						repository.Authorizations[i].Status = models.SETTLED
						break
					}
				}
				repository.Mu.Unlock()

				resultChan <- models.SettlementResults{
					TransactionID: txID,
					AuthID:        item.AuthID,
					CardID:        item.CardID,
					Status:        models.SETTLED,
				}
			}(item, transactionID)
		}

		// Aguarda o processamento de todos os itens e fecha o canal
		wg.Wait()
		close(resultChan)

		var results []models.SettlementResults
		successCount, failureCount := 0, 0

		for result := range resultChan {
			results = append(results, result)
			if result.Status == models.SETTLED {
				successCount++
			} else {
				failureCount++
			}
		}

		batch.Result = &models.SettlementBatchResponse{
			Total_processed: total,
			Authorized:      successCount,
			Failure:         failureCount,
			Results:         results,
		}

		if failureCount == 0 {
			batch.Status = models.ProcessStatusCompleted
		} else if successCount > 0 {
			batch.Status = models.ProcessStatusCompleted
		} else {
			batch.Status = models.ProcessStatusFailed
		}
	}()

	return batch.BatchID
}
