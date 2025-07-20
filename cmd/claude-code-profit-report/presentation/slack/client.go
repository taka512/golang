package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/taka512/golang/cmd/claude-code-profit-report/domain/entity"
)

type Client struct {
	webhookURL string
	httpClient *http.Client
}

func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Message struct {
	Text   string       `json:"text"`
	Blocks []Block      `json:"blocks,omitempty"`
}

type Block struct {
	Type string      `json:"type"`
	Text *TextObject `json:"text,omitempty"`
}

type TextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (c *Client) SendProfitReport(report *entity.ProfitReport) error {
	message := c.formatProfitReport(report)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) formatProfitReport(report *entity.ProfitReport) Message {
	blocks := []Block{
		{
			Type: "section",
			Text: &TextObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*売上・コスト・粗利レポート*\n期間: %s ~ %s",
					report.StartDate.Format("2006-01-02"),
					report.EndDate.Format("2006-01-02")),
			},
		},
		{
			Type: "section",
			Text: &TextObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*会社:* %s\n*倉庫:* %s",
					report.CompanyName,
					report.WarehouseName),
			},
		},
		{
			Type: "divider",
		},
		{
			Type: "section",
			Text: &TextObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*【期間合計】*\n売上高: ¥%,.0f\nコスト: ¥%,.0f\n粗利益: ¥%,.0f\n粗利率: %.2f%%",
					report.TotalSales,
					report.TotalCost,
					report.GrossProfit,
					report.GrossProfitRate),
			},
		},
	}

	// 日別詳細（最新10日分のみ表示）
	dailyText := "*【日別詳細（最新10日分）】*\n```\n"
	dailyText += fmt.Sprintf("%-10s %12s %12s %12s %8s\n", "日付", "売上", "コスト", "粗利", "粗利率")
	dailyText += "─────────────────────────────────────────────────────\n"

	startIdx := len(report.DailyReports) - 10
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(report.DailyReports); i++ {
		daily := report.DailyReports[i]
		dailyText += fmt.Sprintf("%-10s %12.0f %12.0f %12.0f %7.2f%%\n",
			daily.Date.Format("01-02"),
			daily.Sales,
			daily.Cost,
			daily.GrossProfit,
			daily.GrossProfitRate,
		)
	}
	dailyText += "```"

	blocks = append(blocks, Block{
		Type: "section",
		Text: &TextObject{
			Type: "mrkdwn",
			Text: dailyText,
		},
	})

	return Message{
		Text:   fmt.Sprintf("売上・コスト・粗利レポート (%s ~ %s)", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02")),
		Blocks: blocks,
	}
}