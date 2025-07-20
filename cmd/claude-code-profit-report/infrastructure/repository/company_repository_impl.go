package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/repository"
)

type companyRepositoryImpl struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) repository.CompanyRepository {
	return &companyRepositoryImpl{db: db}
}

func (r *companyRepositoryImpl) GetCompanyByID(ctx context.Context, id uint) (*repository.Company, error) {
	query := `
		SELECT id, name, code
		FROM companies
		WHERE id = ?
	`

	var company repository.Company
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&company.ID,
		&company.Name,
		&company.Code,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("company not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}

func (r *companyRepositoryImpl) GetWarehouseByID(ctx context.Context, id uint) (*repository.WarehouseBase, error) {
	query := `
		SELECT id, name, code
		FROM warehouse_bases
		WHERE id = ?
	`

	var warehouse repository.WarehouseBase
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&warehouse.ID,
		&warehouse.Name,
		&warehouse.Code,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("warehouse not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	return &warehouse, nil
}

func (r *companyRepositoryImpl) GetAllCompanies(ctx context.Context) ([]repository.Company, error) {
	query := `
		SELECT id, name, code
		FROM companies
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query companies: %w", err)
	}
	defer rows.Close()

	var companies []repository.Company
	for rows.Next() {
		var company repository.Company
		if err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.Code,
		); err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return companies, nil
}

func (r *companyRepositoryImpl) GetWarehousesByCompanyID(ctx context.Context, companyID uint) ([]repository.WarehouseBase, error) {
	// warehouse_basesテーブルにはcompany_idがないため、全倉庫を返す
	query := `
		SELECT id, name, code
		FROM warehouse_bases
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []repository.WarehouseBase
	for rows.Next() {
		var warehouse repository.WarehouseBase
		if err := rows.Scan(
			&warehouse.ID,
			&warehouse.Name,
			&warehouse.Code,
		); err != nil {
			return nil, fmt.Errorf("failed to scan warehouse: %w", err)
		}
		warehouses = append(warehouses, warehouse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return warehouses, nil
}