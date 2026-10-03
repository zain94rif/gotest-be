package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"ims-finance/internal/database"
	"ims-finance/internal/handler"
	"ims-finance/internal/repository"
	"ims-finance/internal/service"
)

func main() {
	// Load .env jika ada
	_ = godotenv.Load()

	// Database connection
	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()
	log.Println("database connected")

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

	// Router
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		// Loans
		v1.POST("/loans/calculate", loanHandler.Calculate)

		// Clients CRUD
		v1.GET("/clients", clientHandler.GetAll)
		v1.GET("/clients/:id", clientHandler.Get)
		v1.POST("/clients", clientHandler.Create)
		v1.PUT("/clients/:id", clientHandler.Update)
		v1.DELETE("/clients/:id", clientHandler.Delete)

		// Contracts CRUD & Installments/Overdue
		v1.GET("/contracts", contractHandler.GetAll)
		v1.GET("/contracts/:contractNo", contractHandler.Get)
		v1.POST("/contracts", contractHandler.Create)
		v1.PUT("/contracts/:contractNo", contractHandler.Update)
		v1.DELETE("/contracts/:contractNo", contractHandler.Delete)
		v1.GET("/contracts/:contractNo/overdue-summary", contractHandler.GetOverdueSummary)
		v1.GET("/contracts/:contractNo/penalties", contractHandler.GetPenalties)

		// Installments within contracts & general CRUD
		v1.GET("/contracts/:contractNo/installments", installmentHandler.GetInstallments)
		v1.GET("/contracts/:contractNo/overdue", installmentHandler.GetOverdue)
		v1.GET("/installments/:id", installmentHandler.GetByID)
		v1.PUT("/installments/:id", installmentHandler.Update)
		v1.DELETE("/installments/:id", installmentHandler.Delete)

		// Payments CRUD
		v1.GET("/payments", paymentHandler.GetAll)
		v1.GET("/payments/:id", paymentHandler.Get)
		v1.POST("/payments", paymentHandler.Create)
		v1.PUT("/payments/:id", paymentHandler.Update)
		v1.DELETE("/payments/:id", paymentHandler.Delete)

		// Reports
		v1.GET("/reports/client-overdue", installmentHandler.GetClientOverdueSummary)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9531"
	}

	log.Printf("server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

