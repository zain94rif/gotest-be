package model

import "time"

type Contract struct {
	ID                 int64     `json:"id"`
	ContractNo         string    `json:"contract_no"`
	ClientID           int64     `json:"client_id"`
	OTR                int64     `json:"otr"`
	DPPercentage       float64   `json:"dp_percentage"`
	DPAmount           int64     `json:"dp_amount"`
	PrincipalAmount    int64     `json:"principal_amount"`
	TenorMonths        int       `json:"tenor_months"`
	InterestRate       float64   `json:"interest_rate"`
	InterestAmount     int64     `json:"interest_amount"`
	TotalAmount        int64     `json:"total_amount"`
	MonthlyInstallment int64     `json:"monthly_installment"`
	ContractDate       time.Time `json:"contract_date"`
	FirstDueDate       time.Time `json:"first_due_date"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CreateContractRequest adalah request body untuk membuat kontrak baru.
type CreateContractRequest struct {
	ClientID     int64   `json:"client_id" binding:"required"`
	ContractNo   string  `json:"contract_no" binding:"required"`
	OTR          int64   `json:"otr" binding:"required,gt=0"`
	DPPercentage float64 `json:"dp_percentage" binding:"min=0,max=100"`
	TenorMonths  int     `json:"tenor_months" binding:"required,gt=0"`
	ContractDate string  `json:"contract_date" binding:"required"`
	FirstDueDate string  `json:"first_due_date" binding:"required"`
}

// UpdateContractRequest untuk mengubah data non-finansial kontrak.
type UpdateContractRequest struct {
	ClientID     int64  `json:"client_id"`
	ContractDate string `json:"contract_date"`
	FirstDueDate string `json:"first_due_date"`
}

type OverdueSummaryResponse struct {
	ContractNo              string `json:"contract_no"`
	ClientName              string `json:"client_name"`
	TotalAngsuranJatuhTempo int64  `json:"total_angsuran_jatuh_tempo"`
}

type PenaltyDetailResponse struct {
	ContractNo        string `json:"contract_no"`
	ClientName        string `json:"client_name"`
	InstallmentNo     int    `json:"installment_no"`
	HariKeterlambatan int    `json:"hari_keterlambatan"`
	TotalDenda        int64  `json:"total_denda"`
}
