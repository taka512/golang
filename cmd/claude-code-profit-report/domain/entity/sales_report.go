package entity

import (
	"time"
)

type SalesDailyReport struct {
	ID                  uint64
	CompanyID           uint
	WarehouseBaseID     uint
	TargetDate          time.Time
	SalesAccountTitleID uint
	Items               []SalesDailyReportItem
}

type SalesDailyReportItem struct {
	ID                  uint64
	SalesDailyReportID  uint64
	Size                *string
	Quantity            int
	Price               float64
	Amount              float64
}

func (s *SalesDailyReport) CalculateTotalAmount() float64 {
	var total float64
	for _, item := range s.Items {
		total += item.Amount
	}
	return total
}