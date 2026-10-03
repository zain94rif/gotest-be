package service

import (
	"context"
	"fmt"
	"time"

	"ims-finance/internal/model"
	"ims-finance/internal/repository"
)

type PaymentService struct {
	paymentRepo     *repository.PaymentRepository
	installmentRepo *repository.InstallmentRepository
}

func NewPaymentService(
	paymentRepo *repository.PaymentRepository,
	installmentRepo *repository.InstallmentRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		installmentRepo: installmentRepo,
	}
}

// GetAll mengambil semua payment.
func (s *PaymentService) GetAll(ctx context.Context) ([]model.Payment, error) {
	return s.paymentRepo.GetAll(ctx)
}

// GetByID mengambil satu payment berdasarkan ID.
func (s *PaymentService) GetByID(ctx context.Context, id int64) (*model.Payment, error) {
	p, err := s.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("payment not found")
	}
	return p, nil
}

// GetByInstallmentID mengambil semua payment untuk satu installment.
func (s *PaymentService) GetByInstallmentID(ctx context.Context, installmentID int64) ([]model.Payment, error) {
	return s.paymentRepo.GetByInstallmentID(ctx, installmentID)
}

// CreatePayment mencatat pembayaran angsuran.
func (s *PaymentService) CreatePayment(ctx context.Context, req *model.CreatePaymentRequest) error {
	// Validasi installment exist
	_, err := s.installmentRepo.GetByID(ctx, req.InstallmentID)
	if err != nil {
		return fmt.Errorf("installment not found: %w", err)
	}

	// Parse tanggal
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return fmt.Errorf("invalid payment_date format: %w", err)
	}

	payment := &model.Payment{
		InstallmentID:    req.InstallmentID,
		PaymentDate:      paymentDate,
		Amount:           req.Amount,
		PaymentReference: req.PaymentReference,
	}

	tx, err := s.paymentRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = s.paymentRepo.InsertPayment(ctx, tx, payment); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Update mengubah data payment.
func (s *PaymentService) Update(ctx context.Context, id int64, req *model.UpdatePaymentRequest) error {
	return s.paymentRepo.Update(ctx, id, req)
}

// Delete menghapus payment berdasarkan ID.
func (s *PaymentService) Delete(ctx context.Context, id int64) error {
	return s.paymentRepo.Delete(ctx, id)
}
