package cli

import (
	"fmt"
	"strings"

	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/entity"
)

type Formatter interface {
	FormatProfitReport(report *entity.ProfitReport) string
}

type TextFormatter struct{}

func NewTextFormatter() Formatter {
	return &TextFormatter{}
}

func (f *TextFormatter) FormatProfitReport(report *entity.ProfitReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("売上・コスト・粗利レポート\n"))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("=", 60)))
	sb.WriteString(fmt.Sprintf("会社: %s (ID: %d)\n", report.CompanyName, report.CompanyID))
	sb.WriteString(fmt.Sprintf("倉庫: %s (ID: %d)\n", report.WarehouseName, report.WarehouseID))
	sb.WriteString(fmt.Sprintf("期間: %s ~ %s\n", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("%s\n\n", strings.Repeat("=", 60)))

	sb.WriteString(fmt.Sprintf("【期間合計】\n"))
	sb.WriteString(fmt.Sprintf("売上高: %s\n", formatCurrency(report.TotalSales)))
	sb.WriteString(fmt.Sprintf("コスト: %s\n", formatCurrency(report.TotalCost)))
	sb.WriteString(fmt.Sprintf("粗利益: %s\n", formatCurrency(report.GrossProfit)))
	sb.WriteString(fmt.Sprintf("粗利率: %.2f%%\n\n", report.GrossProfitRate))

	sb.WriteString(fmt.Sprintf("【日別詳細】\n"))
	sb.WriteString(fmt.Sprintf("%-12s %15s %15s %15s %8s\n", "日付", "売上", "コスト", "粗利", "粗利率"))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 75)))

	for _, daily := range report.DailyReports {
		sb.WriteString(fmt.Sprintf("%-12s %15s %15s %15s %7.2f%%\n",
			daily.Date.Format("2006-01-02"),
			formatCurrency(daily.Sales),
			formatCurrency(daily.Cost),
			formatCurrency(daily.GrossProfit),
			daily.GrossProfitRate,
		))
	}

	return sb.String()
}

func formatCurrency(amount float64) string {
	// 整数部分を取得
	intPart := int64(amount)
	
	// 負の数の場合の処理
	negative := false
	if intPart < 0 {
		negative = true
		intPart = -intPart
	}
	
	// 3桁ごとにカンマを挿入
	str := fmt.Sprintf("%d", intPart)
	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	
	if negative {
		return fmt.Sprintf("¥-%s", result)
	}
	return fmt.Sprintf("¥%s", result)
}