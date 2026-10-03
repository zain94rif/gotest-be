package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ims-finance/internal/service"
)

type LoanHandler struct{}

func NewLoanHandler() *LoanHandler {
	return &LoanHandler{}
}

type calculateLoanRequest struct {
	OTR          int64   `json:"otr" binding:"required,gt=0"`
	DPPercentage float64 `json:"dp_percentage" binding:"min=0,max=100"`
	TenorMonths  int     `json:"tenor_months" binding:"required,gt=0"`
}

// Calculate godoc
// POST /api/v1/loans/calculate
func (h *LoanHandler) Calculate(c *gin.Context) {
	var req calculateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := service.CalculateLoan(req.OTR, req.DPPercentage, req.TenorMonths)
	c.JSON(http.StatusOK, result)
}
