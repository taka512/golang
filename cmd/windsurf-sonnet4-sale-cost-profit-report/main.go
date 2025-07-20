package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// SaleCostProfitReport represents a row in the profit report
type SaleCostProfitReport struct {
	CompanyID       int     `json:"company_id"`
	CompanyName     string  `json:"company_name"`
	WarehouseBaseID int     `json:"warehouse_base_id"`
	WarehouseName   string  `json:"warehouse_name"`
	TargetDate      string  `json:"target_date"`
	SalesAmount     float64 `json:"sales_amount"`
	CostAmount      float64 `json:"cost_amount"`
	ProfitAmount    float64 `json:"profit_amount"`
	ProfitMargin    float64 `json:"profit_margin"`
}

func main() {
	// Database connection string - adjust as needed
	dsn := "root:mypass@tcp(mysql.local:3306)/sample_mysql?parseTime=true"
	
	// Parse command line arguments for date range
	var startDate, endDate string
	if len(os.Args) >= 3 {
		startDate = os.Args[1]
		endDate = os.Args[2]
	} else {
		// Default to current month
		now := time.Now()
		startDate = fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
		endDate = now.Format("2006-01-02")
	}

	fmt.Printf("Generating Sale Cost Profit Report from %s to %s\n", startDate, endDate)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	reports, err := generateProfitReport(db, startDate, endDate)
	if err != nil {
		log.Fatal("Failed to generate report:", err)
	}

	// Output to CSV file
	filename := fmt.Sprintf("sale_cost_profit_report_%s_to_%s.csv", startDate, endDate)
	if err := writeCSVReport(reports, filename); err != nil {
		log.Fatal("Failed to write CSV report:", err)
	}

	// Output summary to console
	printSummary(reports)
	
	fmt.Printf("Report saved to: %s\n", filename)
}

func generateProfitReport(db *sql.DB, startDate, endDate string) ([]SaleCostProfitReport, error) {
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

	rows, err := db.Query(query, startDate, endDate, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var reports []SaleCostProfitReport
	for rows.Next() {
		var report SaleCostProfitReport
		var targetDate sql.NullString
		
		err := rows.Scan(
			&report.CompanyID,
			&report.CompanyName,
			&report.WarehouseBaseID,
			&report.WarehouseName,
			&targetDate,
			&report.SalesAmount,
			&report.CostAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if targetDate.Valid {
			report.TargetDate = targetDate.String
		}

		// Calculate profit and margin
		report.ProfitAmount = report.SalesAmount - report.CostAmount
		if report.SalesAmount > 0 {
			report.ProfitMargin = (report.ProfitAmount / report.SalesAmount) * 100
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return reports, nil
}

func writeCSVReport(reports []SaleCostProfitReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Company ID", "Company Name", "Warehouse ID", "Warehouse Name",
		"Target Date", "Sales Amount", "Cost Amount", "Profit Amount", "Profit Margin (%)",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for _, report := range reports {
		row := []string{
			strconv.Itoa(report.CompanyID),
			report.CompanyName,
			strconv.Itoa(report.WarehouseBaseID),
			report.WarehouseName,
			report.TargetDate,
			fmt.Sprintf("%.3f", report.SalesAmount),
			fmt.Sprintf("%.3f", report.CostAmount),
			fmt.Sprintf("%.3f", report.ProfitAmount),
			fmt.Sprintf("%.2f", report.ProfitMargin),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

func printSummary(reports []SaleCostProfitReport) {
	if len(reports) == 0 {
		fmt.Println("No data found for the specified date range.")
		return
	}

	var totalSales, totalCosts, totalProfit float64
	companyTotals := make(map[string]struct {
		sales, costs, profit float64
	})

	for _, report := range reports {
		totalSales += report.SalesAmount
		totalCosts += report.CostAmount
		totalProfit += report.ProfitAmount

		key := report.CompanyName
		totals := companyTotals[key]
		totals.sales += report.SalesAmount
		totals.costs += report.CostAmount
		totals.profit += report.ProfitAmount
		companyTotals[key] = totals
	}

	fmt.Println("\n=== PROFIT REPORT SUMMARY ===")
	fmt.Printf("Total Records: %d\n", len(reports))
	fmt.Printf("Total Sales: %.3f\n", totalSales)
	fmt.Printf("Total Costs: %.3f\n", totalCosts)
	fmt.Printf("Total Profit: %.3f\n", totalProfit)
	
	if totalSales > 0 {
		fmt.Printf("Overall Profit Margin: %.2f%%\n", (totalProfit/totalSales)*100)
	}

	fmt.Println("\n=== BY COMPANY ===")
	for company, totals := range companyTotals {
		margin := 0.0
		if totals.sales > 0 {
			margin = (totals.profit / totals.sales) * 100
		}
		fmt.Printf("%s: Sales=%.3f, Costs=%.3f, Profit=%.3f (%.2f%%)\n",
			company, totals.sales, totals.costs, totals.profit, margin)
	}
}
