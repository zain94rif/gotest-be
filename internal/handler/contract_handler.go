package handler

import (
	"net/http"

	"ims-finance/internal/model"
	"ims-finance/internal/service"

	"github.com/gin-gonic/gin"
)

type ContractHandler struct {
	contractService *service.ContractService
}

func NewContractHandler(s *service.ContractService) *ContractHandler {
	return &ContractHandler{contractService: s}
}

// GetAll godoc
// GET /api/v1/contracts
func (h *ContractHandler) GetAll(c *gin.Context) {
	contracts, err := h.contractService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contracts)
}

// Get godoc
// GET /api/v1/contracts/:contractNo
func (h *ContractHandler) Get(c *gin.Context) {
	contractNo := c.Param("contractNo")

	contract, err := h.contractService.GetByContractNo(c.Request.Context(), contractNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	c.JSON(http.StatusOK, contract)
}

// Create godoc
// POST /api/v1/contracts
func (h *ContractHandler) Create(c *gin.Context) {
	var req model.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contract, err := h.contractService.CreateContract(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, contract)
}

// Update godoc
// PUT /api/v1/contracts/:contractNo
func (h *ContractHandler) Update(c *gin.Context) {
	contractNo := c.Param("contractNo")

	var req model.UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.contractService.Update(c.Request.Context(), contractNo, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "contract updated"})
}

// Delete godoc
// DELETE /api/v1/contracts/:contractNo
func (h *ContractHandler) Delete(c *gin.Context) {
	contractNo := c.Param("contractNo")

	if err := h.contractService.Delete(c.Request.Context(), contractNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "contract deleted"})
}

// GetOverdueSummary godoc
// GET /api/v1/contracts/:contractNo/overdue-summary?date=2024-08-14
func (h *ContractHandler) GetOverdueSummary(c *gin.Context) {
	contractNo := c.Param("contractNo")
	dateStr := c.Query("date")

	resp, err := h.contractService.GetOverdueSummary(c.Request.Context(), contractNo, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPenalties godoc
// GET /api/v1/contracts/:contractNo/penalties?date=2024-08-14
func (h *ContractHandler) GetPenalties(c *gin.Context) {
	contractNo := c.Param("contractNo")
	dateStr := c.Query("date")

	resp, err := h.contractService.GetPenalties(c.Request.Context(), contractNo, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp == nil {
		resp = []model.PenaltyDetailResponse{}
	}

	c.JSON(http.StatusOK, resp)
}
