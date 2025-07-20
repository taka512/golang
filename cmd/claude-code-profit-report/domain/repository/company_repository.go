package repository

import (
	"context"
)

type Company struct {
	ID   uint
	Name string
	Code string
}

type WarehouseBase struct {
	ID   uint
	Name string
	Code string
}

type CompanyRepository interface {
	GetCompanyByID(ctx context.Context, id uint) (*Company, error)
	GetWarehouseByID(ctx context.Context, id uint) (*WarehouseBase, error)
	GetAllCompanies(ctx context.Context) ([]Company, error)
	GetWarehousesByCompanyID(ctx context.Context, companyID uint) ([]WarehouseBase, error)
}