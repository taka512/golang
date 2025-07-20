package database

import (
	"database/sql"
	"fmt"
	"time"

	"profit-trend-display/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

// ProfitRepository handles database operations for profit data
type ProfitRepository struct {
	db *sql.DB
}

// NewProfitRepository creates a new profit repository
func NewProfitRepository(dsn string) (*ProfitRepository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &ProfitRepository{db: db}, nil
}

// Close closes the database connection
func (r *ProfitRepository) Close() error {
	return r.db.Close()
}

// GetProfitTrendsForPeriod retrieves profit data for the specified period
func (r *ProfitRepository) GetProfitTrendsForPeriod(startDate, endDate time.Time) ([]models.ProfitData, error) {
	query := `
		SELECT 
			c.id as company_id,
			c.name as company_name,
			wb.id as warehouse_base_id,
			wb.name as warehouse_name,
			DATE(COALESCE(sdr.target_date, cdr.target_date)) as target_date,
			COALESCE(SUM(sdri.amount), 0) as sales_amount,
			COALESCE(SUM(cdri.cost_amount), 0) as cost_amount
		FROM companies c
		CROSS JOIN warehouse_bases wb
		LEFT JOIN sales_daily_reports sdr ON c.id = sdr.company_id 
			AND wb.id = sdr.warehouse_base_id 
			AND sdr.target_date BETWEEN ? AND ?
		LEFT JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
		LEFT JOIN cost_daily_reports cdr ON c.id = cdr.company_id 
			AND wb.id = cdr.warehouse_base_id 
			AND cdr.target_date BETWEEN ? AND ?
		LEFT JOIN cost_daily_report_items cdri ON cdr.id = cdri.cost_daily_report_id
		WHERE (sdr.target_date IS NOT NULL OR cdr.target_date IS NOT NULL)
		GROUP BY c.id, c.name, wb.id, wb.name, DATE(COALESCE(sdr.target_date, cdr.target_date))
		ORDER BY c.name, wb.name, DATE(COALESCE(sdr.target_date, cdr.target_date))
	`

	rows, err := r.db.Query(query, startDate, endDate, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var profitData []models.ProfitData
	for rows.Next() {
		var data models.ProfitData
		var targetDate sql.NullTime

		err := rows.Scan(
			&data.CompanyID,
			&data.CompanyName,
			&data.WarehouseBaseID,
			&data.WarehouseName,
			&targetDate,
			&data.SalesAmount,
			&data.CostAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if targetDate.Valid {
			data.TargetDate = targetDate.Time
		}

		// Calculate profit amount
		data.ProfitAmount = data.SalesAmount - data.CostAmount

		profitData = append(profitData, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return profitData, nil
}

// GetCompaniesWithWarehouses retrieves all companies with their warehouse bases
func (r *ProfitRepository) GetCompaniesWithWarehouses() (map[string][]models.ProfitData, error) {
	query := `
		SELECT 
			c.id as company_id,
			c.name as company_name,
			wb.id as warehouse_base_id,
			wb.name as warehouse_name
		FROM companies c
		CROSS JOIN warehouse_bases wb
		ORDER BY c.name, wb.name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	companiesMap := make(map[string][]models.ProfitData)
	for rows.Next() {
		var data models.ProfitData

		err := rows.Scan(
			&data.CompanyID,
			&data.CompanyName,
			&data.WarehouseBaseID,
			&data.WarehouseName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		key := fmt.Sprintf("%s-%s", data.CompanyName, data.WarehouseName)
		companiesMap[key] = append(companiesMap[key], data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return companiesMap, nil
}