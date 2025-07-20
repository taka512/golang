package entity

import (
	"time"
)

type ProfitReport struct {
	CompanyID      uint
	CompanyName    string
	WarehouseID    uint
	WarehouseName  string
	StartDate      time.Time
	EndDate        time.Time
	TotalSales     float64
	TotalCost      float64
	GrossProfit    float64
	GrossProfitRate float64
	DailyReports   []DailyProfitReport
}

type DailyProfitReport struct {
	Date           time.Time
	Sales          float64
	Cost           float64
	GrossProfit    float64
	GrossProfitRate float64
}

func (p *ProfitReport) CalculateGrossProfit() {
	p.GrossProfit = p.TotalSales - p.TotalCost
	if p.TotalSales > 0 {
		p.GrossProfitRate = (p.GrossProfit / p.TotalSales) * 100
	}
}

func (d *DailyProfitReport) CalculateGrossProfit() {
	d.GrossProfit = d.Sales - d.Cost
	if d.Sales > 0 {
		d.GrossProfitRate = (d.GrossProfit / d.Sales) * 100
	}
}