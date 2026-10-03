package repository

import (
	"context"
	"fmt"
	"time"

	"ims-finance/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContractRepository struct {
	db *pgxpool.Pool
}

func NewContractRepository(db *pgxpool.Pool) *ContractRepository {
	return &ContractRepository{db: db}
}

const contractSelectFields = `
	id, contract_no, client_id, otr, dp_percentage, dp_amount,
	principal_amount, tenor_months, interest_rate, interest_amount,
	total_amount, monthly_installment, contract_date, first_due_date,
	created_at, updated_at`

func scanContract(row interface {
	Scan(...any) error
}) (*model.Contract, error) {
	var c model.Contract
	var contractDate, firstDueDate time.Time
	err := row.Scan(
		&c.ID, &c.ContractNo, &c.ClientID, &c.OTR, &c.DPPercentage, &c.DPAmount,
		&c.PrincipalAmount, &c.TenorMonths, &c.InterestRate, &c.InterestAmount,
		&c.TotalAmount, &c.MonthlyInstallment, &contractDate, &firstDueDate,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.ContractDate = contractDate
	c.FirstDueDate = firstDueDate
	return &c, nil
}

// GetAll mengambil semua kontrak.
func (r *ContractRepository) GetAll(ctx context.Context) ([]model.Contract, error) {
	query := `SELECT ` + contractSelectFields + ` FROM ujicoba.contracts ORDER BY id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all contracts: %w", err)
	}
	defer rows.Close()

	var contracts []model.Contract
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		contracts = append(contracts, *c)
	}
	return contracts, rows.Err()
}

// GetByContractNo mengambil kontrak berdasarkan nomor kontrak.
func (r *ContractRepository) GetByContractNo(ctx context.Context, contractNo string) (*model.Contract, error) {
	query := `SELECT ` + contractSelectFields + ` FROM ujicoba.contracts WHERE contract_no = $1`

	c, err := scanContract(r.db.QueryRow(ctx, query, contractNo))
	if err != nil {
		return nil, fmt.Errorf("get contract by no: %w", err)
	}
	return c, nil
}

// GetByID mengambil kontrak berdasarkan ID.
func (r *ContractRepository) GetByID(ctx context.Context, id int64) (*model.Contract, error) {
	query := `SELECT ` + contractSelectFields + ` FROM ujicoba.contracts WHERE id = $1`

	c, err := scanContract(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("get contract by id: %w", err)
	}
	return c, nil
}

// InsertContract menyimpan kontrak baru ke DB dalam sebuah transaction.
func (r *ContractRepository) InsertContract(ctx context.Context, tx pgx.Tx, c *model.Contract) (int64, error) {
	query := `
		INSERT INTO ujicoba.contracts (
			contract_no, client_id, otr, dp_percentage, dp_amount,
			principal_amount, tenor_months, interest_rate, interest_amount,
			total_amount, monthly_installment, contract_date, first_due_date
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13
		) RETURNING id`

	var id int64
	err := tx.QueryRow(ctx, query,
		c.ContractNo, c.ClientID, c.OTR, c.DPPercentage, c.DPAmount,
		c.PrincipalAmount, c.TenorMonths, c.InterestRate, c.InterestAmount,
		c.TotalAmount, c.MonthlyInstallment, c.ContractDate, c.FirstDueDate,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert contract: %w", err)
	}
	return id, nil
}

// Update mengubah data non-finansial kontrak.
func (r *ContractRepository) Update(ctx context.Context, contractNo string, req *model.UpdateContractRequest) error {
	// Build dynamic update hanya untuk field yang di-isi
	if req.ClientID == 0 && req.ContractDate == "" && req.FirstDueDate == "" {
		return fmt.Errorf("no fields to update")
	}

	// Ambil data lama dulu untuk field yang tidak diubah
	existing, err := r.GetByContractNo(ctx, contractNo)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	clientID := existing.ClientID
	contractDate := existing.ContractDate
	firstDueDate := existing.FirstDueDate

	if req.ClientID != 0 {
		clientID = req.ClientID
	}
	if req.ContractDate != "" {
		contractDate, err = time.Parse("2006-01-02", req.ContractDate)
		if err != nil {
			return fmt.Errorf("invalid contract_date: %w", err)
		}
	}
	if req.FirstDueDate != "" {
		firstDueDate, err = time.Parse("2006-01-02", req.FirstDueDate)
		if err != nil {
			return fmt.Errorf("invalid first_due_date: %w", err)
		}
	}

	query := `
		UPDATE ujicoba.contracts
		SET client_id = $1, contract_date = $2, first_due_date = $3, updated_at = NOW()
		WHERE contract_no = $4`

	cmd, err := r.db.Exec(ctx, query, clientID, contractDate, firstDueDate, contractNo)
	if err != nil {
		return fmt.Errorf("update contract: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("contract not found")
	}
	return nil
}

// Delete menghapus kontrak berdasarkan contract_no.
func (r *ContractRepository) Delete(ctx context.Context, contractNo string) error {
	query := `DELETE FROM ujicoba.contracts WHERE contract_no = $1`

	cmd, err := r.db.Exec(ctx, query, contractNo)
	if err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("contract not found")
	}
	return nil
}

// BeginTx memulai transaction baru.
func (r *ContractRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// GetOverdueSummary calculates SUM(amount) where due_date <= input_date for a given contract.
func (r *ContractRepository) GetOverdueSummary(ctx context.Context, contractNo string, dateStr string) (*model.OverdueSummaryResponse, error) {
	inputDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	query := `
		SELECT
			c.contract_no,
			cl.client_code AS client_name,
			COALESCE(
				(SELECT SUM(i.amount)
				 FROM ujicoba.installments i
				 WHERE i.contract_id = c.id AND i.due_date <= $2::DATE),
				0
			)::BIGINT AS total_angsuran_jatuh_tempo
		FROM ujicoba.contracts c
		JOIN ujicoba.clients cl ON cl.id = c.client_id
		WHERE c.contract_no = $1`

	var resp model.OverdueSummaryResponse
	err = r.db.QueryRow(ctx, query, contractNo, inputDate).Scan(
		&resp.ContractNo,
		&resp.ClientName,
		&resp.TotalAngsuranJatuhTempo,
	)
	if err != nil {
		return nil, fmt.Errorf("get overdue summary: %w", err)
	}

	return &resp, nil
}

// GetPenalties returns installments where due_date <= input_date and no payment exists, along with calculated penalty.
func (r *ContractRepository) GetPenalties(ctx context.Context, contractNo string, dateStr string) ([]model.PenaltyDetailResponse, error) {
	inputDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	query := `
		SELECT
			c.contract_no,
			cl.client_code AS client_name,
			i.installment_no,
			i.due_date,
			i.amount
		FROM ujicoba.installments i
		JOIN ujicoba.contracts c ON c.id = i.contract_id
		JOIN ujicoba.clients cl ON cl.id = c.client_id
		LEFT JOIN ujicoba.payments p ON p.installment_id = i.id
		WHERE c.contract_no = $1
		  AND i.due_date <= $2::DATE
		  AND p.id IS NULL
		ORDER BY i.installment_no`

	rows, err := r.db.Query(ctx, query, contractNo, inputDate)
	if err != nil {
		return nil, fmt.Errorf("query penalties: %w", err)
	}
	defer rows.Close()

	var penalties []model.PenaltyDetailResponse
	for rows.Next() {
		var p model.PenaltyDetailResponse
		var dueDate time.Time
		var amount int64

		err := rows.Scan(
			&p.ContractNo,
			&p.ClientName,
			&p.InstallmentNo,
			&dueDate,
			&amount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan penalty row: %w", err)
		}

		// Cari tanggal akhir bulan dari dueDate
		firstDayNextMonth := time.Date(dueDate.Year(), dueDate.Month()+1, 1, 0, 0, 0, 0, dueDate.Location())
		endOfMonth := firstDayNextMonth.AddDate(0, 0, -1)

		// Calculate days late dari dueDate ke akhir bulan
		daysLate := int(endOfMonth.Sub(dueDate).Hours() / 24)
		if daysLate < 0 {
			daysLate = 0
		}

		// Calculate penalty (denda)
		// TOTAL_DENDA = amount * 0.001 * HARI_KETERLAMBATAN
		penalty := int64(float64(amount) * 0.001 * float64(daysLate))

		p.HariKeterlambatan = daysLate
		p.TotalDenda = penalty

		penalties = append(penalties, p)
	}

	return penalties, rows.Err()
}
