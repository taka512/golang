package entity

import (
	"time"
)

type CostDailyReport struct {
	ID                 uint64
	CompanyID          uint
	WarehouseBaseID    uint
	TargetDate         time.Time
	CostAccountTitleID uint
	Items              []CostDailyReportItem
}

type CostDailyReportItem struct {
	ID                uint64
	CostDailyReportID uint64
	Size              *string
	Quantity          int
	CostPrice         float64
	CostAmount        float64
}

func (c *CostDailyReport) CalculateTotalAmount() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.CostAmount
	}
	return total
}