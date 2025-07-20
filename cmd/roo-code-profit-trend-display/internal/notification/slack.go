package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"profit-trend-display/internal/models"
)

// SlackNotifier handles Slack notifications
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier creates a new Slack notifier instance
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsEnabled returns whether Slack notifications are enabled
func (s *SlackNotifier) IsEnabled() bool {
	return s.webhookURL != ""
}

// SendProfitSummary sends a profit summary notification to Slack
func (s *SlackNotifier) SendProfitSummary(trends []models.ProfitTrend, period int) error {
	if !s.IsEnabled() {
		return fmt.Errorf("slack notifications not enabled")
	}

	message := s.formatProfitMessage(trends, period)
	return s.sendMessage(message)
}

// SendError sends an error notification to Slack
func (s *SlackNotifier) SendError(err error) error {
	if !s.IsEnabled() {
		return fmt.Errorf("slack notifications not enabled")
	}

	message := models.SlackMessage{
		Text: "❌ 粗利分析エラー",
		Attachments: []models.Attachment{
			{
				Color: "danger",
				Title: "エラー詳細",
				Text:  fmt.Sprintf("エラー内容: %v", err),
				Fields: []models.Field{
					{
						Title: "発生時刻",
						Value: time.Now().Format("2006-01-02 15:04:05"),
						Short: true,
					},
				},
			},
		},
	}

	return s.sendMessage(message)
}

// formatProfitMessage formats profit trends data into a Slack message
func (s *SlackNotifier) formatProfitMessage(trends []models.ProfitTrend, period int) models.SlackMessage {
	// Calculate overall statistics
	var totalProfit, maxProfit, minProfit float64
	var orgCount int
	var maxDate, minDate time.Time
	var allData []models.ProfitData

	if len(trends) > 0 {
		// Initialize with first trend's data
		firstTrend := trends[0]
		maxProfit = firstTrend.Stats.MaxProfit
		minProfit = firstTrend.Stats.MinProfit
		maxDate = firstTrend.Stats.MaxDate
		minDate = firstTrend.Stats.MinDate

		// Aggregate all trends
		for _, trend := range trends {
			totalProfit += trend.Stats.TotalProfit
			orgCount++

			if trend.Stats.MaxProfit > maxProfit {
				maxProfit = trend.Stats.MaxProfit
				maxDate = trend.Stats.MaxDate
			}
			if trend.Stats.MinProfit < minProfit {
				minProfit = trend.Stats.MinProfit
				minDate = trend.Stats.MinDate
			}

			// Collect all data for average calculation
			allData = append(allData, trend.Data...)
		}
	}

	// Calculate average profit
	avgProfit := float64(0)
	if len(allData) > 0 {
		totalDays := len(allData)
		avgProfit = totalProfit / float64(totalDays)
	}

	// Create main message
	message := models.SlackMessage{
		Text: fmt.Sprintf("📊 粗利推移分析結果 (過去%d日間)", period),
		Attachments: []models.Attachment{
			{
				Color: "good",
				Title: "全体統計",
				Fields: []models.Field{
					{
						Title: "合計粗利",
						Value: s.formatCurrency(totalProfit),
						Short: true,
					},
					{
						Title: "平均粗利",
						Value: s.formatCurrency(avgProfit),
						Short: true,
					},
					{
						Title: "最大粗利",
						Value: fmt.Sprintf("%s (%s)", s.formatCurrency(maxProfit), maxDate.Format("01/02")),
						Short: true,
					},
					{
						Title: "最小粗利",
						Value: fmt.Sprintf("%s (%s)", s.formatCurrency(minProfit), minDate.Format("01/02")),
						Short: true,
					},
					{
						Title: "対象組織数",
						Value: strconv.Itoa(orgCount),
						Short: true,
					},
					{
						Title: "実行日時",
						Value: time.Now().Format("2006-01-02 15:04:05"),
						Short: true,
					},
				},
			},
		},
	}

	// Add top organizations if there are multiple trends
	if len(trends) > 1 {
		topOrgs := s.getTopOrganizations(trends, 3)
		if len(topOrgs) > 0 {
			fields := make([]models.Field, 0, len(topOrgs))
			for i, org := range topOrgs {
				fields = append(fields, models.Field{
					Title: fmt.Sprintf("%d位", i+1),
					Value: fmt.Sprintf("%s: %s", org.Name, s.formatCurrency(org.TotalProfit)),
					Short: false,
				})
			}

			topOrgAttachment := models.Attachment{
				Color:  "warning",
				Title:  "組織別トップ3",
				Fields: fields,
			}
			message.Attachments = append(message.Attachments, topOrgAttachment)
		}
	}

	return message
}

// OrganizationSummary represents a summary for an organization
type OrganizationSummary struct {
	Name        string
	TotalProfit float64
}

// getTopOrganizations returns the top N organizations by total profit
func (s *SlackNotifier) getTopOrganizations(trends []models.ProfitTrend, limit int) []OrganizationSummary {
	orgs := make([]OrganizationSummary, 0, len(trends))

	for _, trend := range trends {
		orgName := fmt.Sprintf("%s - %s", trend.CompanyName, trend.WarehouseName)
		orgs = append(orgs, OrganizationSummary{
			Name:        orgName,
			TotalProfit: trend.Stats.TotalProfit,
		})
	}

	// Sort by total profit in descending order
	sort.Slice(orgs, func(i, j int) bool {
		return orgs[i].TotalProfit > orgs[j].TotalProfit
	})

	// Return top N organizations
	if len(orgs) > limit {
		orgs = orgs[:limit]
	}

	return orgs
}

// formatCurrency formats a float64 value as Japanese Yen currency
func (s *SlackNotifier) formatCurrency(amount float64) string {
	// Convert to integer for display
	intAmount := int64(amount)
	
	// Add thousand separators
	str := strconv.FormatInt(intAmount, 10)
	
	// Add commas for thousands
	if len(str) > 3 {
		var result strings.Builder
		for i, digit := range str {
			if i > 0 && (len(str)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		str = result.String()
	}
	
	return str + "円"
}

// sendMessage sends a message to Slack via webhook
func (s *SlackNotifier) sendMessage(message models.SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}