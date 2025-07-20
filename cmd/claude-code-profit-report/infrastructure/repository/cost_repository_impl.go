package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/entity"
	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/repository"
)

type costRepositoryImpl struct {
	db *sql.DB
}

func NewCostRepository(db *sql.DB) repository.CostRepository {
	return &costRepositoryImpl{db: db}
}

func (r *costRepositoryImpl) GetDailyReportsByPeriod(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) ([]entity.CostDailyReport, error) {
	query := `
		SELECT 
			cdr.id,
			cdr.company_id,
			cdr.warehouse_base_id,
			cdr.target_date,
			cdr.cost_account_title_id
		FROM cost_daily_reports cdr
		WHERE cdr.company_id = ?
			AND cdr.warehouse_base_id = ?
			AND cdr.target_date BETWEEN ? AND ?
		ORDER BY cdr.target_date, cdr.id
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, warehouseID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query cost daily reports: %w", err)
	}
	defer rows.Close()

	var reports []entity.CostDailyReport
	for rows.Next() {
		var report entity.CostDailyReport
		if err := rows.Scan(
			&report.ID,
			&report.CompanyID,
			&report.WarehouseBaseID,
			&report.TargetDate,
			&report.CostAccountTitleID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost daily report: %w", err)
		}

		items, err := r.getReportItems(ctx, report.ID)
		if err != nil {
			return nil, err
		}
		report.Items = items

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return reports, nil
}

func (r *costRepositoryImpl) getReportItems(ctx context.Context, reportID uint64) ([]entity.CostDailyReportItem, error) {
	query := `
		SELECT 
			id,
			cost_daily_report_id,
			size,
			quantity,
			cost_price,
			cost_amount
		FROM cost_daily_report_items
		WHERE cost_daily_report_id = ?
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cost daily report items: %w", err)
	}
	defer rows.Close()

	var items []entity.CostDailyReportItem
	for rows.Next() {
		var item entity.CostDailyReportItem
		if err := rows.Scan(
			&item.ID,
			&item.CostDailyReportID,
			&item.Size,
			&item.Quantity,
			&item.CostPrice,
			&item.CostAmount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost daily report item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}

func (r *costRepositoryImpl) GetDailySummaryByPeriod(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) (map[time.Time]float64, error) {
	var query string
	var args []interface{}

	if companyID == 0 && warehouseID == 0 {
		query = `
			SELECT 
				cdr.target_date,
				COALESCE(SUM(cdri.cost_amount), 0) as total_amount
			FROM cost_daily_reports cdr
			LEFT JOIN cost_daily_report_items cdri ON cdr.id = cdri.cost_daily_report_id
			WHERE cdr.target_date BETWEEN ? AND ?
			GROUP BY cdr.target_date
			ORDER BY cdr.target_date
		`
		args = []interface{}{startDate, endDate}
	} else if companyID > 0 && warehouseID == 0 {
		query = `
			SELECT 
				cdr.target_date,
				COALESCE(SUM(cdri.cost_amount), 0) as total_amount
			FROM cost_daily_reports cdr
			LEFT JOIN cost_daily_report_items cdri ON cdr.id = cdri.cost_daily_report_id
			WHERE cdr.company_id = ?
				AND cdr.target_date BETWEEN ? AND ?
			GROUP BY cdr.target_date
			ORDER BY cdr.target_date
		`
		args = []interface{}{companyID, startDate, endDate}
	} else if companyID == 0 && warehouseID > 0 {
		query = `
			SELECT 
				cdr.target_date,
				COALESCE(SUM(cdri.cost_amount), 0) as total_amount
			FROM cost_daily_reports cdr
			LEFT JOIN cost_daily_report_items cdri ON cdr.id = cdri.cost_daily_report_id
			WHERE cdr.warehouse_base_id = ?
				AND cdr.target_date BETWEEN ? AND ?
			GROUP BY cdr.target_date
			ORDER BY cdr.target_date
		`
		args = []interface{}{warehouseID, startDate, endDate}
	} else {
		query = `
			SELECT 
				cdr.target_date,
				COALESCE(SUM(cdri.cost_amount), 0) as total_amount
			FROM cost_daily_reports cdr
			LEFT JOIN cost_daily_report_items cdri ON cdr.id = cdri.cost_daily_report_id
			WHERE cdr.company_id = ?
				AND cdr.warehouse_base_id = ?
				AND cdr.target_date BETWEEN ? AND ?
			GROUP BY cdr.target_date
			ORDER BY cdr.target_date
		`
		args = []interface{}{companyID, warehouseID, startDate, endDate}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cost daily summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[time.Time]float64)
	rowCount := 0
	for rows.Next() {
		var date time.Time
		var amount float64
		if err := rows.Scan(&date, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan cost daily summary: %w", err)
		}
		// 日付部分のみを使用（時刻を00:00:00、ローカルタイムゾーンに正規化）
		normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		summary[normalizedDate] = amount
		rowCount++
		log.Printf("DEBUG: Cost - Date: %s (normalized: %v), Amount: %.2f\n", 
			date.Format("2006-01-02 15:04:05"), normalizedDate, amount)
	}
	log.Printf("DEBUG: Cost - Total rows: %d, Query args: %v\n", rowCount, args)

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return summary, nil
}