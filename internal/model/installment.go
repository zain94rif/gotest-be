package model

import "time"

type Installment struct {
	ID              int64     `json:"id"`
	ContractID      int64     `json:"contract_id"`
	InstallmentNo   int       `json:"installment_no"`
	DueDate         time.Time `json:"due_date"`
	Amount          int64     `json:"amount"`
	PrincipalAmount int64     `json:"principal_amount,omitempty"`
	InterestAmount  int64     `json:"interest_amount,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// OverdueInstallment adalah data angsuran yang sudah jatuh tempo dan belum lunas.
type OverdueInstallment struct {
	InstallmentNo   int    `json:"installment_no"`
	DueDate         string `json:"due_date"`
	Amount          int64  `json:"amount"`
	PaidAmount      int64  `json:"paid_amount"`
	RemainingAmount int64  `json:"remaining_amount"`
	DaysLate        int    `json:"days_late"`
	Penalty         int64  `json:"penalty"`
}

// UpdateInstallmentRequest untuk mengubah jadwal angsuran.
type UpdateInstallmentRequest struct {
	DueDate string `json:"due_date" binding:"required"`
	Amount  int64  `json:"amount" binding:"required,gt=0"`
}
