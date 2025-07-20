package repository

import (
	"context"
	"time"

	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/entity"
)

type SalesRepository interface {
	GetDailyReportsByPeriod(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) ([]entity.SalesDailyReport, error)
	GetDailySummaryByPeriod(ctx context.Context, companyID, warehouseID uint, startDate, endDate time.Time) (map[time.Time]float64, error)
}