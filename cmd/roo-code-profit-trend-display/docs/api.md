# API仕様書

## 1. 概要

`roo-code-profit-trend-display` は現在コマンドラインインターフェース（CLI）のみを提供していますが、将来的なAPI化を見据えた内部アーキテクチャを採用しています。

## 2. CLI API仕様

### 2.1 コマンド構文

```bash
profit-trend-display [OPTIONS] [DAYS]
```

### 2.2 引数・オプション仕様

#### 2.2.1 基本オプション

| オプション | 型 | デフォルト値 | 必須 | 説明 |
|------------|-----|-------------|------|------|
| `-dsn` | string | `root:mypass@tcp(mysql.local:3306)/sample_mysql?parseTime=true` | No | データベース接続文字列 |
| `-days` | int | 30 | No | 分析対象日数（1-365） |
| `-help` | bool | false | No | ヘルプメッセージ表示 |

#### 2.2.2 表示制御オプション

| オプション | 型 | デフォルト値 | 必須 | 説明 |
|------------|-----|-------------|------|------|
| `-width` | int | 60 | No | チャート幅（20-200） |
| `-height` | int | 15 | No | チャート高さ（5-50） |
| `-grid` | bool | true | No | グリッド線表示フラグ |
| `-stats` | bool | true | No | 統計情報表示フラグ |
| `-summary` | bool | false | No | サマリーのみ表示フラグ |

#### 2.2.3 位置引数

| 引数 | 型 | デフォルト値 | 必須 | 説明 |
|------|-----|-------------|------|------|
| `DAYS` | int | 30 | No | 分析対象日数（`-days`より優先） |

### 2.3 使用例

#### 2.3.1 基本実行

```bash
# デフォルト設定で実行（過去30日間）
./bin/profit-trend-display

# 過去7日間の分析
./bin/profit-trend-display 7
./bin/profit-trend-display -days 7
```

#### 2.3.2 表示カスタマイズ

```bash
# 大きなチャートで表示
./bin/profit-trend-display -width 100 -height 25

# サマリーのみ表示
./bin/profit-trend-display -summary

# グリッド線なしで表示
./bin/profit-trend-display -grid=false
```

#### 2.3.3 データベース接続設定

```bash
# カスタムデータベースに接続
./bin/profit-trend-display -dsn "user:pass@tcp(localhost:3306)/mydb?parseTime=true"

# 本番環境接続例
./bin/profit-trend-display -dsn "prod_user:${DB_PASS}@tcp(prod-db:3306)/production_db?parseTime=true"
```

### 2.4 戻り値

#### 2.4.1 終了コード

| コード | 意味 | 説明 |
|--------|------|------|
| 0 | 正常終了 | 処理が正常に完了 |
| 1 | 一般エラー | 予期しないエラーが発生 |
| 2 | 設定エラー | 引数やDSNの設定に問題 |
| 3 | データベースエラー | DB接続や SQL実行でエラー |
| 4 | データ不足エラー | 指定期間にデータが存在しない |

#### 2.4.2 標準出力

```
=== 粗利推移表示プログラム ===
分析期間: 過去30日間
接続先: root:***@tcp(mysql.local:3306)/sample_mysql?parseTime=true

対象期間: 2024-06-21 から 2024-07-20 まで

データを取得中...
取得データ数: 150件
データを分析中...

=== 分析結果 ===
対象組織数: 2

(1/2) [会社A - 倉庫1] 粗利推移 (過去30日間)
========================================

   1000 ┬                    ●
    800 ┤                  ●   ●
    600 ┤                ●       ●
    400 ┤              ●           ●
    200 ┤            ●               ●
      0 └─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─
        06/21   06/25   06/29   07/03   07/07

統計情報:
  最大粗利:       1000 (07/01)
  最小粗利:        200 (07/15)
  平均粗利:        600
  合計粗利:      18000
  データ日数: 30日

分析完了!
```

#### 2.4.3 標準エラー出力

```bash
# データベース接続エラー
データベース接続エラー: dial tcp 127.0.0.1:3306: connect: connection refused

# データ不足エラー  
指定された期間にデータが見つかりませんでした。

# 引数エラー
無効な日数が指定されました: -5
```

## 3. 内部API仕様

### 3.1 データモデル

#### 3.1.1 基本データ構造

```go
// ProfitData - 粗利データの基本単位
type ProfitData struct {
    CompanyID       int       `json:"company_id"`       // 会社ID
    CompanyName     string    `json:"company_name"`     // 会社名
    WarehouseBaseID int       `json:"warehouse_base_id"` // 倉庫ID
    WarehouseName   string    `json:"warehouse_name"`   // 倉庫名
    TargetDate      time.Time `json:"target_date"`      // 対象日
    SalesAmount     float64   `json:"sales_amount"`     // 売上金額
    CostAmount      float64   `json:"cost_amount"`      // 原価金額
    ProfitAmount    float64   `json:"profit_amount"`    // 粗利金額
}
```

#### 3.1.2 集計データ構造

```go
// ProfitTrend - 粗利トレンド分析結果
type ProfitTrend struct {
    CompanyID       int           `json:"company_id"`
    CompanyName     string        `json:"company_name"`
    WarehouseBaseID int           `json:"warehouse_base_id"`
    WarehouseName   string        `json:"warehouse_name"`
    Data            []ProfitData  `json:"data"`           // 日次データ配列
    Stats           ProfitStats   `json:"stats"`          // 統計情報
}

// ProfitStats - 統計情報
type ProfitStats struct {
    MaxProfit     float64   `json:"max_profit"`    // 最大粗利
    MinProfit     float64   `json:"min_profit"`    // 最小粗利
    AvgProfit     float64   `json:"avg_profit"`    // 平均粗利
    TotalProfit   float64   `json:"total_profit"`  // 合計粗利
    MaxDate       time.Time `json:"max_date"`      // 最大粗利日
    MinDate       time.Time `json:"min_date"`      // 最小粗利日
    DaysCount     int       `json:"days_count"`    // データ日数
}
```

#### 3.1.3 チャート設定構造

```go
// ChartConfig - チャート描画設定
type ChartConfig struct {
    Width     int     `json:"width"`      // チャート幅
    Height    int     `json:"height"`     // チャート高さ
    MinValue  float64 `json:"min_value"`  // Y軸最小値
    MaxValue  float64 `json:"max_value"`  // Y軸最大値
    ShowGrid  bool    `json:"show_grid"`  // グリッド表示
    ShowStats bool    `json:"show_stats"` // 統計表示
}
```

### 3.2 内部サービスAPI

#### 3.2.1 データベースサービス

```go
type ProfitRepositoryInterface interface {
    // 期間指定粗利データ取得
    GetProfitTrendsForPeriod(startDate, endDate time.Time) ([]ProfitData, error)
    
    // 会社・倉庫一覧取得
    GetCompaniesWithWarehouses() (map[string][]ProfitData, error)
    
    // 接続クローズ
    Close() error
}
```

**メソッド詳細**:

##### GetProfitTrendsForPeriod

```go
func (r *ProfitRepository) GetProfitTrendsForPeriod(
    startDate, endDate time.Time,
) ([]models.ProfitData, error)
```

**パラメータ**:
- `startDate`: 分析開始日（time.Time）
- `endDate`: 分析終了日（time.Time）

**戻り値**:
- `[]models.ProfitData`: 粗利データ配列
- `error`: エラー情報

**エラーパターン**:
```go
var (
    ErrDatabaseConnection = errors.New("database connection failed")
    ErrInvalidDateRange   = errors.New("invalid date range")
    ErrNoDataFound        = errors.New("no data found for specified period")
    ErrQueryExecution     = errors.New("query execution failed")
)
```

#### 3.2.2 計算サービス

```go
type ProfitCalculatorInterface interface {
    // 会社・倉庫別グループ化
    GroupByCompanyWarehouse(data []ProfitData) map[string][]ProfitData
    
    // 粗利トレンド作成
    CreateProfitTrends(groupedData map[string][]ProfitData) []ProfitTrend
    
    // 欠損日補完
    FillMissingDates(data []ProfitData, start, end time.Time) []ProfitData
    
    // 日付範囲計算
    GetDateRange(days int) (time.Time, time.Time)
}
```

**メソッド詳細**:

##### GroupByCompanyWarehouse

```go
func (c *ProfitCalculator) GroupByCompanyWarehouse(
    data []models.ProfitData,
) map[string][]models.ProfitData
```

**処理内容**:
1. 会社名-倉庫名をキーとしてデータをグループ化
2. 各グループ内で日付順にソート
3. キー形式: `"{会社名}-{倉庫名}"`

**計算量**: O(n log n) - nはデータ件数

##### CreateProfitTrends

```go
func (c *ProfitCalculator) CreateProfitTrends(
    groupedData map[string][]models.ProfitData,
) []models.ProfitTrend
```

**処理内容**:
1. 各グループの統計値を計算
2. ProfitTrend構造体を作成
3. 会社名・倉庫名でソート

**統計計算**:
```go
// 統計値計算ロジック
stats := ProfitStats{
    MaxProfit:   max(profits),
    MinProfit:   min(profits),
    AvgProfit:   sum(profits) / count(profits),
    TotalProfit: sum(profits),
    DaysCount:   len(data),
}
```

#### 3.2.3 チャートサービス

```go
type ChartRendererInterface interface {
    // 粗利トレンドチャート描画
    RenderProfitTrend(trend ProfitTrend) string
    
    // サマリー情報描画
    RenderSummary(trends []ProfitTrend) string
}
```

**メソッド詳細**:

##### RenderProfitTrend

```go
func (c *TextChart) RenderProfitTrend(
    trend models.ProfitTrend,
) string
```

**描画アルゴリズム**:

1. **データ正規化**:
```go
// Y座標計算
scaledY := int((maxValue - value) / valueRange * float64(height-1))

// X座標計算  
scaledX := int(float64(index) / float64(dataCount-1) * float64(width-1))
```

2. **シンボル選択**:
```go
func selectSymbol(profit float64) string {
    switch {
    case profit > 0:  return "●"  // 正の粗利
    case profit == 0: return "○"  // ゼロ粗利
    case profit < 0:  return "▼"  // 負の粗利
    }
}
```

3. **グリッド描画**:
```go
// 10カラムごとに縦線
if showGrid && column%10 == 0 {
    chart[row][column] = '┊'
}
```

## 4. 将来のREST API設計

### 4.1 エンドポイント設計

#### 4.1.1 基本エンドポイント

```
GET /api/v1/profit-trends
```

**パラメータ**:

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|------------|-----|------|-----------|------|
| `start_date` | string(date) | No | 30日前 | 開始日（YYYY-MM-DD） |
| `end_date` | string(date) | No | 今日 | 終了日（YYYY-MM-DD） |
| `company_ids` | []int | No | 全て | 対象会社ID配列 |
| `warehouse_ids` | []int | No | 全て | 対象倉庫ID配列 |
| `format` | string | No | json | 出力形式（json/csv/text） |

**レスポンス例**:

```json
{
    "meta": {
        "start_date": "2024-06-21",
        "end_date": "2024-07-20",
        "total_organizations": 2,
        "total_data_points": 60
    },
    "trends": [
        {
            "company_id": 1,
            "company_name": "会社A",
            "warehouse_base_id": 1,
            "warehouse_name": "倉庫1",
            "data": [
                {
                    "target_date": "2024-06-21",
                    "sales_amount": 10000.0,
                    "cost_amount": 7000.0,
                    "profit_amount": 3000.0
                }
            ],
            "stats": {
                "max_profit": 5000.0,
                "min_profit": 1000.0,
                "avg_profit": 3000.0,
                "total_profit": 90000.0,
                "max_date": "2024-07-01",
                "min_date": "2024-06-25",
                "days_count": 30
            }
        }
    ]
}
```

#### 4.1.2 チャート生成エンドポイント

```
GET /api/v1/profit-trends/chart
```

**パラメータ**:

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|------------|-----|------|-----------|------|
| `start_date` | string(date) | No | 30日前 | 開始日 |
| `end_date` | string(date) | No | 今日 | 終了日 |
| `width` | int | No | 60 | チャート幅 |
| `height` | int | No | 15 | チャート高さ |
| `format` | string | No | text | 出力形式（text/svg/png） |

**レスポンス（text形式）**:

```
Content-Type: text/plain; charset=utf-8

[会社A - 倉庫1] 粗利推移 (過去30日間)
========================================

   1000 ┬                    ●
    800 ┤                  ●   ●
    600 ┤                ●       ●
    400 ┤              ●           ●
    200 ┤            ●               ●
      0 └─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─
        06/21   06/25   06/29   07/03   07/07
```

#### 4.1.3 統計情報エンドポイント

```
GET /api/v1/profit-trends/summary
```

**レスポンス例**:

```json
{
    "period": {
        "start_date": "2024-06-21",
        "end_date": "2024-07-20",
        "days": 30
    },
    "overall_stats": {
        "total_profit": 450000.0,
        "average_profit": 15000.0,
        "max_profit": 25000.0,
        "min_profit": 5000.0,
        "organizations_count": 6
    },
    "organization_summary": [
        {
            "company_name": "会社A",
            "warehouse_name": "倉庫1",
            "total_profit": 90000.0,
            "average_profit": 3000.0,
            "profit_ratio": 0.20
        }
    ]
}
```

### 4.2 エラーレスポンス

#### 4.2.1 標準エラー形式

```json
{
    "error": {
        "code": "INVALID_DATE_RANGE",
        "message": "Start date must be before end date",
        "details": {
            "start_date": "2024-07-20",
            "end_date": "2024-06-20"
        },
        "timestamp": "2024-07-20T12:00:00Z"
    }
}
```

#### 4.2.2 エラーコード一覧

| HTTPコード | エラーコード | 説明 |
|------------|-------------|------|
| 400 | `INVALID_DATE_RANGE` | 日付範囲が不正 |
| 400 | `INVALID_PARAMETER` | パラメータ値が不正 |
| 404 | `DATA_NOT_FOUND` | 指定期間にデータなし |
| 500 | `DATABASE_ERROR` | データベースエラー |
| 500 | `INTERNAL_ERROR` | 内部処理エラー |
| 503 | `SERVICE_UNAVAILABLE` | サービス利用不可 |

### 4.3 認証・認可

#### 4.3.1 API キー認証

```http
GET /api/v1/profit-trends
Authorization: Bearer YOUR_API_KEY
```

#### 4.3.2 JWT トークン認証

```http
GET /api/v1/profit-trends  
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 4.4 レート制限

| エンドポイント | 制限 | ウィンドウ |
|----------------|------|-----------|
| データ取得API | 100リクエスト/分 | 1分 |
| チャート生成API | 20リクエスト/分 | 1分 |
| 統計情報API | 50リクエスト/分 | 1分 |

## 5. SDK設計

### 5.1 Go SDK

```go
package profitclient

import (
    "context"
    "time"
)

type Client struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

type ProfitTrendsOptions struct {
    StartDate    *time.Time
    EndDate      *time.Time
    CompanyIDs   []int
    WarehouseIDs []int
    Format       string
}

func NewClient(baseURL, apiKey string) *Client {
    return &Client{
        baseURL:    baseURL,
        apiKey:     apiKey,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *Client) GetProfitTrends(ctx context.Context, opts *ProfitTrendsOptions) (*ProfitTrendsResponse, error) {
    // Implementation
}

func (c *Client) GetChart(ctx context.Context, opts *ChartOptions) (string, error) {
    // Implementation  
}

func (c *Client) GetSummary(ctx context.Context, opts *SummaryOptions) (*SummaryResponse, error) {
    // Implementation
}
```

### 5.2 Python SDK

```python
from dataclasses import dataclass
from datetime import date, datetime
from typing import List, Optional

@dataclass
class ProfitTrendsOptions:
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    company_ids: Optional[List[int]] = None
    warehouse_ids: Optional[List[int]] = None
    format: str = "json"

class ProfitTrendClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.api_key = api_key
        
    def get_profit_trends(self, options: ProfitTrendsOptions) -> dict:
        """粗利トレンドデータを取得"""
        pass
        
    def get_chart(self, options: dict) -> str:
        """チャート文字列を取得"""
        pass
        
    def get_summary(self, options: dict) -> dict:
        """サマリー情報を取得"""
        pass
```

## 6. OpenAPI仕様

### 6.1 OpenAPI定義

```yaml
openapi: 3.0.3
info:
  title: Profit Trend Display API
  description: 粗利推移分析・表示API
  version: 1.0.0
  contact:
    name: API Support
    email: api-support@company.com

servers:
  - url: https://api.company.com/v1
    description: Production server
  - url: https://staging-api.company.com/v1
    description: Staging server

paths:
  /profit-trends:
    get:
      summary: 粗利トレンドデータ取得
      description: 指定期間の粗利トレンドデータを取得します
      parameters:
        - name: start_date
          in: query
          description: 開始日 (YYYY-MM-DD)
          schema:
            type: string
            format: date
        - name: end_date
          in: query
          description: 終了日 (YYYY-MM-DD)
          schema:
            type: string
            format: date
        - name: company_ids
          in: query
          description: 対象会社ID (複数指定可)
          schema:
            type: array
            items:
              type: integer
        - name: warehouse_ids
          in: query
          description: 対象倉庫ID (複数指定可)
          schema:
            type: array
            items:
              type: integer
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProfitTrendsResponse'
        '400':
          description: リクエストエラー
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

components:
  schemas:
    ProfitTrendsResponse:
      type: object
      properties:
        meta:
          $ref: '#/components/schemas/ResponseMeta'
        trends:
          type: array
          items:
            $ref: '#/components/schemas/ProfitTrend'
    
    ProfitTrend:
      type: object
      properties:
        company_id:
          type: integer
        company_name:
          type: string
        warehouse_base_id:
          type: integer
        warehouse_name:
          type: string
        data:
          type: array
          items:
            $ref: '#/components/schemas/ProfitData'
        stats:
          $ref: '#/components/schemas/ProfitStats'
    
    ProfitData:
      type: object
      properties:
        target_date:
          type: string
          format: date
        sales_amount:
          type: number
          format: double
        cost_amount:
          type: number
          format: double
        profit_amount:
          type: number
          format: double
    
    ErrorResponse:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
            details:
              type: object
            timestamp:
              type: string
              format: date-time

  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: Authorization

security:
  - ApiKeyAuth: []
```

## 7. パフォーマンス要件

### 7.1 レスポンス時間

| エンドポイント | 目標レスポンス時間 | 最大許容時間 |
|----------------|-------------------|-------------|
| データ取得（30日） | < 500ms | < 2秒 |
| チャート生成 | < 200ms | < 1秒 |
| サマリー取得 | < 100ms | < 500ms |

### 7.2 スループット

| メトリクス | 目標値 | 最大値 |
|------------|--------|--------|
| 同時接続数 | 100 | 500 |
| リクエスト/秒 | 50 | 200 |
| データサイズ | 1MB | 10MB |

## 8. セキュリティ要件

### 8.1 データ保護

- **暗号化**: HTTPS必須（TLS 1.2以上）
- **認証**: API Key または JWT Token
- **認可**: ロールベースアクセス制御
- **監査ログ**: 全APIアクセスの記録

### 8.2 入力検証

```go
// 入力検証例
func validateDateRange(start, end time.Time) error {
    if start.After(end) {
        return errors.New("start date must be before end date")
    }
    
    maxDays := 365
    if end.Sub(start).Hours() > float64(maxDays*24) {
        return fmt.Errorf("date range cannot exceed %d days", maxDays)
    }
    
    return nil
}
```

## 9. モニタリング

### 9.1 メトリクス

| メトリクス | 説明 | アラート閾値 |
|------------|------|-------------|
| `api_requests_total` | API総リクエスト数 | - |
| `api_request_duration_seconds` | レスポンス時間 | > 2秒 |
| `api_errors_total` | エラー総数 | > 5% |
| `database_connections_active` | アクティブDB接続数 | > 80% |

### 9.2 ログ形式

```json
{
  "timestamp": "2024-07-20T12:00:00Z",
  "level": "INFO",
  "endpoint": "/api/v1/profit-trends",
  "method": "GET",
  "status_code": 200,
  "response_time_ms": 245,
  "request_id": "req-123456789",
  "user_id": "user-123",
  "params": {
    "start_date": "2024-06-20",
    "end_date": "2024-07-20",
    "days": 30
  }
}
```

## 10. 今後の拡張計画

### 10.1 機能拡張

1. **リアルタイム更新**
   - WebSocket対応
   - Server-Sent Events
   - Push通知

2. **分析機能強化**
   - 予測分析API
   - 異常検知API
   - 比較分析API

3. **出力形式拡張**
   - PDF レポート生成
   - Excel ファイル出力
   - SVG/PNG チャート

### 10.2 インフラ拡張

1. **スケーラビリティ**
   - マイクロサービス化
   - キャッシュレイヤー追加
   - CDN対応

2. **可用性向上**
   - マルチリージョン展開
   - 障害時フェイルオーバー
   - サーキットブレーカー