package models

import "time"

// ProfitData represents daily profit data for a company-warehouse combination
type ProfitData struct {
	CompanyID       int       `json:"company_id"`
	CompanyName     string    `json:"company_name"`
	WarehouseBaseID int       `json:"warehouse_base_id"`
	WarehouseName   string    `json:"warehouse_name"`
	TargetDate      time.Time `json:"target_date"`
	SalesAmount     float64   `json:"sales_amount"`
	CostAmount      float64   `json:"cost_amount"`
	ProfitAmount    float64   `json:"profit_amount"`
}

// ProfitTrend represents a series of profit data for trend analysis
type ProfitTrend struct {
	CompanyID       int           `json:"company_id"`
	CompanyName     string        `json:"company_name"`
	WarehouseBaseID int           `json:"warehouse_base_id"`
	WarehouseName   string        `json:"warehouse_name"`
	Data            []ProfitData  `json:"data"`
	Stats           ProfitStats   `json:"stats"`
}

// ProfitStats contains statistical information about profit trends
type ProfitStats struct {
	MaxProfit     float64   `json:"max_profit"`
	MinProfit     float64   `json:"min_profit"`
	AvgProfit     float64   `json:"avg_profit"`
	TotalProfit   float64   `json:"total_profit"`
	MaxDate       time.Time `json:"max_date"`
	MinDate       time.Time `json:"min_date"`
	DaysCount     int       `json:"days_count"`
}

// ChartPoint represents a point in the text-based chart
type ChartPoint struct {
	Date   time.Time `json:"date"`
	Value  float64   `json:"value"`
	Symbol string    `json:"symbol"`
}

// ChartConfig contains configuration for chart rendering
type ChartConfig struct {
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	MinValue  float64 `json:"min_value"`
	MaxValue  float64 `json:"max_value"`
	ShowGrid  bool    `json:"show_grid"`
	ShowStats bool    `json:"show_stats"`
}

// NotificationConfig contains configuration for notifications
type NotificationConfig struct {
	SlackEnabled    bool   `json:"slack_enabled"`
	SlackWebhookURL string `json:"slack_webhook_url"`
}

// SlackMessage represents a Slack message structure
type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color  string  `json:"color"`
	Title  string  `json:"title"`
	Text   string  `json:"text,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}

// Field represents a field in a Slack attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}