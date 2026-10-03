package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ims-finance/internal/model"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

const paymentSelectFields = `id, installment_id, payment_date, amount, COALESCE(payment_reference, ''), created_at`

func scanPayment(row interface{ Scan(...any) error }) (*model.Payment, error) {
	var p model.Payment
	err := row.Scan(&p.ID, &p.InstallmentID, &p.PaymentDate, &p.Amount, &p.PaymentReference, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetAll mengambil semua data pembayaran.
func (r *PaymentRepository) GetAll(ctx context.Context) ([]model.Payment, error) {
	query := `SELECT ` + paymentSelectFields + ` FROM ujicoba.payments ORDER BY id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all payments: %w", err)
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, *p)
	}
	return payments, rows.Err()
}

// GetByID mengambil payment berdasarkan ID.
func (r *PaymentRepository) GetByID(ctx context.Context, id int64) (*model.Payment, error) {
	query := `SELECT ` + paymentSelectFields + ` FROM ujicoba.payments WHERE id = $1`

	p, err := scanPayment(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("get payment by id: %w", err)
	}
	return p, nil
}

// GetByInstallmentID mengambil semua payment untuk satu installment.
func (r *PaymentRepository) GetByInstallmentID(ctx context.Context, installmentID int64) ([]model.Payment, error) {
	query := `SELECT ` + paymentSelectFields + ` FROM ujicoba.payments WHERE installment_id = $1 ORDER BY id`

	rows, err := r.db.Query(ctx, query, installmentID)
	if err != nil {
		return nil, fmt.Errorf("get payments by installment: %w", err)
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, *p)
	}
	return payments, rows.Err()
}

// InsertPayment menyimpan pembayaran angsuran dalam transaction.
func (r *PaymentRepository) InsertPayment(ctx context.Context, tx pgx.Tx, p *model.Payment) error {
	query := `
		INSERT INTO ujicoba.payments (installment_id, payment_date, amount, payment_reference)
		VALUES ($1, $2, $3, $4)`

	_, err := tx.Exec(ctx, query, p.InstallmentID, p.PaymentDate, p.Amount, p.PaymentReference)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

// Update mengubah data pembayaran berdasarkan ID.
func (r *PaymentRepository) Update(ctx context.Context, id int64, req *model.UpdatePaymentRequest) error {
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return fmt.Errorf("invalid payment_date format: %w", err)
	}

	query := `
		UPDATE ujicoba.payments
		SET payment_date = $1, amount = $2, payment_reference = $3
		WHERE id = $4`

	cmd, err := r.db.Exec(ctx, query, paymentDate, req.Amount, req.PaymentReference, id)
	if err != nil {
		return fmt.Errorf("update payment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("payment not found")
	}
	return nil
}

// Delete menghapus payment berdasarkan ID.
func (r *PaymentRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM ujicoba.payments WHERE id = $1`

	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete payment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("payment not found")
	}
	return nil
}

// BeginTx memulai transaction baru untuk payment.
func (r *PaymentRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}
