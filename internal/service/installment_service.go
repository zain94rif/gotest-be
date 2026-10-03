package service

import (
	"context"
	"fmt"
	"time"

	"ims-finance/internal/model"
	"ims-finance/internal/repository"
)

type InstallmentService struct {
	installmentRepo *repository.InstallmentRepository
}

func NewInstallmentService(repo *repository.InstallmentRepository) *InstallmentService {
	return &InstallmentService{installmentRepo: repo}
}

// InstallmentResponse adalah response untuk satu baris angsuran.
type InstallmentResponse struct {
	InstallmentNo int    `json:"installment_no"`
	Amount        int64  `json:"amount"`
	DueDate       string `json:"due_date"`
}

// GetInstallments mengambil jadwal angsuran. upTo=0 berarti ambil semua.
func (s *InstallmentService) GetInstallments(ctx context.Context, contractNo string, upTo int) ([]InstallmentResponse, error) {
	var (
		installments []model.Installment
		err          error
	)

	if upTo > 0 {
		installments, err = s.installmentRepo.GetByContractNoUpTo(ctx, contractNo, upTo)
	} else {
		installments, err = s.installmentRepo.GetByContractNo(ctx, contractNo)
	}
	if err != nil {
		return nil, fmt.Errorf("get installments: %w", err)
	}

	result := make([]InstallmentResponse, 0, len(installments))
	for _, inst := range installments {
		result = append(result, InstallmentResponse{
			InstallmentNo: inst.InstallmentNo,
			Amount:        inst.Amount,
			DueDate:       inst.DueDate.Format("2006-01-02"),
		})
	}
	return result, nil
}

// GetByID mengambil detail satu angsuran berdasarkan ID.
func (s *InstallmentService) GetByID(ctx context.Context, id int64) (*model.Installment, error) {
	inst, err := s.installmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("installment not found")
	}
	return inst, nil
}

// Update mengubah data angsuran.
func (s *InstallmentService) Update(ctx context.Context, id int64, req *model.UpdateInstallmentRequest) error {
	return s.installmentRepo.Update(ctx, id, req)
}

// Delete menghapus angsuran berdasarkan ID.
func (s *InstallmentService) Delete(ctx context.Context, id int64) error {
	return s.installmentRepo.Delete(ctx, id)
}

// OverdueResponse adalah response lengkap untuk endpoint overdue.
type OverdueResponse struct {
	ContractNo   string                     `json:"contract_no"`
	AsOf         string                     `json:"as_of"`
	TotalOverdue int64                      `json:"total_overdue"`
	TotalPenalty int64                      `json:"total_penalty"`
	Installments []model.OverdueInstallment `json:"installments"`
}

// GetOverdue menghitung angsuran overdue dan denda untuk suatu kontrak.
func (s *InstallmentService) GetOverdue(ctx context.Context, contractNo string, asOfDate string) (*OverdueResponse, error) {
	asOf, err := time.Parse("2006-01-02", asOfDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	rows, err := s.installmentRepo.GetOverdue(ctx, contractNo, asOfDate)
	if err != nil {
		return nil, err
	}

	var totalOverdue, totalPenalty int64
	overdueList := make([]model.OverdueInstallment, 0, len(rows))

	for _, row := range rows {
		dueDate, _ := time.Parse("2006-01-02", row.DueDate)
		daysLate := int(asOf.Sub(dueDate).Hours() / 24)

		remaining := row.Amount - row.PaidAmount
		penalty := CalculatePenalty(remaining, daysLate)

		totalOverdue += remaining
		totalPenalty += penalty

		overdueList = append(overdueList, model.OverdueInstallment{
			InstallmentNo:   row.InstallmentNo,
			DueDate:         row.DueDate,
			Amount:          row.Amount,
			PaidAmount:      row.PaidAmount,
			RemainingAmount: remaining,
			DaysLate:        daysLate,
			Penalty:         penalty,
		})
	}

	return &OverdueResponse{
		ContractNo:   contractNo,
		AsOf:         asOfDate,
		TotalOverdue: totalOverdue,
		TotalPenalty: totalPenalty,
		Installments: overdueList,
	}, nil
}

// GetClientOverdueSummary mengambil total overdue semua kontrak milik satu client.
func (s *InstallmentService) GetClientOverdueSummary(ctx context.Context, clientCode string, asOfDate string) (*model.ClientOverdueSummary, error) {
	if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	summary, err := s.installmentRepo.GetClientOverdueSummary(ctx, clientCode, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("no overdue data found for client %s: %w", clientCode, err)
	}
	summary.AsOf = asOfDate
	return summary, nil
}
