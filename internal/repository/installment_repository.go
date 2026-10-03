package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ims-finance/internal/model"
)

type InstallmentRepository struct {
	db *pgxpool.Pool
}

func NewInstallmentRepository(db *pgxpool.Pool) *InstallmentRepository {
	return &InstallmentRepository{db: db}
}

// InsertInstallment menyimpan satu angsuran dalam konteks transaction.
func (r *InstallmentRepository) InsertInstallment(ctx context.Context, tx pgx.Tx, inst *model.Installment) error {
	query := `
		INSERT INTO ujicoba.installments (contract_id, installment_no, due_date, amount)
		VALUES ($1, $2, $3, $4)`

	_, err := tx.Exec(ctx, query, inst.ContractID, inst.InstallmentNo, inst.DueDate, inst.Amount)
	if err != nil {
		return fmt.Errorf("insert installment: %w", err)
	}
	return nil
}

// GetByContractNo mengambil semua jadwal angsuran berdasarkan nomor kontrak.
func (r *InstallmentRepository) GetByContractNo(ctx context.Context, contractNo string) ([]model.Installment, error) {
	return r.queryInstallments(ctx, contractNo, 0)
}

// GetByContractNoUpTo mengambil jadwal angsuran sampai installment_no tertentu.
func (r *InstallmentRepository) GetByContractNoUpTo(ctx context.Context, contractNo string, upTo int) ([]model.Installment, error) {
	return r.queryInstallments(ctx, contractNo, upTo)
}

// queryInstallments query helper internal. upTo=0 berarti ambil semua.
func (r *InstallmentRepository) queryInstallments(ctx context.Context, contractNo string, upTo int) ([]model.Installment, error) {
	query := `
		SELECT i.id, i.contract_id, i.installment_no, i.due_date, i.amount, i.created_at
		FROM ujicoba.installments i
		JOIN ujicoba.contracts c ON c.id = i.contract_id
		WHERE c.contract_no = $1`

	args := []any{contractNo}
	if upTo > 0 {
		query += ` AND i.installment_no <= $2`
		args = append(args, upTo)
	}
	query += ` ORDER BY i.installment_no`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get installments: %w", err)
	}
	defer rows.Close()

	var installments []model.Installment
	for rows.Next() {
		var inst model.Installment
		if err := rows.Scan(
			&inst.ID, &inst.ContractID, &inst.InstallmentNo,
			&inst.DueDate, &inst.Amount, &inst.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan installment: %w", err)
		}
		installments = append(installments, inst)
	}
	return installments, rows.Err()
}

// GetByID mengambil installment berdasarkan ID.
func (r *InstallmentRepository) GetByID(ctx context.Context, id int64) (*model.Installment, error) {
	query := `
		SELECT id, contract_id, installment_no, due_date, amount, created_at
		FROM ujicoba.installments
		WHERE id = $1`

	var inst model.Installment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&inst.ID, &inst.ContractID, &inst.InstallmentNo,
		&inst.DueDate, &inst.Amount, &inst.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get installment by id: %w", err)
	}
	return &inst, nil
}

// Update mengubah due_date dan amount angsuran berdasarkan ID.
func (r *InstallmentRepository) Update(ctx context.Context, id int64, req *model.UpdateInstallmentRequest) error {
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return fmt.Errorf("invalid due_date format: %w", err)
	}

	query := `UPDATE ujicoba.installments SET due_date = $1, amount = $2 WHERE id = $3`

	cmd, err := r.db.Exec(ctx, query, dueDate, req.Amount, id)
	if err != nil {
		return fmt.Errorf("update installment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("installment not found")
	}
	return nil
}

// Delete menghapus angsuran berdasarkan ID.
func (r *InstallmentRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM ujicoba.installments WHERE id = $1`

	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete installment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("installment not found")
	}
	return nil
}

// OverdueRow adalah hasil query overdue dari database.
type OverdueRow struct {
	ID            int64
	InstallmentNo int
	DueDate       string
	Amount        int64
	PaidAmount    int64
}

// GetOverdue mengambil angsuran yang sudah jatuh tempo dan belum lunas.
func (r *InstallmentRepository) GetOverdue(ctx context.Context, contractNo string, asOfDate string) ([]OverdueRow, error) {
	query := `
		SELECT
			i.id,
			i.installment_no,
			i.due_date::TEXT,
			i.amount,
			COALESCE(SUM(p.amount), 0) AS paid_amount
		FROM ujicoba.installments i
		JOIN ujicoba.contracts c ON c.id = i.contract_id
		LEFT JOIN ujicoba.payments p ON p.installment_id = i.id
		WHERE c.contract_no = $1
		  AND i.due_date < $2::DATE
		GROUP BY i.id, i.installment_no, i.due_date, i.amount
		HAVING COALESCE(SUM(p.amount), 0) < i.amount
		ORDER BY i.installment_no`

	rows, err := r.db.Query(ctx, query, contractNo, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("get overdue: %w", err)
	}
	defer rows.Close()

	var result []OverdueRow
	for rows.Next() {
		var row OverdueRow
		if err := rows.Scan(&row.ID, &row.InstallmentNo, &row.DueDate, &row.Amount, &row.PaidAmount); err != nil {
			return nil, fmt.Errorf("scan overdue row: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// GetClientOverdueSummary mengambil total overdue seluruh kontrak milik satu client.
// Digunakan untuk endpoint report: /reports/client-overdue?client_code=SUGUS&date=2024-08-14
func (r *InstallmentRepository) GetClientOverdueSummary(ctx context.Context, clientCode string, asOfDate string) (*model.ClientOverdueSummary, error) {
	query := `
		SELECT
			cl.client_code,
			cl.name,
			COUNT(i.id)::INT AS overdue_count,
			CAST(SUM(i.amount - COALESCE(p_sum.paid_amount, 0)) AS BIGINT) AS total_overdue
		FROM ujicoba.clients cl
		JOIN ujicoba.contracts c ON c.client_id = cl.id
		JOIN ujicoba.installments i ON i.contract_id = c.id
		LEFT JOIN (
			SELECT installment_id, SUM(amount) AS paid_amount
			FROM ujicoba.payments
			GROUP BY installment_id
		) p_sum ON p_sum.installment_id = i.id
		WHERE cl.client_code = $1
		  AND i.due_date < $2::DATE
		  AND i.amount > COALESCE(p_sum.paid_amount, 0)
		GROUP BY cl.client_code, cl.name`

	var s model.ClientOverdueSummary
	err := r.db.QueryRow(ctx, query, clientCode, asOfDate).Scan(
		&s.ClientCode, &s.Name, &s.OverdueCount, &s.TotalOverdue,
	)
	if err != nil {
		return nil, fmt.Errorf("get client overdue summary: %w", err)
	}
	return &s, nil
}
