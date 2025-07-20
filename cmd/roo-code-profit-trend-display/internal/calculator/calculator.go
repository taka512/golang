package calculator

import (
	"fmt"
	"sort"
	"time"

	"profit-trend-display/internal/models"
)

// ProfitCalculator handles profit calculation and trend analysis
type ProfitCalculator struct{}

// NewProfitCalculator creates a new profit calculator
func NewProfitCalculator() *ProfitCalculator {
	return &ProfitCalculator{}
}

// GroupByCompanyWarehouse groups profit data by company and warehouse combination
func (c *ProfitCalculator) GroupByCompanyWarehouse(data []models.ProfitData) map[string][]models.ProfitData {
	grouped := make(map[string][]models.ProfitData)

	for _, item := range data {
		key := fmt.Sprintf("%s-%s", item.CompanyName, item.WarehouseName)
		grouped[key] = append(grouped[key], item)
	}

	// Sort each group by date
	for key := range grouped {
		sort.Slice(grouped[key], func(i, j int) bool {
			return grouped[key][i].TargetDate.Before(grouped[key][j].TargetDate)
		})
	}

	return grouped
}

// CreateProfitTrends converts grouped data into profit trends with statistics
func (c *ProfitCalculator) CreateProfitTrends(groupedData map[string][]models.ProfitData) []models.ProfitTrend {
	var trends []models.ProfitTrend

	for _, data := range groupedData {
		if len(data) == 0 {
			continue
		}

		trend := models.ProfitTrend{
			CompanyID:       data[0].CompanyID,
			CompanyName:     data[0].CompanyName,
			WarehouseBaseID: data[0].WarehouseBaseID,
			WarehouseName:   data[0].WarehouseName,
			Data:            data,
			Stats:           c.calculateStats(data),
		}

		trends = append(trends, trend)
	}

	// Sort trends by company name, then warehouse name
	sort.Slice(trends, func(i, j int) bool {
		if trends[i].CompanyName != trends[j].CompanyName {
			return trends[i].CompanyName < trends[j].CompanyName
		}
		return trends[i].WarehouseName < trends[j].WarehouseName
	})

	return trends
}

// calculateStats computes statistical information for profit data
func (c *ProfitCalculator) calculateStats(data []models.ProfitData) models.ProfitStats {
	if len(data) == 0 {
		return models.ProfitStats{}
	}

	stats := models.ProfitStats{
		MaxProfit: data[0].ProfitAmount,
		MinProfit: data[0].ProfitAmount,
		MaxDate:   data[0].TargetDate,
		MinDate:   data[0].TargetDate,
		DaysCount: len(data),
	}

	var totalProfit float64

	for _, item := range data {
		totalProfit += item.ProfitAmount

		if item.ProfitAmount > stats.MaxProfit {
			stats.MaxProfit = item.ProfitAmount
			stats.MaxDate = item.TargetDate
		}

		if item.ProfitAmount < stats.MinProfit {
			stats.MinProfit = item.ProfitAmount
			stats.MinDate = item.TargetDate
		}
	}

	stats.TotalProfit = totalProfit
	stats.AvgProfit = totalProfit / float64(len(data))

	return stats
}

// FillMissingDates fills in missing dates with zero profit for complete trend visualization
func (c *ProfitCalculator) FillMissingDates(data []models.ProfitData, startDate, endDate time.Time) []models.ProfitData {
	if len(data) == 0 {
		return data
	}

	// Create a map for quick lookup
	dataMap := make(map[string]models.ProfitData)
	for _, item := range data {
		key := item.TargetDate.Format("2006-01-02")
		dataMap[key] = item
	}

	// Fill missing dates
	var result []models.ProfitData
	current := startDate

	// Get template data for company/warehouse info
	template := data[0]

	for current.Before(endDate) || current.Equal(endDate) {
		key := current.Format("2006-01-02")
		
		if item, exists := dataMap[key]; exists {
			result = append(result, item)
		} else {
			// Create zero profit entry for missing date
			zeroEntry := models.ProfitData{
				CompanyID:       template.CompanyID,
				CompanyName:     template.CompanyName,
				WarehouseBaseID: template.WarehouseBaseID,
				WarehouseName:   template.WarehouseName,
				TargetDate:      current,
				SalesAmount:     0,
				CostAmount:      0,
				ProfitAmount:    0,
			}
			result = append(result, zeroEntry)
		}

		current = current.AddDate(0, 0, 1)
	}

	return result
}

// GetDateRange calculates the appropriate date range for the last N days
func (c *ProfitCalculator) GetDateRange(days int) (time.Time, time.Time) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// Truncate to date only (remove time component)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	return startDate, endDate
}