package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"ims-finance/internal/database"
	"ims-finance/internal/handler"
	"ims-finance/internal/repository"
	"ims-finance/internal/service"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	// Load .env if present
	_ = godotenv.Load("../.env")

	// Set to test mode
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database connection failed: %v. Make sure DATABASE_URL is set in .env or environment.", err)
	}

	// Repositories
	clientRepo := repository.NewClientRepository(pool)
	contractRepo := repository.NewContractRepository(pool)
	installmentRepo := repository.NewInstallmentRepository(pool)
	paymentRepo := repository.NewPaymentRepository(pool)

	// Services
	clientSvc := service.NewClientService(clientRepo)
	contractSvc := service.NewContractService(contractRepo, installmentRepo)
	installmentSvc := service.NewInstallmentService(installmentRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, installmentRepo)

	// Handlers
	loanHandler := handler.NewLoanHandler()
	clientHandler := handler.NewClientHandler(clientSvc)
	contractHandler := handler.NewContractHandler(contractSvc)
	installmentHandler := handler.NewInstallmentHandler(installmentSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/loans/calculate", loanHandler.Calculate)

		v1.GET("/clients", clientHandler.GetAll)
		v1.GET("/clients/:id", clientHandler.Get)
		v1.POST("/clients", clientHandler.Create)
		v1.PUT("/clients/:id", clientHandler.Update)
		v1.DELETE("/clients/:id", clientHandler.Delete)

		v1.GET("/contracts", contractHandler.GetAll)
		v1.GET("/contracts/:contractNo", contractHandler.Get)
		v1.POST("/contracts", contractHandler.Create)
		v1.PUT("/contracts/:contractNo", contractHandler.Update)
		v1.DELETE("/contracts/:contractNo", contractHandler.Delete)
		v1.GET("/contracts/:contractNo/overdue-summary", contractHandler.GetOverdueSummary)
		v1.GET("/contracts/:contractNo/penalties", contractHandler.GetPenalties)

		v1.GET("/contracts/:contractNo/installments", installmentHandler.GetInstallments)
		v1.GET("/contracts/:contractNo/overdue", installmentHandler.GetOverdue)
		v1.GET("/installments/:id", installmentHandler.GetByID)
		v1.PUT("/installments/:id", installmentHandler.Update)
		v1.DELETE("/installments/:id", installmentHandler.Delete)

		v1.GET("/payments", paymentHandler.GetAll)
		v1.GET("/payments/:id", paymentHandler.Get)
		v1.POST("/payments", paymentHandler.Create)
		v1.PUT("/payments/:id", paymentHandler.Update)
		v1.DELETE("/payments/:id", paymentHandler.Delete)

		v1.GET("/reports/client-overdue", installmentHandler.GetClientOverdueSummary)
	}

	return r
}

// 1. Test Loan Calculation Endpoint
func TestCalculateLoanEndpoint(t *testing.T) {
	r := setupTestRouter(t)
	if r == nil {
		return
	}

	reqBody, _ := json.Marshal(map[string]any{
		"otr":           240000000,
		"dp_percentage": 20,
		"tenor_months":  18,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/loans/calculate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify fields
	if int64(resp["dp"].(float64)) != 48000000 {
		t.Errorf("Expected dp 48000000, got %v", resp["dp"])
	}
	if int64(resp["principal"].(float64)) != 192000000 {
		t.Errorf("Expected principal 192000000, got %v", resp["principal"])
	}
	if resp["interest_rate"].(float64) != 0.14 {
		t.Errorf("Expected interest_rate 0.14, got %v", resp["interest_rate"])
	}
	if int64(resp["interest_amount"].(float64)) != 40320000 {
		t.Errorf("Expected interest_amount 40320000, got %v", resp["interest_amount"])
	}
	if int64(resp["total_amount"].(float64)) != 232320000 {
		t.Errorf("Expected total_amount 232320000, got %v", resp["total_amount"])
	}
	if int64(resp["monthly_installment"].(float64)) != 12907000 {
		t.Errorf("Expected monthly_installment 12907000, got %v", resp["monthly_installment"])
	}
}

// 2. Test Get Installments Endpoint (including the ?to=12 filter)
func TestGetInstallmentsEndpoint(t *testing.T) {
	r := setupTestRouter(t)
	if r == nil {
		return
	}

	// Test case: limit to 12 installments
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/contracts/AGR00001/installments?to=12", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Skipf("Contract AGR00001 seed data might not exist in database: status %d. Response: %s", w.Code, w.Body.String())
		return
	}

	var resp struct {
		ContractNo   string `json:"contract_no"`
		Installments []struct {
			InstallmentNo int    `json:"installment_no"`
			Amount        int64  `json:"amount"`
			DueDate       string `json:"due_date"`
		} `json:"installments"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.Installments) != 12 {
		t.Errorf("Expected exactly 12 installments, got %d", len(resp.Installments))
	}

	// Also verify standard behavior (all installments)
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/contracts/AGR00001/installments", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var resp2 struct {
		Installments []any `json:"installments"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &resp2)

	if len(resp2.Installments) != 18 {
		t.Errorf("Expected all 18 installments without filter, got %d", len(resp2.Installments))
	}
}

// 3. Test Get Overdue Endpoint (expecting specific denda/penalty for Pak Sugus on 2024-08-14)
func TestGetOverdueEndpoint(t *testing.T) {
	r := setupTestRouter(t)
	if r == nil {
		return
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/contracts/AGR00001/overdue?date=2024-08-14", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Skipf("Contract AGR00001 seed data might not exist in database: status %d. Response: %s", w.Code, w.Body.String())
		return
	}

	var resp struct {
		ContractNo   string `json:"contract_no"`
		AsOf         string `json:"as_of"`
		TotalOverdue int64  `json:"total_overdue"`
		TotalPenalty int64  `json:"total_penalty"`
		Installments []struct {
			InstallmentNo   int    `json:"installment_no"`
			DueDate         string `json:"due_date"`
			Amount          int64  `json:"amount"`
			PaidAmount      int64  `json:"paid_amount"`
			RemainingAmount int64  `json:"remaining_amount"`
			DaysLate        int    `json:"days_late"`
			Penalty         int64  `json:"penalty"`
		} `json:"installments"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify overdue calculations
	if resp.TotalOverdue != 25814000 {
		t.Errorf("Expected total_overdue 25814000, got %d", resp.TotalOverdue)
	}

	if resp.TotalPenalty != 903490 {
		t.Errorf("Expected total_penalty 903490, got %d", resp.TotalPenalty)
	}

	if len(resp.Installments) != 2 {
		t.Fatalf("Expected 2 overdue installments, got %d", len(resp.Installments))
	}

	// Verify Installment 6
	i6 := resp.Installments[0]
	if i6.InstallmentNo != 6 {
		t.Errorf("Expected first overdue installment to be No. 6, got %d", i6.InstallmentNo)
	}
	if i6.DaysLate != 50 {
		t.Errorf("Expected installment 6 days_late to be 50, got %d", i6.DaysLate)
	}
	if i6.Penalty != 645350 {
		t.Errorf("Expected installment 6 penalty to be 645350, got %d", i6.Penalty)
	}

	// Verify Installment 7
	i7 := resp.Installments[1]
	if i7.InstallmentNo != 7 {
		t.Errorf("Expected second overdue installment to be No. 7, got %d", i7.InstallmentNo)
	}
	if i7.DaysLate != 20 {
		t.Errorf("Expected installment 7 days_late to be 20, got %d", i7.DaysLate)
	}
	if i7.Penalty != 258140 {
		t.Errorf("Expected installment 7 penalty to be 258140, got %d", i7.Penalty)
	}
}

// 4. Test Client Overdue Summary Report Endpoint
func TestClientOverdueSummaryReport(t *testing.T) {
	r := setupTestRouter(t)
	if r == nil {
		return
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/reports/client-overdue?client_code=SUGUS&date=2024-08-14", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Skipf("Client SUGUS seed data might not exist in database: status %d. Response: %s", w.Code, w.Body.String())
		return
	}

	var resp struct {
		ClientCode   string `json:"client_code"`
		Name         string `json:"name"`
		AsOf         string `json:"as_of"`
		OverdueCount int    `json:"overdue_count"`
		TotalOverdue int64  `json:"total_overdue"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.ClientCode != "SUGUS" {
		t.Errorf("Expected client_code SUGUS, got %s", resp.ClientCode)
	}
	if resp.OverdueCount != 2 {
		t.Errorf("Expected overdue_count 2, got %d", resp.OverdueCount)
	}
	if resp.TotalOverdue != 25814000 {
		t.Errorf("Expected total_overdue 25814000, got %d", resp.TotalOverdue)
	}
}

// 5. Test GET /api/v1/contracts/:contractNo/overdue-summary
func TestGetOverdueSummaryEndpoint(t *testing.T) {
	r := setupTestRouter(t)
	if r == nil {
		return
	}

	// input date defaulting to 2024-08-14 is tested, wait, if date=2024-08-25 is queried it returns 103256000
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/contracts/AGR00001/overdue-summary?date=2024-08-25", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Skipf("Contract AGR00001 seed data might not exist in database: status %d. Response: %s", w.Code, w.Body.String())
		return
	}

	var resp struct {
		ContractNo              string `json:"contract_no"`
		ClientName              string `json:"client_name"`
		TotalAngsuranJatuhTempo int64  `json:"total_angsuran_jatuh_tempo"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.ContractNo != "AGR00001" {
		t.Errorf("Expected contract_no AGR00001, got %s", resp.ContractNo)
	}
	if resp.ClientName != "SUGUS" {
		t.Errorf("Expected client_name SUGUS, got %s", resp.ClientName)
	}
	if resp.TotalAngsuranJatuhTempo != 103256000 {
		t.Errorf("Expected total_angsuran_jatuh_tempo 103256000, got %d", resp.TotalAngsuranJatuhTempo)
	}

	// Also check default date behavior (2024-08-14 -> should be 7 installments = 90349000)
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/contracts/AGR00001/overdue-summary", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var resp2 struct {
		TotalAngsuranJatuhTempo int64 `json:"total_angsuran_jatuh_tempo"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &resp2)
	if resp2.TotalAngsuranJatuhTempo != 90349000 {
		t.Errorf("Expected default overdue-summary to be 90349000, got %d", resp2.TotalAngsuranJatuhTempo)
	}
}

// 6. Test GET /api/v1/contracts/:contractNo/penalties
func TestGetPenaltiesEndpoint(t *testing.T) {
	r := setupTestRouter(t)
	if r == nil {
		return
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/contracts/AGR00001/penalties?date=2024-08-14", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Skipf("Contract AGR00001 seed data might not exist in database: status %d. Response: %s", w.Code, w.Body.String())
		return
	}

	var resp []struct {
		ContractNo        string `json:"contract_no"`
		ClientName        string `json:"client_name"`
		InstallmentNo     int    `json:"installment_no"`
		HariKeterlambatan int    `json:"hari_keterlambatan"`
		TotalDenda        int64  `json:"total_denda"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("Expected exactly 2 penalty details, got %d", len(resp))
	}

	// Verify Installment 6
	i6 := resp[0]
	if i6.ContractNo != "AGR00001" || i6.ClientName != "SUGUS" || i6.InstallmentNo != 6 || i6.HariKeterlambatan != 50 || i6.TotalDenda != 645350 {
		t.Errorf("Unexpected penalty details for installment 6: %+v", i6)
	}

	// Verify Installment 7
	i7 := resp[1]
	if i7.ContractNo != "AGR00001" || i7.ClientName != "SUGUS" || i7.InstallmentNo != 7 || i7.HariKeterlambatan != 20 || i7.TotalDenda != 258140 {
		t.Errorf("Unexpected penalty details for installment 7: %+v", i7)
	}
}

