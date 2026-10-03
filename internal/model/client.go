package model

import "time"

type Client struct {
	ID         int64     `json:"id"`
	ClientCode string    `json:"client_code"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateClientRequest struct {
	ClientCode string `json:"client_code" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

type UpdateClientRequest struct {
	Name string `json:"name" binding:"required"`
}

// ClientOverdueSummary adalah hasil agregat overdue seluruh kontrak milik satu client.
type ClientOverdueSummary struct {
	ClientCode   string `json:"client_code"`
	Name         string `json:"name"`
	AsOf         string `json:"as_of"`
	OverdueCount int    `json:"overdue_count"`
	TotalOverdue int64  `json:"total_overdue"`
}
