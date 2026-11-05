package models

const (
	NotAuthorize = "NotAuthorize"
	REVERSED     = "CANCELLED"
	AUTHORIZED   = "AUTHORIZED"
	CLEARED      = "CLEARED"
	SETTLED      = "SETTLED"
	NOT_CLEARED  = "NOT CLEARED"
	REFUNDED     = "REFUNDED"

	ProcessStatusCompleted  = "COMPLETED"
	ProcessStatusFailed     = "FAILED"
	ProcessStatusProcessing = "PROCESSING"
)

type VirtualCard struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	CurrentBalance float64 `json:"current_balance"`
	CardLimit      float64 `json:"card_limit"`
	CardActive     bool    `json:"card_active"`
}
type AuthorizationRequest struct {
	CardID string  `json:"card_id"`
	Amount float64 `json:"amount"`
}

type AuthorizationResponse struct {
	AuthID         string  `json:"auth_id"`
	CardID         string  `json:"card_id"`
	Amount         float64 `json:"amount"`
	CurrentBalance float64 `json:"current_balance"`
	CardLimit      float64 `json:"card_limit"`
	Status         string  `json:"status"`
}

type ClearingItem struct {
	CardID         string  `json:"card_id"`
	AuthID         string  `json:"auth_id"`
	Amount         float64 `json:"amount"`
	CurrentBalance float64 `json:"current_balance"`
	CardLimit      float64 `json:"card_limit"`
	Status         string  `json:"status"`
}

type ClearingBatchRequest struct {
	Clearings []ClearingItem `json:"clearings"`
}

type ClearingResults struct {
	CardID string `json:"card_id"`
	AuthID string `json:"auth_id"`
	Status string `json:"status"`
}

type ClearingBatchResponse struct {
	Total_processed int               `json:"total_processed"`
	Authorized      int               `json:"authorized"`
	Failure         int               `json:"failure"`
	Results         []ClearingResults `json:"results"`
}

type SettlementItem struct {
	CardID         string  `json:"card_id"`
	AuthID         string  `json:"auth_id"`
	Amount         float64 `json:"amount"`
	CurrentBalance float64 `json:"current_balance"`
	CardLimit      float64 `json:"card_limit"`
	Status         string  `json:"status"`
}

type SettlementBatchRequest struct {
	Settlements []SettlementItem `json:"settlements"`
}

type SettlementBatchResponse struct {
	Total_processed int               `json:"total_processed"`
	Authorized      int               `json:"authorized"`
	Failure         int               `json:"failure"`
	Results         []ClearingResults `json:"results"`
}
type SettlementResults struct {
	TransactionID string `json:"transaction_id"`
	CardID        string `json:"card_id"`
	AuthID        string `json:"auth_id"`
	Status        string `json:"status"`
}

type ReversalRequest struct {
	AuthID string  `json:"auth_id"`
	Amount float64 `json:"amount"`
}

type ReversalResponse struct {
	CardID         string  `json:"card_id"`
	AuthID         string  `json:"auth_id"`
	Amount         float64 `json:"amount"`
	CurrentBalance float64 `json:"current_balance"`
	CardLimit      float64 `json:"card_limit"`
	Status         string  `json:"status"`
}

type RefundRequest struct {
	AuthID string  `json:"auth_id"`
	Amount float64 `json:"amount"`
}

type RefundResponse struct {
	CardID         string  `json:"card_id"`
	AuthID         string  `json:"auth_id"`
	Amount         float64 `json:"amount"`
	CurrentBalance float64 `json:"current_balance"`
	CardLimit      float64 `json:"card_limit"`
	Status         string  `json:"status"`
}

type CardResponse struct {
	VirtualCard
	Authorizations []AuthorizationResponse `json:"authorizations"`
}

type BatchStatus struct {
	BatchID   string                   `json:"batch_id"`
	Status    string                   `json:"status"`
	CreatedAt string                   `json:"created_at"`
	Result    *SettlementBatchResponse `json:"result"`
}
