package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"ims-finance/internal/model"
)

type ClientRepository struct {
	db *pgxpool.Pool
}

func NewClientRepository(db *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{db: db}
}

const clientSelectFields = `id, client_code, name, created_at, updated_at`

// GetAll mengambil semua client.
func (r *ClientRepository) GetAll(ctx context.Context) ([]model.Client, error) {
	query := `SELECT ` + clientSelectFields + ` FROM ujicoba.clients ORDER BY id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all clients: %w", err)
	}
	defer rows.Close()

	var clients []model.Client
	for rows.Next() {
		var c model.Client
		if err := rows.Scan(&c.ID, &c.ClientCode, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

// GetByID mengambil client berdasarkan ID.
func (r *ClientRepository) GetByID(ctx context.Context, id int64) (*model.Client, error) {
	query := `SELECT ` + clientSelectFields + ` FROM ujicoba.clients WHERE id = $1`

	var c model.Client
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.ClientCode, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get client by id: %w", err)
	}
	return &c, nil
}

// Insert menyimpan client baru dan mengembalikan ID-nya.
func (r *ClientRepository) Insert(ctx context.Context, req *model.CreateClientRequest) (int64, error) {
	query := `
		INSERT INTO ujicoba.clients (client_code, name)
		VALUES ($1, $2)
		RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, query, req.ClientCode, req.Name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert client: %w", err)
	}
	return id, nil
}

// Update mengubah nama client berdasarkan ID.
func (r *ClientRepository) Update(ctx context.Context, id int64, req *model.UpdateClientRequest) error {
	query := `
		UPDATE ujicoba.clients
		SET name = $1, updated_at = NOW()
		WHERE id = $2`

	cmd, err := r.db.Exec(ctx, query, req.Name, id)
	if err != nil {
		return fmt.Errorf("update client: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("client not found")
	}
	return nil
}

// Delete menghapus client berdasarkan ID.
func (r *ClientRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM ujicoba.clients WHERE id = $1`

	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("client not found")
	}
	return nil
}
