# roo-code-profit-trend-display 技術仕様書

## 1. システム概要

### 1.1 目的

**roo-code-profit-trend-display** は、企業の売上と原価データから粗利の推移を分析し、テキストベースのグラフで視覚化するコマンドラインアプリケーションです。複数の会社・倉庫を横断した粗利分析を可能にし、経営判断に必要な情報を迅速に提供します。また、Slack通知機能により分析結果を即座にチームに共有できます。

### 1.2 対象ユーザー

- 経営陣・管理職
- 財務・経理担当者
- データアナリスト
- システム管理者

### 1.3 主要機能

| 機能分類 | 機能名 | 説明 |
|----------|--------|------|
| データ取得 | 売上・原価データ抽出 | MySQLから指定期間のデータを取得 |
| 計算処理 | 粗利計算 | 売上金額 - 原価金額で粗利を算出 |
| データ加工 | 組織別グループ化 | 会社・倉庫の組み合わせでデータを分類 |
| データ補完 | 欠損日補完 | データがない日を0として補完 |
| 統計分析 | 統計値算出 | 最大・最小・平均・合計値を計算 |
| 可視化 | テキストグラフ表示 | ASCIIアートによるトレンドチャート |
| 出力制御 | フォーマット選択 | 詳細表示・サマリー表示の切り替え |
| 通知機能 | Slack通知 | 分析結果をSlackチャンネルに送信 |

## 2. 技術仕様

### 2.1 システム要件

#### 実行環境
- **OS**: Linux, macOS, Windows
- **Go バージョン**: 1.21以上
- **メモリ**: 最小512MB、推奨1GB以上
- **ディスク容量**: 50MB以上
- **ネットワーク**: HTTPS通信可能（Slack通知使用時）

#### データベース要件
- **RDBMS**: MySQL 8.0以上
- **文字コード**: UTF-8
- **接続方式**: TCP/IP
- **必要権限**: SELECT権限（指定テーブルに対して）

#### 外部サービス要件（オプション）
- **Slack**: Incoming Webhook URL（通知機能使用時）

### 2.2 アーキテクチャ

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Interface                        │
│                      (main.go)                         │
├─────────────────────────────────────────────────────────┤
│                 Business Logic Layer                    │
├─────────────────┬─────────────────┬─────────────────────┤
│   Calculator    │      Chart      │       Models        │
│  (calculator/)  │    (chart/)     │     (models/)       │
├─────────────────┴─────────────────┴─────────────────────┤
│                  Data Access Layer                      │
│                   (database/)                           │
├─────────────────────────────────────────────────────────┤
│                     MySQL Database                      │
├─────────────────────────────────────────────────────────┤
│                  External Services                      │
│                   (Slack API)                           │
└─────────────────────────────────────────────────────────┘
```

### 2.3 モジュール構成

#### 2.3.1 メインモジュール (`main.go`)

**責任**: アプリケーションのエントリーポイント、CLIインターフェースの提供

**主要機能**:
- コマンドライン引数の解析
- 各種設定の管理
- 処理フローの制御
- エラーハンドリング
- Slack通知の制御

**コマンドライン引数**:

| 引数 | 型 | デフォルト値 | 説明 |
|------|----|-----------|----|
| `-dsn` | string | `root:mypass@tcp(mysql.local:3306)/sample_mysql?parseTime=true` | データベース接続文字列 |
| `-days` | int | 30 | 分析対象日数 |
| `-width` | int | 60 | チャート幅 |
| `-height` | int | 15 | チャート高さ |
| `-grid` | bool | true | グリッド線表示 |
| `-stats` | bool | true | 統計情報表示 |
| `-summary` | bool | false | サマリーのみ表示 |
| `-slack` | bool | false | Slack通知有効化 |
| `-help` | bool | false | ヘルプメッセージ表示 |

**環境変数**:

| 変数名 | 型 | 必須 | 説明 |
|--------|-----|------|------|
| `SLACK_HOOK` | string | No | SlackのIncoming Webhook URL |

#### 2.3.2 データベースモジュール (`internal/database/`)

**責任**: データベースアクセス、SQLクエリの実行

**型定義**:
```go
type ProfitRepository struct {
    db *sql.DB
}
```

**主要メソッド**:

| メソッド | 説明 | 戻り値 |
|----------|------|--------|
| `NewProfitRepository(dsn string)` | リポジトリインスタンス作成 | `*ProfitRepository, error` |
| `GetProfitTrendsForPeriod(start, end time.Time)` | 期間指定での粗利データ取得 | `[]models.ProfitData, error` |
| `GetCompaniesWithWarehouses()` | 会社・倉庫の組み合わせ取得 | `map[string][]models.ProfitData, error` |
| `Close()` | DB接続クローズ | `error` |

**SQLクエリ仕様**:
```sql
SELECT 
    c.id as company_id,
    c.name as company_name,
    wb.id as warehouse_base_id,
    wb.name as warehouse_name,
    DATE(COALESCE(sdr.target_date, cdr.target_date)) as target_date,
    COALESCE(SUM(sdri.amount), 0) as sales_amount,
    COALESCE(SUM(cdri.cost_amount), 0) as cost_amount
FROM companies c
CROSS JOIN warehouse_bases wb
LEFT JOIN sales_daily_reports sdr ON c.id = sdr.company_id 
    AND wb.id = sdr.warehouse_base_id 
    AND sdr.target_date BETWEEN ? AND ?
LEFT JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
LEFT JOIN cost_daily_reports cdr ON c.id = cdr.company_id 
    AND wb.id = cdr.warehouse_base_id 
    AND cdr.target_date BETWEEN ? AND ?
LEFT JOIN cost_daily_report_items cdri ON cdr.id = cdri.cost_daily_report_id
WHERE (sdr.target_date IS NOT NULL OR cdr.target_date IS NOT NULL)
GROUP BY c.id, c.name, wb.id, wb.name, DATE(COALESCE(sdr.target_date, cdr.target_date))
ORDER BY c.name, wb.name, DATE(COALESCE(sdr.target_date, cdr.target_date))
```

#### 2.3.3 計算モジュール (`internal/calculator/`)

**責任**: 粗利計算、統計処理、データ加工

**型定義**:
```go
type ProfitCalculator struct{}
```

**主要メソッド**:

| メソッド | 説明 | 計算量 |
|----------|------|--------|
| `GroupByCompanyWarehouse(data []ProfitData)` | 会社・倉庫別データグループ化 | O(n) |
| `CreateProfitTrends(groupedData map[string][]ProfitData)` | トレンドデータ作成 | O(n×m) |
| `calculateStats(data []ProfitData)` | 統計値計算 | O(n) |
| `FillMissingDates(data []ProfitData, start, end time.Time)` | 欠損日補完 | O(d) |
| `GetDateRange(days int)` | 日付範囲計算 | O(1) |

**アルゴリズム詳細**:

1. **粗利計算**:
   ```
   粗利 = 売上金額 - 原価金額
   ```

2. **統計値計算**:
   - 最大値: `max(profit_amounts)`
   - 最小値: `min(profit_amounts)`
   - 平均値: `sum(profit_amounts) / count(days)`
   - 合計値: `sum(profit_amounts)`

3. **欠損日補完アルゴリズム**:
   ```go
   for current := startDate; current <= endDate; current = current.AddDate(0, 0, 1) {
       if data[current] exists {
           result = append(result, data[current])
       } else {
           result = append(result, zeroProfitEntry(current))
       }
   }
   ```

#### 2.3.4 チャートモジュール (`internal/chart/`)

**責任**: テキストベースのグラフ描画

**型定義**:
```go
type TextChart struct {
    config models.ChartConfig
}
```

**主要メソッド**:

| メソッド | 説明 | 出力 |
|----------|------|------|
| `RenderProfitTrend(trend ProfitTrend)` | 粗利トレンドチャート描画 | ASCII文字列 |
| `RenderSummary(trends []ProfitTrend)` | サマリー情報描画 | ASCII文字列 |
| `prepareChartData(data []ProfitData)` | チャートデータ準備 | `[]ChartPoint` |
| `renderChart(points []ChartPoint)` | チャート本体描画 | `[]string` |
| `renderDateAxis(data []ProfitData)` | 日付軸描画 | `string` |
| `renderStats(stats ProfitStats)` | 統計情報描画 | `string` |

**チャート描画仕様**:

1. **座標系**:
   - X軸: 日付（左から右へ時系列）
   - Y軸: 粗利金額（下から上へ昇順）

2. **シンボル**:
   - `●`: 正の粗利
   - `○`: ゼロ粗利
   - `▼`: 負の粗利

3. **グリッド**:
   - 縦線: 10カラムごとに `┊`
   - 横線: Y軸境界線

4. **スケーリング**:
   ```go
   scaledValue = (value - minValue) / (maxValue - minValue) * (height - 1)
   xPosition = index / (dataCount - 1) * (width - 1)
   ```

#### 2.3.5 通知モジュール (`internal/notification/`)

**責任**: 外部サービスへの通知送信

**型定義**:
```go
type SlackNotifier struct {
    webhookURL string
    client     *http.Client
}

type SlackMessage struct {
    Text        string       `json:"text"`
    Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
    Color  string  `json:"color"`
    Title  string  `json:"title"`
    Text   string  `json:"text"`
    Fields []Field `json:"fields"`
}

type Field struct {
    Title string `json:"title"`
    Value string `json:"value"`
    Short bool   `json:"short"`
}
```

**主要メソッド**:

| メソッド | 説明 | 戻り値 |
|----------|------|--------|
| `NewSlackNotifier(webhookURL string)` | Slack通知インスタンス作成 | `*SlackNotifier` |
| `SendProfitSummary(trends []ProfitTrend, period int)` | 粗利サマリーをSlackに送信 | `error` |
| `SendError(err error)` | エラー情報をSlackに送信 | `error` |
| `formatProfitMessage(trends []ProfitTrend, period int)` | 粗利データのフォーマット | `SlackMessage` |

**通知メッセージ仕様**:

1. **サマリー通知**:
   ```json
   {
     "text": "📊 粗利推移分析結果 (過去30日間)",
     "attachments": [
       {
         "color": "good",
         "title": "全体統計",
         "fields": [
           {"title": "合計粗利", "value": "1,250,000円", "short": true},
           {"title": "平均粗利", "value": "41,667円", "short": true},
           {"title": "対象組織数", "value": "3", "short": true}
         ]
       }
     ]
   }
   ```

2. **エラー通知**:
   ```json
   {
     "text": "❌ 粗利分析エラー",
     "attachments": [
       {
         "color": "danger",
         "title": "エラー詳細",
         "text": "データベース接続エラー: connection refused"
       }
     ]
   }
   ```

#### 2.3.6 モデルモジュール (`internal/models/`)

**責任**: データ構造定義

**主要型定義**:

```go
// 粗利データ
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

// 粗利トレンド
type ProfitTrend struct {
    CompanyID       int           `json:"company_id"`
    CompanyName     string        `json:"company_name"`
    WarehouseBaseID int           `json:"warehouse_base_id"`
    WarehouseName   string        `json:"warehouse_name"`
    Data            []ProfitData  `json:"data"`
    Stats           ProfitStats   `json:"stats"`
}

// 統計情報
type ProfitStats struct {
    MaxProfit     float64   `json:"max_profit"`
    MinProfit     float64   `json:"min_profit"`
    AvgProfit     float64   `json:"avg_profit"`
    TotalProfit   float64   `json:"total_profit"`
    MaxDate       time.Time `json:"max_date"`
    MinDate       time.Time `json:"min_date"`
    DaysCount     int       `json:"days_count"`
}

// チャートポイント
type ChartPoint struct {
    Date   time.Time `json:"date"`
    Value  float64   `json:"value"`
    Symbol string    `json:"symbol"`
}

// チャート設定
type ChartConfig struct {
    Width     int     `json:"width"`
    Height    int     `json:"height"`
    MinValue  float64 `json:"min_value"`
    MaxValue  float64 `json:"max_value"`
    ShowGrid  bool    `json:"show_grid"`
    ShowStats bool    `json:"show_stats"`
}

// 通知設定
type NotificationConfig struct {
    SlackEnabled   bool   `json:"slack_enabled"`
    SlackWebhookURL string `json:"slack_webhook_url"`
}
```

## 3. データフロー

### 3.1 処理シーケンス

```
1. コマンドライン引数解析
   ↓
2. 環境変数取得（SLACK_HOOK）
   ↓
3. データベース接続確立
   ↓
4. 日付範囲計算 (Calculator)
   ↓
5. 粗利データ取得 (Database)
   ↓
6. 会社・倉庫別グループ化 (Calculator)
   ↓
7. 欠損日補完 (Calculator)
   ↓
8. 統計値計算 (Calculator)
   ↓
9. チャート描画 (Chart)
   ↓
10. 結果出力 (CLI)
   ↓
11. Slack通知送信 (Notification) ※オプション
```

### 3.2 Slack通知フロー

```mermaid
sequenceDiagram
    participant CLI as main.go
    participant ENV as Environment
    participant CALC as Calculator
    participant SLACK as SlackNotifier
    participant API as Slack API

    CLI->>ENV: SLACK_HOOK環境変数取得
    ENV-->>CLI: Webhook URL
    
    alt Slack通知有効
        CLI->>CALC: 粗利分析実行
        CALC-->>CLI: 分析結果
        
        CLI->>SLACK: NewSlackNotifier(webhookURL)
        SLACK-->>CLI: notifier instance
        
        CLI->>SLACK: SendProfitSummary(trends, period)
        SLACK->>API: HTTP POST request
        API-->>SLACK: HTTP response
        SLACK-->>CLI: success/error
        
        alt エラー発生時
            CLI->>SLACK: SendError(err)
            SLACK->>API: エラー通知送信
        end
    end
```

### 3.3 エラーハンドリング

| エラー種別 | 原因 | 対処法 |
|------------|------|--------|
| データベース接続エラー | DSN不正、サーバー停止 | 接続文字列確認、サーバー状態確認 |
| SQLエラー | テーブル不存在、権限不足 | スキーマ確認、権限設定確認 |
| データ不足エラー | 指定期間にデータなし | 期間変更、データ投入確認 |
| メモリ不足エラー | 大量データ処理 | メモリ増設、期間短縮 |
| Slack通知エラー | Webhook URL不正、ネットワーク障害 | URL確認、ネットワーク状態確認 |

### 3.4 パフォーマンス特性

#### 時間計算量
- データ取得: O(n) - nは対象レコード数
- グループ化: O(n log n) - ソート処理含む
- 統計計算: O(n) - 線形スキャン
- チャート描画: O(w×h) - w=幅、h=高さ
- Slack通知: O(1) - HTTP リクエスト

#### 空間計算量
- メモリ使用量: O(n×m) - n=日数、m=組織数
- 一時的なマップ: O(k) - k=ユニークな組織数
- 通知メッセージ: O(1) - 固定サイズ

#### 推奨制限値
- 最大分析日数: 365日
- 最大組織数: 100組織
- 最大チャートサイズ: 200×50
- Slack通知タイムアウト: 10秒

## 4. セキュリティ

### 4.1 データベースアクセス

- **接続暗号化**: TLS 1.2以上を推奨
- **認証**: ユーザー名・パスワード認証
- **権限制御**: 最小権限の原則（SELECT権限のみ）
- **接続プール**: 適切な接続数制限

### 4.2 Slack通知セキュリティ

- **Webhook URL保護**: 環境変数での管理
- **HTTPS通信**: TLS 1.2以上必須
- **メッセージサイズ制限**: Slack API制限に準拠
- **レート制限**: 1分間に1回までの通知制限

### 4.3 入力検証

- **DSN検証**: 接続文字列のフォーマットチェック
- **数値範囲**: 日数、サイズパラメータの妥当性確認
- **SQLインジェクション対策**: プリペアドステートメント使用
- **Webhook URL検証**: HTTPS URLフォーマットチェック

### 4.4 ログセキュリティ

- **機密情報除外**: パスワード、Webhook URL等の出力抑制
- **アクセスログ**: データベースアクセス履歴の記録
- **通知ログ**: Slack通知の送信履歴記録（URL除外）
- **エラーログ**: セキュリティ関連エラーの詳細記録

## 5. 運用・保守

### 5.1 ログ出力

```
=== 粗利推移表示プログラム ===
分析期間: 過去30日間
接続先: root:***@tcp(mysql.local:3306)/sample_mysql?parseTime=true
Slack通知: 有効

対象期間: 2024-06-21 から 2024-07-20 まで

データを取得中...
取得データ数: 150件
データを分析中...

=== 分析結果 ===
対象組織数: 2

Slack通知を送信中...
Slack通知送信完了

分析完了!
```

### 5.2 監視ポイント

| 項目 | 監視内容 | 閾値 |
|------|----------|------|
| 実行時間 | 処理開始から終了まで | 30秒以内 |
| メモリ使用量 | ピーク時メモリ消費 | 1GB以内 |
| データベース接続 | 接続成功率 | 99%以上 |
| エラー率 | 実行エラー発生率 | 1%未満 |
| Slack通知成功率 | 通知送信成功率 | 95%以上 |

### 5.3 バックアップ・リカバリ

- **設定ファイル**: Makefile、go.mod等の版数管理
- **実行ファイル**: バイナリのバージョン管理
- **ログファイル**: 実行履歴の保管
- **Webhook設定**: 環境変数のバックアップ

## 6. 今後の拡張予定

### 6.1 短期的改善

- **出力フォーマット**: JSON、CSV出力対応
- **フィルタリング**: 会社・倉庫の選択機能
- **期間指定**: 相対日付以外の指定方法
- **通知カスタマイズ**: Slackメッセージフォーマット設定

### 6.2 中期的改善

- **Webインターフェース**: ブラウザベースの操作画面
- **リアルタイム監視**: 定期実行・アラート機能
- **レポート機能**: PDF出力、メール送信
- **多チャンネル通知**: Teams、Discord対応

### 6.3 長期的改善

- **機械学習**: 売上予測、異常検知
- **API化**: REST API提供
- **マルチテナント**: 複数企業での利用
- **ダッシュボード**: リアルタイム監視画面

## 7. Slack通知機能詳細仕様

### 7.1 通知トリガー

| イベント | 通知タイミング | 通知内容 |
|----------|---------------|----------|
| 正常終了 | 分析完了時 | サマリー統計情報 |
| エラー発生 | エラー検出時 | エラー詳細情報 |
| データ異常 | 異常値検出時 | 警告メッセージ |

### 7.2 メッセージフォーマット

#### サマリー通知
```
📊 粗利推移分析結果 (過去30日間)

【全体統計】
• 合計粗利: 1,250,000円
• 平均粗利: 41,667円  
• 最大粗利: 85,000円 (07/15)
• 最小粗利: 12,000円 (07/02)
• 対象組織数: 3

【組織別トップ3】
1. 株式会社A - 東京倉庫: 550,000円
2. 株式会社A - 大阪倉庫: 420,000円
3. 株式会社B - 福岡倉庫: 280,000円

実行日時: 2024-07-20 09:00:00
```

#### エラー通知
```
❌ 粗利分析エラー

エラー内容: データベース接続エラー
詳細: dial tcp 127.0.0.1:3306: connect: connection refused

対処方法:
• MySQLサーバーの起動状態を確認してください
• ネットワーク接続を確認してください
• DSN設定を確認してください

発生日時: 2024-07-20 09:05:23
```

### 7.3 設定例

#### 環境変数設定
```bash
# Slack Webhook URL
export SLACK_HOOK="https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
```

#### 実行例
```bash
# Slack通知付きで実行
./bin/profit-trend-display -slack -days 7 -summary

# 環境変数が設定されていない場合は通知なしで実行
./bin/profit-trend-display -slack -days 30
```

### 7.4 エラーハンドリング

```go
// Slack通知エラーは処理を停止させない
func sendSlackNotification(trends []models.ProfitTrend) {
    if notifier == nil {
        log.Println("Slack通知が無効のため、通知をスキップします")
        return
    }
    
    if err := notifier.SendProfitSummary(trends, days); err != nil {
        log.Printf("Slack通知送信に失敗しました: %v", err)
        // 処理は継続
    } else {
        log.Println("Slack通知送信完了")
    }
}