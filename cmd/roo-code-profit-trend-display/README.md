# 粗利推移表示プログラム (Profit Trend Display)

売上原価から粗利の推移をテキストベースのグラフで視覚的に表示するGoプログラムです。

## 概要

このプログラムは、MySQLデータベースから売上データと原価データを取得し、粗利（売上 - 原価）を計算して、会社別・倉庫別に日次推移をテキストグラフで表示します。

## 機能

- 📊 **日次粗利推移の視覚化**: ASCIIアートによるテキストベースグラフ
- 🏢 **会社別・倉庫別グループ化**: 組織単位での分析
- 📈 **統計情報表示**: 最大・最小・平均・合計値
- ⚙️ **カスタマイズ可能**: グラフサイズ、期間、表示オプション
- 🔄 **欠損データ補完**: データがない日は0として表示

## 前提条件

- Go 1.21以上
- MySQL 8.0
- 以下のテーブルが存在すること:
  - `companies` (会社マスタ)
  - `warehouse_bases` (倉庫マスタ)
  - `sales_daily_reports` (日次売上レポート)
  - `sales_daily_report_items` (売上明細)
  - `cost_daily_reports` (日次原価レポート)
  - `cost_daily_report_items` (原価明細)

## インストール

```bash
# リポジトリクローン後、プロジェクトディレクトリに移動
cd cmd/profit-trend-display

# 依存関係インストール
make deps

# ビルド
make build
```

## 使用方法

### 基本的な使用方法

```bash
# デフォルト設定で実行（過去30日間）
make run

# または直接実行
./bin/profit-trend-display
```

### オプション指定

```bash
# 過去7日間の推移を表示
make run-days DAYS=7

# サマリーのみ表示
make run-summary

# 大きなグラフで表示
make run-large WIDTH=100 HEIGHT=25

# カスタムデータベース接続
make run-db DSN="user:pass@tcp(host:port)/database?parseTime=true"
```

### コマンドラインオプション

| オプション | デフォルト値 | 説明 |
|-----------|-------------|------|
| `-days` | 30 | 分析対象日数 |
| `-width` | 60 | グラフ幅 |
| `-height` | 15 | グラフ高さ |
| `-grid` | true | グリッド線表示 |
| `-stats` | true | 統計情報表示 |
| `-summary` | false | サマリーのみ表示 |
| `-dsn` | root:mypass@tcp... | DB接続文字列 |
| `-help` | false | ヘルプ表示 |

## 出力例

```
=== 粗利推移表示プログラム ===
分析期間: 過去30日間
接続先: root:mypass@tcp(mysql.local:3306)/sample_mysql?parseTime=true

対象期間: 2024-06-21 から 2024-07-20 まで

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

--------------------------------------------------------------------------------

(2/2) [会社B - 倉庫2] 粗利推移 (過去30日間)
========================================
...
```

## プロジェクト構造

```
cmd/profit-trend-display/
├── main.go                 # エントリーポイント
├── go.mod                  # Go modules
├── Makefile               # ビルド設定
├── README.md              # このファイル
├── internal/
│   ├── database/          # データベース接続・クエリ
│   │   └── database.go
│   ├── models/            # データ構造定義
│   │   └── models.go
│   ├── chart/             # テキストグラフ描画
│   │   └── chart.go
│   └── calculator/        # 粗利計算・統計処理
│       └── calculator.go
└── bin/                   # ビルド成果物
    └── profit-trend-display
```

## 開発

### テスト実行

```bash
# テスト実行
make test

# カバレッジ付きテスト
make test-coverage
```

### コード品質

```bash
# フォーマット
make fmt

# リント（golangci-lintが必要）
make lint
```

### サンプル実行

```bash
# 7日間の推移
make example-week

# サマリーのみ
make example-summary

# 大きなグラフ
make example-large
```

## データベース設計

### 主要テーブル

- **sales_daily_reports**: 日次売上レポート
  - company_id, warehouse_base_id, target_date
- **sales_daily_report_items**: 売上明細
  - amount (売上金額)
- **cost_daily_reports**: 日次原価レポート
  - company_id, warehouse_base_id, target_date
- **cost_daily_report_items**: 原価明細
  - cost_amount (原価金額)

### 粗利計算式

```
粗利 = 売上金額 - 原価金額
粗利率 = (粗利 / 売上金額) × 100
```

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   - DSN文字列を確認
   - MySQLサーバーが起動していることを確認

2. **データが表示されない**
   - 指定期間にデータが存在するか確認
   - テーブル名とスキーマが正しいか確認

3. **文字化け**
   - ターミナルがUTF-8をサポートしているか確認

### ログ出力

プログラムは詳細な進行状況を表示します：

```bash
=== 粗利推移表示プログラム ===
分析期間: 過去30日間
データを取得中...
取得データ数: 150件
データを分析中...
=== 分析結果 ===
```

## ライセンス

このプロジェクトは内部利用のためのものです。

## 更新履歴

- v1.0.0: 初回リリース
  - 基本的な粗利推移表示機能
  - テキストベースグラフ描画
  - 統計情報表示
  - 会社別・倉庫別グループ化