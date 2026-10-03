package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ims-finance/internal/model"
	"ims-finance/internal/service"
)

type InstallmentHandler struct {
	installmentService *service.InstallmentService
}

func NewInstallmentHandler(s *service.InstallmentService) *InstallmentHandler {
	return &InstallmentHandler{installmentService: s}
}

// GetInstallments godoc
// GET /api/v1/contracts/:contractNo/installments?to=12
// Parameter opsional: to=N (tampilkan sampai installment ke-N)
func (h *InstallmentHandler) GetInstallments(c *gin.Context) {
	contractNo := c.Param("contractNo")

	upTo := 0
	if toStr := c.Query("to"); toStr != "" {
		val, err := strconv.Atoi(toStr)
		if err != nil || val <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "param 'to' must be a positive integer"})
			return
		}
		upTo = val
	}

	installments, err := h.installmentService.GetInstallments(c.Request.Context(), contractNo, upTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(installments) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found or no installments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contract_no":  contractNo,
		"installments": installments,
	})
}

// GetByID godoc
// GET /api/v1/installments/:id
func (h *InstallmentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	inst, err := h.installmentService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "installment not found"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// Update godoc
// PUT /api/v1/installments/:id
func (h *InstallmentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req model.UpdateInstallmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.installmentService.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "installment updated"})
}

// Delete godoc
// DELETE /api/v1/installments/:id
func (h *InstallmentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.installmentService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "installment deleted"})
}

// GetOverdue godoc
// GET /api/v1/contracts/:contractNo/overdue?date=2024-08-14
func (h *InstallmentHandler) GetOverdue(c *gin.Context) {
	contractNo := c.Param("contractNo")
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query param 'date' is required"})
		return
	}

	result, err := h.installmentService.GetOverdue(c.Request.Context(), contractNo, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetClientOverdueSummary godoc
// GET /api/v1/reports/client-overdue?client_code=SUGUS&date=2024-08-14
func (h *InstallmentHandler) GetClientOverdueSummary(c *gin.Context) {
	clientCode := c.Query("client_code")
	date := c.Query("date")

	if clientCode == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query params 'client_code' and 'date' are required"})
		return
	}

	summary, err := h.installmentService.GetClientOverdueSummary(c.Request.Context(), clientCode, date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}
