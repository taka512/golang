package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type ShipmentReport struct {
	Date          string
	Company       string
	Warehouse     string
	SalesAmount   float64
	CostAmount    float64
	Profit        float64
	ProfitMargin  float64
	SalesQuantity int
	CostQuantity  int
}

type Summary struct {
	TotalSales      float64
	TotalCost       float64
	TotalProfit     float64
	AvgProfitMargin float64
	TotalSalesQty   int
	TotalCostQty    int
	Days            int
}

func main() {
	// データベース接続
	db, err := sql.Open("mysql", "root:mypass@(mysql.local:3306)/sample_mysql?parseTime=true")
	if err != nil {
		log.Fatal("データベース接続エラー:", err)
	}
	defer db.Close()

	// 出荷実績レポートを取得
	reports, err := getShipmentReports(db)
	if err != nil {
		log.Fatal("レポート取得エラー:", err)
	}

	if len(reports) == 0 {
		fmt.Println("データが見つかりませんでした。")
		return
	}

	// レポート表示
	printHeader()
	for _, report := range reports {
		printReportRow(report)
	}
	printSeparator()

	// サマリー計算と表示
	summary := calculateSummary(reports)
	printSummary(summary)
}

func getShipmentReports(db *sql.DB) ([]ShipmentReport, error) {
	query := `
SELECT 
s.target_date,
c.name as company_name,
w.name as warehouse_name,
COALESCE(sales_summary.total_amount, 0) as sales_amount,
COALESCE(cost_summary.total_amount, 0) as cost_amount,
COALESCE(sales_summary.total_quantity, 0) as sales_quantity,
COALESCE(cost_summary.total_quantity, 0) as cost_quantity
FROM (
SELECT DISTINCT target_date, company_id, warehouse_base_id
FROM sales_daily_reports 
WHERE sales_account_title_id = (SELECT id FROM sales_account_titles WHERE code = 'shipment')
UNION
SELECT DISTINCT target_date, company_id, warehouse_base_id  
FROM cost_daily_reports
WHERE cost_account_title_id = (SELECT id FROM cost_account_titles WHERE code = 'shipment')
) s
JOIN companies c ON s.company_id = c.id
JOIN warehouse_bases w ON s.warehouse_base_id = w.id
LEFT JOIN (
SELECT 
sdr.target_date,
sdr.company_id,
sdr.warehouse_base_id,
SUM(si.amount) as total_amount,
SUM(si.quantity) as total_quantity
FROM sales_daily_reports sdr
JOIN sales_daily_report_items si ON sdr.id = si.sales_daily_report_id
WHERE sdr.sales_account_title_id = (SELECT id FROM sales_account_titles WHERE code = 'shipment')
GROUP BY sdr.target_date, sdr.company_id, sdr.warehouse_base_id
) sales_summary ON s.target_date = sales_summary.target_date 
AND s.company_id = sales_summary.company_id 
AND s.warehouse_base_id = sales_summary.warehouse_base_id
LEFT JOIN (
SELECT 
cdr.target_date,
cdr.company_id,
cdr.warehouse_base_id,
SUM(ci.cost_amount) as total_amount,
SUM(ci.quantity) as total_quantity
FROM cost_daily_reports cdr
JOIN cost_daily_report_items ci ON cdr.id = ci.cost_daily_report_id
WHERE cdr.cost_account_title_id = (SELECT id FROM cost_account_titles WHERE code = 'shipment')
GROUP BY cdr.target_date, cdr.company_id, cdr.warehouse_base_id
) cost_summary ON s.target_date = cost_summary.target_date 
AND s.company_id = cost_summary.company_id 
AND s.warehouse_base_id = cost_summary.warehouse_base_id
ORDER BY s.target_date DESC, c.name, w.name
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []ShipmentReport
	for rows.Next() {
		var report ShipmentReport
		var salesQty, costQty sql.NullInt32

		err := rows.Scan(
			&report.Date,
			&report.Company,
			&report.Warehouse,
			&report.SalesAmount,
			&report.CostAmount,
			&salesQty,
			&costQty,
		)
		if err != nil {
			return nil, err
		}

		// Null値の処理
		if salesQty.Valid {
			report.SalesQuantity = int(salesQty.Int32)
		}
		if costQty.Valid {
			report.CostQuantity = int(costQty.Int32)
		}

		// 粗利計算
		report.Profit = report.SalesAmount - report.CostAmount

		// 粗利率計算（売上が0でない場合のみ）
		if report.SalesAmount > 0 {
			report.ProfitMargin = (report.Profit / report.SalesAmount) * 100
		}

		reports = append(reports, report)
	}

	return reports, nil
}

func printHeader() {
	fmt.Println("=== 出荷粗利・実績レポート ===")
	fmt.Println()
	fmt.Printf("%-12s %-15s %-10s %12s %12s %12s %8s %8s %8s\n",
		"日付", "会社名", "倉庫", "売上金額", "原価金額", "粗利", "粗利率%", "売上数量", "原価数量")
	printSeparator()
}

func printReportRow(report ShipmentReport) {
	profitMarginStr := ""
	if report.SalesAmount > 0 {
		profitMarginStr = fmt.Sprintf("%.1f%%", report.ProfitMargin)
	} else {
		profitMarginStr = "N/A"
	}

	fmt.Printf("%-12s %-15s %-10s %12s %12s %12s %8s %8d %8d\n",
		report.Date,
		truncateString(report.Company, 15),
		truncateString(report.Warehouse, 10),
		formatCurrency(report.SalesAmount),
		formatCurrency(report.CostAmount),
		formatCurrency(report.Profit),
		profitMarginStr,
		report.SalesQuantity,
		report.CostQuantity)
}

func printSeparator() {
	fmt.Println(strings.Repeat("-", 120))
}

func calculateSummary(reports []ShipmentReport) Summary {
	var summary Summary
	dateMap := make(map[string]bool)

	for _, report := range reports {
		summary.TotalSales += report.SalesAmount
		summary.TotalCost += report.CostAmount
		summary.TotalProfit += report.Profit
		summary.TotalSalesQty += report.SalesQuantity
		summary.TotalCostQty += report.CostQuantity
		dateMap[report.Date] = true
	}

	summary.Days = len(dateMap)

	// 平均粗利率計算
	if summary.TotalSales > 0 {
		summary.AvgProfitMargin = (summary.TotalProfit / summary.TotalSales) * 100
	}

	return summary
}

func printSummary(summary Summary) {
	fmt.Println("=== サマリー ===")
	fmt.Printf("対象日数: %d日\n", summary.Days)
	fmt.Printf("総売上金額: %s\n", formatCurrency(summary.TotalSales))
	fmt.Printf("総原価金額: %s\n", formatCurrency(summary.TotalCost))
	fmt.Printf("総粗利: %s\n", formatCurrency(summary.TotalProfit))
	fmt.Printf("平均粗利率: %.1f%%\n", summary.AvgProfitMargin)
	fmt.Printf("総売上数量: %d\n", summary.TotalSalesQty)
	fmt.Printf("総原価数量: %d\n", summary.TotalCostQty)
	fmt.Println()

	// 日別平均
	if summary.Days > 0 {
		fmt.Println("=== 日別平均 ===")
		fmt.Printf("平均日売上: %s\n", formatCurrency(summary.TotalSales/float64(summary.Days)))
		fmt.Printf("平均日原価: %s\n", formatCurrency(summary.TotalCost/float64(summary.Days)))
		fmt.Printf("平均日粗利: %s\n", formatCurrency(summary.TotalProfit/float64(summary.Days)))
		fmt.Printf("平均日売上数量: %.0f\n", float64(summary.TotalSalesQty)/float64(summary.Days))
		fmt.Printf("平均日原価数量: %.0f\n", float64(summary.TotalCostQty)/float64(summary.Days))
	}
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("¥%s", addCommas(fmt.Sprintf("%.0f", amount)))
}

func addCommas(s string) string {
	if len(s) <= 3 {
		return s
	}

	var result []string
	for i, char := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = append(result, ",")
		}
		result = append(result, string(char))
	}

	return reverse(strings.Join(result, ""))
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
