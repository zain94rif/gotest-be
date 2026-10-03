package service

import (
	"context"
	"fmt"
	"time"

	"ims-finance/internal/model"
	"ims-finance/internal/repository"
)

type ContractService struct {
	contractRepo    *repository.ContractRepository
	installmentRepo *repository.InstallmentRepository
}

func NewContractService(
	contractRepo *repository.ContractRepository,
	installmentRepo *repository.InstallmentRepository,
) *ContractService {
	return &ContractService{
		contractRepo:    contractRepo,
		installmentRepo: installmentRepo,
	}
}

// GetAll mengambil semua kontrak.
func (s *ContractService) GetAll(ctx context.Context) ([]model.Contract, error) {
	return s.contractRepo.GetAll(ctx)
}

// GetByContractNo mengambil detail satu kontrak.
func (s *ContractService) GetByContractNo(ctx context.Context, contractNo string) (*model.Contract, error) {
	c, err := s.contractRepo.GetByContractNo(ctx, contractNo)
	if err != nil {
		return nil, fmt.Errorf("contract not found")
	}
	return c, nil
}

// CreateContract menghitung kredit, menyimpan kontrak dan jadwal angsuran dalam satu transaction.
func (s *ContractService) CreateContract(ctx context.Context, req *model.CreateContractRequest) (*model.Contract, error) {
	// Parse tanggal
	contractDate, err := time.Parse("2006-01-02", req.ContractDate)
	if err != nil {
		return nil, fmt.Errorf("invalid contract_date format: %w", err)
	}
	firstDueDate, err := time.Parse("2006-01-02", req.FirstDueDate)
	if err != nil {
		return nil, fmt.Errorf("invalid first_due_date format: %w", err)
	}

	// Hitung kredit
	calc := CalculateLoan(req.OTR, req.DPPercentage, req.TenorMonths)

	contract := &model.Contract{
		ContractNo:         req.ContractNo,
		ClientID:           req.ClientID,
		OTR:                calc.OTR,
		DPPercentage:       req.DPPercentage,
		DPAmount:           calc.DP,
		PrincipalAmount:    calc.Principal,
		TenorMonths:        calc.TenorMonths,
		InterestRate:       calc.InterestRate,
		InterestAmount:     calc.InterestAmount,
		TotalAmount:        calc.TotalAmount,
		MonthlyInstallment: calc.MonthlyInstallment,
		ContractDate:       contractDate,
		FirstDueDate:       firstDueDate,
	}

	// Mulai transaction
	tx, err := s.contractRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Insert contract
	contractID, err := s.contractRepo.InsertContract(ctx, tx, contract)
	if err != nil {
		return nil, err
	}
	contract.ID = contractID

	// Generate jadwal angsuran
	for i := 0; i < req.TenorMonths; i++ {
		dueDate := firstDueDate.AddDate(0, i, 0)
		inst := &model.Installment{
			ContractID:    contractID,
			InstallmentNo: i + 1,
			DueDate:       dueDate,
			Amount:        calc.MonthlyInstallment,
		}
		if err = s.installmentRepo.InsertInstallment(ctx, tx, inst); err != nil {
			return nil, err
		}
	}

	// Commit
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return contract, nil
}

// Update mengubah data non-finansial kontrak.
func (s *ContractService) Update(ctx context.Context, contractNo string, req *model.UpdateContractRequest) error {
	return s.contractRepo.Update(ctx, contractNo, req)
}

// Delete menghapus kontrak berdasarkan contract_no.
func (s *ContractService) Delete(ctx context.Context, contractNo string) error {
	return s.contractRepo.Delete(ctx, contractNo)
}

func (s *ContractService) GetOverdueSummary(ctx context.Context, contractNo string, dateStr string) (*model.OverdueSummaryResponse, error) {
	if dateStr == "" {
		dateStr = "2024-08-14"
	}
	return s.contractRepo.GetOverdueSummary(ctx, contractNo, dateStr)
}

func (s *ContractService) GetPenalties(ctx context.Context, contractNo string, dateStr string) ([]model.PenaltyDetailResponse, error) {
	if dateStr == "" {
		dateStr = "2024-08-14"
	}
	return s.contractRepo.GetPenalties(ctx, contractNo, dateStr)
}
