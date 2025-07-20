package chart

import (
	"fmt"
	"strings"
	"time"

	"profit-trend-display/internal/models"
)

// TextChart handles text-based chart rendering
type TextChart struct {
	config models.ChartConfig
}

// NewTextChart creates a new text chart renderer
func NewTextChart(config models.ChartConfig) *TextChart {
	// Set default values if not provided
	if config.Width == 0 {
		config.Width = 60
	}
	if config.Height == 0 {
		config.Height = 15
	}

	return &TextChart{config: config}
}

// RenderProfitTrend renders a profit trend as ASCII art
func (c *TextChart) RenderProfitTrend(trend models.ProfitTrend) string {
	if len(trend.Data) == 0 {
		return "No data available for this trend"
	}

	var result strings.Builder

	// Header
	header := fmt.Sprintf("[%s - %s] 粗利推移 (過去%d日間)", 
		trend.CompanyName, trend.WarehouseName, len(trend.Data))
	result.WriteString(header + "\n")
	result.WriteString(strings.Repeat("=", len(header)) + "\n\n")

	// Prepare data for chart
	chartData := c.prepareChartData(trend.Data)
	
	// Render the chart
	chartLines := c.renderChart(chartData)
	result.WriteString(strings.Join(chartLines, "\n") + "\n")

	// Date axis
	result.WriteString(c.renderDateAxis(trend.Data) + "\n\n")

	// Statistics
	if c.config.ShowStats {
		result.WriteString(c.renderStats(trend.Stats))
	}

	result.WriteString("\n")
	return result.String()
}

// prepareChartData converts profit data to chart points
func (c *TextChart) prepareChartData(data []models.ProfitData) []models.ChartPoint {
	var points []models.ChartPoint

	// Find min and max values for scaling
	if len(data) == 0 {
		return points
	}

	minVal := data[0].ProfitAmount
	maxVal := data[0].ProfitAmount

	for _, item := range data {
		if item.ProfitAmount < minVal {
			minVal = item.ProfitAmount
		}
		if item.ProfitAmount > maxVal {
			maxVal = item.ProfitAmount
		}
	}

	// Update config with actual min/max if not set
	if c.config.MinValue == 0 && c.config.MaxValue == 0 {
		c.config.MinValue = minVal
		c.config.MaxValue = maxVal
		
		// Add some padding
		range_ := maxVal - minVal
		if range_ > 0 {
			padding := range_ * 0.1
			c.config.MinValue -= padding
			c.config.MaxValue += padding
		} else {
			// Handle case where all values are the same
			if maxVal == 0 {
				c.config.MinValue = -100
				c.config.MaxValue = 100
			} else {
				c.config.MinValue = maxVal * 0.9
				c.config.MaxValue = maxVal * 1.1
			}
		}
	}

	// Convert to chart points
	for _, item := range data {
		symbol := "●"
		if item.ProfitAmount == 0 {
			symbol = "○"
		} else if item.ProfitAmount < 0 {
			symbol = "▼"
		}

		points = append(points, models.ChartPoint{
			Date:   item.TargetDate,
			Value:  item.ProfitAmount,
			Symbol: symbol,
		})
	}

	return points
}

// renderChart creates the main chart visualization
func (c *TextChart) renderChart(points []models.ChartPoint) []string {
	lines := make([]string, c.config.Height)
	
	// Calculate value range
	valueRange := c.config.MaxValue - c.config.MinValue
	if valueRange == 0 {
		valueRange = 1 // Prevent division by zero
	}

	// Y-axis labels and chart area
	for row := 0; row < c.config.Height; row++ {
		// Calculate the value for this row
		rowValue := c.config.MaxValue - (float64(row)/float64(c.config.Height-1))*valueRange
		
		// Format the Y-axis label
		yLabel := fmt.Sprintf("%7.0f", rowValue)
		
		// Chart line
		chartLine := make([]rune, c.config.Width)
		for i := range chartLine {
			chartLine[i] = ' '
		}

		// Plot points
		for i, point := range points {
			// Calculate X position
			var xPos int
			if len(points) > 1 {
				xPos = int(float64(i) / float64(len(points)-1) * float64(c.config.Width-1))
			} else {
				xPos = c.config.Width / 2 // Center single point
			}
			
			if xPos >= c.config.Width {
				xPos = c.config.Width - 1
			}
			if xPos < 0 {
				xPos = 0
			}

			// Calculate Y position for this point
			pointY := int((c.config.MaxValue - point.Value) / valueRange * float64(c.config.Height-1))
			
			// If this point should be on the current row
			if pointY == row && xPos < len(chartLine) {
				chartLine[xPos] = []rune(point.Symbol)[0]
			}
		}

		// Grid lines
		if c.config.ShowGrid && row > 0 && row < c.config.Height-1 {
			for i := 0; i < c.config.Width; i++ {
				if chartLine[i] == ' ' && i%10 == 0 {
					chartLine[i] = '┊'
				}
			}
		}

		// Border
		if row == 0 {
			lines[row] = yLabel + " ┬" + string(chartLine)
		} else if row == c.config.Height-1 {
			lines[row] = yLabel + " └" + string(chartLine)
		} else {
			lines[row] = yLabel + " ┤" + string(chartLine)
		}
	}

	return lines
}

// renderDateAxis creates the date axis at the bottom
func (c *TextChart) renderDateAxis(data []models.ProfitData) string {
	if len(data) == 0 {
		return ""
	}

	// Create date labels
	axisLine := make([]rune, c.config.Width+9) // +9 for Y-axis space
	for i := range axisLine {
		axisLine[i] = ' '
	}

	// Add axis markers
	copy(axisLine[8:], []rune("─"))
	for i := 9; i < c.config.Width+9; i++ {
		axisLine[i] = '─'
	}

	// Create a slice for date labels positioning
	dateLabelLine := make([]rune, c.config.Width+9)
	for i := range dateLabelLine {
		dateLabelLine[i] = ' '
	}

	// Show date markers every few days
	step := len(data) / 8 // Show about 8 date markers
	if step < 1 {
		step = 1
	}

	for i := 0; i < len(data); i += step {
		var xPos int
		if len(data) > 1 {
			xPos = int(float64(i) / float64(len(data)-1) * float64(c.config.Width-1))
		} else {
			xPos = c.config.Width / 2 // Center single point
		}
		
		if xPos+9 < len(axisLine) {
			axisLine[xPos+9] = '┬'
		}
		
		// Add date label
		dateStr := data[i].TargetDate.Format("01/02")
		labelStartPos := xPos + 9 - len(dateStr)/2
		if labelStartPos >= 9 && labelStartPos+len(dateStr) <= len(dateLabelLine) {
			copy(dateLabelLine[labelStartPos:labelStartPos+len(dateStr)], []rune(dateStr))
		}
	}

	dateLabels := string(dateLabelLine)

	return string(axisLine) + "\n" + dateLabels
}

// renderStats creates a statistics summary
func (c *TextChart) renderStats(stats models.ProfitStats) string {
	var result strings.Builder

	result.WriteString("統計情報:\n")
	result.WriteString(fmt.Sprintf("  最大粗利: %10.0f (%s)\n", stats.MaxProfit, stats.MaxDate.Format("01/02")))
	result.WriteString(fmt.Sprintf("  最小粗利: %10.0f (%s)\n", stats.MinProfit, stats.MinDate.Format("01/02")))
	result.WriteString(fmt.Sprintf("  平均粗利: %10.0f\n", stats.AvgProfit))
	result.WriteString(fmt.Sprintf("  合計粗利: %10.0f\n", stats.TotalProfit))
	result.WriteString(fmt.Sprintf("  データ日数: %d日\n", stats.DaysCount))

	return result.String()
}

// RenderSummary renders a summary of all trends
func (c *TextChart) RenderSummary(trends []models.ProfitTrend) string {
	var result strings.Builder

	result.WriteString("=== 粗利推移サマリー ===\n\n")

	if len(trends) == 0 {
		result.WriteString("表示するデータがありません。\n")
		return result.String()
	}

	// Overall statistics
	var totalProfit, maxProfit, minProfit float64
	var maxDate, minDate time.Time
	totalDays := 0
	
	for i, trend := range trends {
		if i == 0 {
			maxProfit = trend.Stats.MaxProfit
			minProfit = trend.Stats.MinProfit
			maxDate = trend.Stats.MaxDate
			minDate = trend.Stats.MinDate
		} else {
			if trend.Stats.MaxProfit > maxProfit {
				maxProfit = trend.Stats.MaxProfit
				maxDate = trend.Stats.MaxDate
			}
			if trend.Stats.MinProfit < minProfit {
				minProfit = trend.Stats.MinProfit
				minDate = trend.Stats.MinDate
			}
		}
		
		totalProfit += trend.Stats.TotalProfit
		totalDays += trend.Stats.DaysCount
	}

	avgProfit := 0.0
	if totalDays > 0 {
		avgProfit = totalProfit / float64(totalDays)
	}

	result.WriteString("全体統計:\n")
	result.WriteString(fmt.Sprintf("  合計粗利: %10.0f\n", totalProfit))
	result.WriteString(fmt.Sprintf("  平均粗利: %10.0f\n", avgProfit))
	result.WriteString(fmt.Sprintf("  最大粗利: %10.0f (%s)\n", maxProfit, maxDate.Format("01/02")))
	result.WriteString(fmt.Sprintf("  最小粗利: %10.0f (%s)\n", minProfit, minDate.Format("01/02")))
	result.WriteString(fmt.Sprintf("  対象組織数: %d\n", len(trends)))
	result.WriteString("\n")

	// Individual trend summaries
	result.WriteString("組織別サマリー:\n")
	for _, trend := range trends {
		result.WriteString(fmt.Sprintf("  %s - %s: 合計=%.0f, 平均=%.0f\n",
			trend.CompanyName, trend.WarehouseName,
			trend.Stats.TotalProfit, trend.Stats.AvgProfit))
	}

	return result.String()
}