package model

import "time"

type Payment struct {
	ID               int64     `json:"id"`
	InstallmentID    int64     `json:"installment_id"`
	PaymentDate      time.Time `json:"payment_date"`
	Amount           int64     `json:"amount"`
	PaymentReference string    `json:"payment_reference,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreatePaymentRequest adalah request body untuk mencatat pembayaran angsuran.
type CreatePaymentRequest struct {
	InstallmentID    int64  `json:"installment_id" binding:"required"`
	PaymentDate      string `json:"payment_date" binding:"required"`
	Amount           int64  `json:"amount" binding:"required,gt=0"`
	PaymentReference string `json:"payment_reference"`
}

// UpdatePaymentRequest untuk mengubah data pembayaran.
type UpdatePaymentRequest struct {
	PaymentDate      string `json:"payment_date" binding:"required"`
	Amount           int64  `json:"amount" binding:"required,gt=0"`
	PaymentReference string `json:"payment_reference"`
}
