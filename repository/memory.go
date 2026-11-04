package repository

import (
	"errors"
	"sync"

	"github.com/mariaandraddz/api-payment_go/models"
)

var (
	ErrCardNotFound      = errors.New("card not found")
	ErrInactiveCard      = errors.New("inactive card")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrNegativeValue     = errors.New("negative value")
)

var Cards = map[string]*models.VirtualCard{
	"1": {
		ID:             "1",
		Name:           "Card 1",
		CurrentBalance: 600.00,
		CardLimit:      500.00,
		CardActive:     true,
	},
	"2": {
		ID:             "2",
		Name:           "Card 2",
		CurrentBalance: 650.00,
		CardLimit:      550.00,
		CardActive:     true,
	},
	"3": {
		ID:             "3",
		Name:           "Card 3",
		CurrentBalance: 750.00,
		CardLimit:      550.00,
		CardActive:     false,
	},
	"4": {
		ID:             "4",
		Name:           "Card 4",
		CurrentBalance: 750.00,
		CardLimit:      550.00,
		CardActive:     true,
	},
	"5": {
		ID:             "5",
		Name:           "Card 5",
		CurrentBalance: 750.00,
		CardLimit:      550.00,
		CardActive:     true,
	},
}

var Authorizations = map[string]*models.AuthorizationResponse{}
var ClearedTransactions = map[string]*models.ClearingItem{}
var AuthCounter uint64
var BatchStore sync.Map
var BatchCounter uint64
var TransactionCounter uint64
var Mu sync.Mutex
