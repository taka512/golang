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

type salesRepositoryImpl struct {
	db *sql.DB
}

func NewSalesRepository(db *sql.DB) repository.SalesRepository {
	return &salesRepositoryImpl{db: db}
}

func (r *salesRepositoryImpl) GetDailyReportsByPeriod(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) ([]entity.SalesDailyReport, error) {
	query := `
		SELECT 
			sdr.id,
			sdr.company_id,
			sdr.warehouse_base_id,
			sdr.target_date,
			sdr.sales_account_title_id
		FROM sales_daily_reports sdr
		WHERE sdr.company_id = ?
			AND sdr.warehouse_base_id = ?
			AND sdr.target_date BETWEEN ? AND ?
		ORDER BY sdr.target_date, sdr.id
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, warehouseID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales daily reports: %w", err)
	}
	defer rows.Close()

	var reports []entity.SalesDailyReport
	for rows.Next() {
		var report entity.SalesDailyReport
		if err := rows.Scan(
			&report.ID,
			&report.CompanyID,
			&report.WarehouseBaseID,
			&report.TargetDate,
			&report.SalesAccountTitleID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sales daily report: %w", err)
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

func (r *salesRepositoryImpl) getReportItems(ctx context.Context, reportID uint64) ([]entity.SalesDailyReportItem, error) {
	query := `
		SELECT 
			id,
			sales_daily_report_id,
			size,
			quantity,
			price,
			amount
		FROM sales_daily_report_items
		WHERE sales_daily_report_id = ?
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales daily report items: %w", err)
	}
	defer rows.Close()

	var items []entity.SalesDailyReportItem
	for rows.Next() {
		var item entity.SalesDailyReportItem
		if err := rows.Scan(
			&item.ID,
			&item.SalesDailyReportID,
			&item.Size,
			&item.Quantity,
			&item.Price,
			&item.Amount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sales daily report item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}

func (r *salesRepositoryImpl) GetDailySummaryByPeriod(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) (map[time.Time]float64, error) {
	var query string
	var args []interface{}

	if companyID == 0 && warehouseID == 0 {
		query = `
			SELECT 
				sdr.target_date,
				COALESCE(SUM(sdri.amount), 0) as total_amount
			FROM sales_daily_reports sdr
			LEFT JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
			WHERE sdr.target_date BETWEEN ? AND ?
			GROUP BY sdr.target_date
			ORDER BY sdr.target_date
		`
		args = []interface{}{startDate, endDate}
	} else if companyID > 0 && warehouseID == 0 {
		query = `
			SELECT 
				sdr.target_date,
				COALESCE(SUM(sdri.amount), 0) as total_amount
			FROM sales_daily_reports sdr
			LEFT JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
			WHERE sdr.company_id = ?
				AND sdr.target_date BETWEEN ? AND ?
			GROUP BY sdr.target_date
			ORDER BY sdr.target_date
		`
		args = []interface{}{companyID, startDate, endDate}
	} else if companyID == 0 && warehouseID > 0 {
		query = `
			SELECT 
				sdr.target_date,
				COALESCE(SUM(sdri.amount), 0) as total_amount
			FROM sales_daily_reports sdr
			LEFT JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
			WHERE sdr.warehouse_base_id = ?
				AND sdr.target_date BETWEEN ? AND ?
			GROUP BY sdr.target_date
			ORDER BY sdr.target_date
		`
		args = []interface{}{warehouseID, startDate, endDate}
	} else {
		query = `
			SELECT 
				sdr.target_date,
				COALESCE(SUM(sdri.amount), 0) as total_amount
			FROM sales_daily_reports sdr
			LEFT JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
			WHERE sdr.company_id = ?
				AND sdr.warehouse_base_id = ?
				AND sdr.target_date BETWEEN ? AND ?
			GROUP BY sdr.target_date
			ORDER BY sdr.target_date
		`
		args = []interface{}{companyID, warehouseID, startDate, endDate}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales daily summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[time.Time]float64)
	rowCount := 0
	for rows.Next() {
		var date time.Time
		var amount float64
		if err := rows.Scan(&date, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan sales daily summary: %w", err)
		}
		// 日付部分のみを使用（時刻を00:00:00、ローカルタイムゾーンに正規化）
		normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		summary[normalizedDate] = amount
		rowCount++
		log.Printf("DEBUG: Sales - Date: %s (normalized: %v), Amount: %.2f\n", 
			date.Format("2006-01-02 15:04:05"), normalizedDate, amount)
	}
	log.Printf("DEBUG: Sales - Total rows: %d, Query args: %v\n", rowCount, args)

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return summary, nil
}