package config

import (
	"database/sql"

	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/repository"
	infraRepo "github.com/taka512/golang/cmd/claude-code-profit-report/infrastructure/repository"
	"github.com/taka512/golang/cmd/claude-code-profit-report/usecase"
)

type Container struct {
	DB                  *sql.DB
	SalesRepository     repository.SalesRepository
	CostRepository      repository.CostRepository
	CompanyRepository   repository.CompanyRepository
	ProfitReportUseCase usecase.ProfitReportUseCase
}

func NewContainer(db *sql.DB) *Container {
	salesRepo := infraRepo.NewSalesRepository(db)
	costRepo := infraRepo.NewCostRepository(db)
	companyRepo := infraRepo.NewCompanyRepository(db)
	
	profitReportUseCase := usecase.NewProfitReportUseCase(salesRepo, costRepo, companyRepo)

	return &Container{
		DB:                  db,
		SalesRepository:     salesRepo,
		CostRepository:      costRepo,
		CompanyRepository:   companyRepo,
		ProfitReportUseCase: profitReportUseCase,
	}
}