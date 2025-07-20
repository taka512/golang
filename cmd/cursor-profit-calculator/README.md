# 出荷粗利計算プログラム

売り上げと原価から出荷の粗利を計算するGoプログラムです。データベースから実際の売上・原価データを取得して計算を行います。

## 機能

- **データベース連携**: MySQLデータベースから売上・原価データを取得
- **粗利計算**: 売上金額 - 原価金額
- **粗利率計算**: (粗利 / 売上金額) × 100
- **レポート表示**: 日別・会社別・倉庫別の詳細レポート
- **サマリー表示**: 総計、平均、日別平均の表示

## データベース構造

このプログラムは以下のテーブルからデータを取得します：

### 売上関連テーブル
- `sales_daily_reports` - 日次売上レポート
- `sales_daily_report_items` - 日次売上レポート明細
- `sales_account_titles` - 売上科目マスタ

### 原価関連テーブル
- `cost_daily_reports` - 日次原価レポート
- `cost_daily_report_items` - 日次原価レポート明細
- `cost_account_titles` - 原価科目マスタ

### マスタテーブル
- `companies` - 会社マスタ
- `warehouse_bases` - 倉庫マスタ

## データ取得クエリ

プログラムは以下のSQLクエリで出荷（shipment）に関する売上・原価データを取得します：

```sql
SELECT 
    s.target_date,
    c.name as company_name,
    w.name as warehouse_name,
    COALESCE(sales_summary.total_amount, 0) as sales_amount,
    COALESCE(cost_summary.total_amount, 0) as cost_amount,
    COALESCE(sales_summary.total_quantity, 0) as sales_quantity,
    COALESCE(cost_summary.total_quantity, 0) as cost_quantity
FROM (
    SELECT DISTINCT target_date, company_id, warehouse_base_id
    FROM sales_daily_reports 
    WHERE sales_account_title_id = (SELECT id FROM sales_account_titles WHERE code = 'shipment')
    UNION
    SELECT DISTINCT target_date, company_id, warehouse_base_id  
    FROM cost_daily_reports
    WHERE cost_account_title_id = (SELECT id FROM cost_account_titles WHERE code = 'shipment')
) s
JOIN companies c ON s.company_id = c.id
JOIN warehouse_bases w ON s.warehouse_base_id = w.id
LEFT JOIN (
    -- 売上サマリー
    SELECT 
        sdr.target_date,
        sdr.company_id,
        sdr.warehouse_base_id,
        SUM(si.amount) as total_amount,
        SUM(si.quantity) as total_quantity
    FROM sales_daily_reports sdr
    JOIN sales_daily_report_items si ON sdr.id = si.sales_daily_report_id
    WHERE sdr.sales_account_title_id = (SELECT id FROM sales_account_titles WHERE code = 'shipment')
    GROUP BY sdr.target_date, sdr.company_id, sdr.warehouse_base_id
) sales_summary ON s.target_date = sales_summary.target_date 
    AND s.company_id = sales_summary.company_id 
    AND s.warehouse_base_id = sales_summary.warehouse_base_id
LEFT JOIN (
    -- 原価サマリー
    SELECT 
        cdr.target_date,
        cdr.company_id,
        cdr.warehouse_base_id,
        SUM(ci.cost_amount) as total_amount,
        SUM(ci.quantity) as total_quantity
    FROM cost_daily_reports cdr
    JOIN cost_daily_report_items ci ON cdr.id = ci.cost_daily_report_id
    WHERE cdr.cost_account_title_id = (SELECT id FROM cost_account_titles WHERE code = 'shipment')
    GROUP BY cdr.target_date, cdr.company_id, cdr.warehouse_base_id
) cost_summary ON s.target_date = cost_summary.target_date 
    AND s.company_id = cost_summary.company_id 
    AND s.warehouse_base_id = cost_summary.warehouse_base_id
ORDER BY s.target_date DESC, c.name, w.name
```

## プログラムの構造

### データ構造

```go
// 出荷データ
type ShipmentData struct {
    Date          string  // 日付
    Company       string  // 会社名
    Warehouse     string  // 倉庫名
    SalesAmount   float64 // 売上金額
    CostAmount    float64 // 原価金額
    SalesQuantity int     // 売上数量
    CostQuantity  int     // 原価数量
}

// 粗利レポート
type ProfitReport struct {
    ShipmentData
    Profit       float64 // 粗利
    ProfitMargin float64 // 粗利率(%)
}
```

### 主要な関数

- `getShipmentData()`: データベースから出荷データを取得
- `calculateProfits()`: 粗利と粗利率を計算
- `calculateSummary()`: サマリー情報を計算
- `printReportRow()`: レポート行を表示
- `formatCurrency()`: 通貨形式でフォーマット

## 実行方法

### 前提条件

1. **MySQLデータベース**が起動していること
2. **データベース接続情報**が正しく設定されていること
   - ホスト: mysql.local:3306
   - データベース: sample_mysql
   - ユーザー: root
   - パスワード: mypass

### 実行コマンド

```bash
cd cmd/profit-calculator

# 依存関係をダウンロード
go mod tidy

# プログラムを実行
go run main.go

# または Makefileを使用
make run
```

## 出力例

```
=== 出荷粗利計算プログラム ===

=== 出荷粗利・実績レポート ===

日付           会社名           倉庫           売上金額       原価金額         粗利    粗利率% 売上数量 原価数量
------------------------------------------------------------------------------------------------------------------------
2024-01-15    株式会社A        東京倉庫      ¥1,500,000    ¥1,200,000    ¥300,000   20.0%     100     100
2024-01-15    株式会社B        大阪倉庫      ¥2,000,000    ¥1,600,000    ¥400,000   20.0%     150     150
2024-01-16    株式会社A        東京倉庫      ¥1,800,000    ¥1,400,000    ¥400,000   22.2%     120     120
2024-01-16    株式会社B        大阪倉庫      ¥2,200,000    ¥1,800,000    ¥400,000   18.2%     180     180
2024-01-17    株式会社A        東京倉庫      ¥1,600,000    ¥1,300,000    ¥300,000   18.8%     110     110
------------------------------------------------------------------------------------------------------------------------

=== サマリー ===
対象日数: 3日
総売上金額: ¥9,100,000
総原価金額: ¥7,300,000
総粗利: ¥1,800,000
平均粗利率: 19.8%
総売上数量: 660
総原価数量: 660

=== 日別平均 ===
平均日売上: ¥3,033,333
平均日原価: ¥2,433,333
平均日粗利: ¥600,000
平均日売上数量: 220
平均日原価数量: 220
```

## カスタマイズ

### データベース接続設定の変更

`main.go`の以下の部分でデータベース接続情報を変更できます：

```go
db, err := sql.Open("mysql", "root:mypass@(mysql.local:3306)/sample_mysql?parseTime=true")
```

### 計算式の変更

粗利計算式は `calculateProfits()` 関数内で定義されています：

```go
// 粗利計算
report.Profit = item.SalesAmount - item.CostAmount

// 粗利率計算
if item.SalesAmount > 0 {
    report.ProfitMargin = (report.Profit / item.SalesAmount) * 100
}
```

## 拡張可能な機能

- 期間指定での絞り込み
- 会社別・倉庫別の集計
- CSV/Excelファイルへの出力
- グラフ表示機能
- 他の科目（入荷、保管など）の計算
- リアルタイムデータ更新 