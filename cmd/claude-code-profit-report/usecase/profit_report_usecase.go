package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/entity"
	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/repository"
)

type ProfitReportUseCase interface {
	GenerateProfitReport(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) (*entity.ProfitReport, error)
}

type profitReportUseCaseImpl struct {
	salesRepo   repository.SalesRepository
	costRepo    repository.CostRepository
	companyRepo repository.CompanyRepository
}

func NewProfitReportUseCase(
	salesRepo repository.SalesRepository,
	costRepo repository.CostRepository,
	companyRepo repository.CompanyRepository,
) ProfitReportUseCase {
	return &profitReportUseCaseImpl{
		salesRepo:   salesRepo,
		costRepo:    costRepo,
		companyRepo: companyRepo,
	}
}

func (u *profitReportUseCaseImpl) GenerateProfitReport(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) (*entity.ProfitReport, error) {
	var companyName, warehouseName string

	if companyID > 0 {
		company, err := u.companyRepo.GetCompanyByID(ctx, companyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get company: %w", err)
		}
		companyName = company.Name
	} else {
		companyName = "全社"
	}

	if warehouseID > 0 {
		warehouse, err := u.companyRepo.GetWarehouseByID(ctx, warehouseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get warehouse: %w", err)
		}
		warehouseName = warehouse.Name
	} else {
		warehouseName = "全倉庫"
	}

	salesSummary, err := u.salesRepo.GetDailySummaryByPeriod(ctx, companyID, warehouseID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales summary: %w", err)
	}

	costSummary, err := u.costRepo.GetDailySummaryByPeriod(ctx, companyID, warehouseID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %w", err)
	}

	report := &entity.ProfitReport{
		CompanyID:     companyID,
		CompanyName:   companyName,
		WarehouseID:   warehouseID,
		WarehouseName: warehouseName,
		StartDate:     startDate,
		EndDate:       endDate,
	}

	var dailyReports []entity.DailyProfitReport
	var totalSales, totalCost float64

	log.Printf("DEBUG: UseCase - Processing date range: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	log.Printf("DEBUG: UseCase - Sales data count: %d\n", len(salesSummary))
	log.Printf("DEBUG: UseCase - Cost data count: %d\n", len(costSummary))
	
	// salesSummaryの内容を出力
	log.Println("DEBUG: UseCase - Sales Summary:")
	for date, amount := range salesSummary {
		log.Printf("  %s: %.2f\n", date.Format("2006-01-02"), amount)
	}
	
	// costSummaryの内容を出力
	log.Println("DEBUG: UseCase - Cost Summary:")
	for date, amount := range costSummary {
		log.Printf("  %s: %.2f\n", date.Format("2006-01-02"), amount)
	}
	
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		// 日付を正規化（時刻を00:00:00、ローカルタイムゾーンに設定）
		normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		sales, hasSales := salesSummary[normalizedDate]
		cost, hasCost := costSummary[normalizedDate]

		log.Printf("DEBUG: UseCase - Date: %s (normalized: %v), Sales: %.2f (exists: %v), Cost: %.2f (exists: %v)\n",
			date.Format("2006-01-02"), normalizedDate, sales, hasSales, cost, hasCost)

		dailyReport := entity.DailyProfitReport{
			Date:  date,
			Sales: sales,
			Cost:  cost,
		}
		dailyReport.CalculateGrossProfit()

		dailyReports = append(dailyReports, dailyReport)
		totalSales += sales
		totalCost += cost
	}
	
	log.Printf("DEBUG: UseCase - Total Sales: %.2f, Total Cost: %.2f\n", totalSales, totalCost)

	report.DailyReports = dailyReports
	report.TotalSales = totalSales
	report.TotalCost = totalCost
	report.CalculateGrossProfit()

	return report, nil
}