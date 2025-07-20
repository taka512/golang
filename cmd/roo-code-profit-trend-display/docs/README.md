# roo-code-profit-trend-display 仕様書

## 目次

1. [プロジェクト概要](#プロジェクト概要)
2. [アーキテクチャ](#アーキテクチャ)
3. [システム要件](#システム要件)
4. [データベース設計](#データベース設計)
5. [API仕様](#api仕様)
6. [コンポーネント設計](#コンポーネント設計)
7. [使用方法](#使用方法)
8. [Slack通知機能](#slack通知機能)
9. [開発ガイド](#開発ガイド)
10. [トラブルシューティング](#トラブルシューティング)

## ドキュメント構成

- [`specification.md`](./specification.md) - 詳細な技術仕様
- [`api.md`](./api.md) - API仕様書（Slack通知機能含む）
- [`database.md`](./database.md) - データベーススキーマ設計
- [`architecture.md`](./architecture.md) - システムアーキテクチャ
- [`examples/`](./examples/) - 使用例とサンプルコード
- [`tutorials/`](./tutorials/) - チュートリアルとガイド
- [`troubleshooting.md`](./troubleshooting.md) - トラブルシューティングガイド

## プロジェクト概要

**roo-code-profit-trend-display** は、売上と原価データから粗利の推移をテキストベースのグラフで視覚化するGoアプリケーションです。Slack通知機能により、分析結果をリアルタイムでチームに共有することができます。

### 主な機能

🏢 **多組織対応**
- 会社別・倉庫別の粗利分析
- 組織横断的なデータ集計

📊 **高度な可視化**
- ASCIIアートによるテキストグラフ
- カスタマイズ可能なチャートサイズ
- グリッドライン表示

📈 **統計分析**
- 最大・最小・平均・合計値の算出
- 日別トレンド分析
- 欠損データの自動補完

💬 **Slack通知機能**
- 分析結果の自動Slack通知
- エラー発生時の即座の通知
- カスタマイズ可能な通知メッセージ

⚙️ **柔軟な設定**
- コマンドライン引数による詳細設定
- データベース接続の切り替え
- 表示期間の自由設定
- 環境変数による通知設定

🚀 **自動化対応**
- cron等での定期実行
- CI/CD統合
- Docker/Kubernetes対応

### 技術スタック

- **言語**: Go 1.21+
- **データベース**: MySQL 8.0
- **アーキテクチャ**: Clean Architecture
- **パッケージ管理**: Go Modules
- **ビルドツール**: Make
- **通知**: Slack Incoming Webhooks

## クイックスタート

### 基本実行

```bash
# プロジェクトディレクトリに移動
cd cmd/roo-code-profit-trend-display

# 依存関係のインストール
make deps

# ビルド
make build

# 基本実行（過去30日間）
./bin/profit-trend-display

# 過去7日間の分析
./bin/profit-trend-display -days 7

# サマリーのみ表示
./bin/profit-trend-display -summary
```

### Slack通知付き実行

```bash
# Slack Webhook URLを設定
export SLACK_HOOK="https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

# Slack通知付きで実行
./bin/profit-trend-display -slack -days 7

# サマリーをSlackに通知
./bin/profit-trend-display -slack -summary
```

## 使用例

### 基本的な使用例

```bash
# デフォルト設定で実行
./bin/profit-trend-display

# カスタム期間での分析
./bin/profit-trend-display -days 14 -width 80 -height 20

# データベース接続設定
./bin/profit-trend-display -dsn "user:pass@tcp(localhost:3306)/mydb?parseTime=true"
```

### Slack通知を活用した例

```bash
# 日次業務での使用
export SLACK_HOOK="https://hooks.slack.com/services/..."
./bin/profit-trend-display -slack -days 7 -summary

# アラート機能として使用
./bin/profit-trend-display -slack -days 1  # 前日分の即座の確認

# 週次レポートとして使用
./bin/profit-trend-display -slack -days 30 -width 100 -height 25
```

### 自動化での使用例

```bash
# cron設定例
# 毎日朝8時に過去7日間の分析をSlack通知
0 8 * * * SLACK_HOOK="https://hooks.slack.com/..." /path/to/profit-trend-display -slack -days 7 -summary

# 週次レポート（毎週月曜日）
0 9 * * 1 SLACK_HOOK="https://hooks.slack.com/..." /path/to/profit-trend-display -slack -days 30
```

## コマンドラインオプション

| オプション | 型 | デフォルト | 説明 |
|------------|-----|-----------|------|
| `-dsn` | string | `root:mypass@tcp(mysql.local:3306)/sample_mysql?parseTime=true` | データベース接続文字列 |
| `-days` | int | 30 | 分析対象日数（1-365） |
| `-width` | int | 60 | チャート幅（20-200） |
| `-height` | int | 15 | チャート高さ（5-50） |
| `-grid` | bool | true | グリッド線表示 |
| `-stats` | bool | true | 統計情報表示 |
| `-summary` | bool | false | サマリーのみ表示 |
| `-slack` | bool | false | Slack通知有効化 |
| `-help` | bool | false | ヘルプ表示 |

## 環境変数

| 変数名 | 必須 | 説明 |
|--------|------|------|
| `SLACK_HOOK` | No | SlackのIncoming Webhook URL |
| `DB_DSN` | No | データベース接続文字列（`-dsn`より優先度低） |

## 出力例

### 標準出力（通常実行）

```
=== 粗利推移表示プログラム ===
分析期間: 過去7日間
接続先: root:***@tcp(mysql.local:3306)/sample_mysql?parseTime=true

対象期間: 2024-07-14 から 2024-07-20 まで

データを取得中...
取得データ数: 14件
データを分析中...

=== 分析結果 ===
対象組織数: 2

(1/2) [株式会社A - 東京倉庫] 粗利推移 (過去7日間)
========================================

    500 ┬   ●       ●
    400 ┤ ●   ●   ●   ●
    300 ┤       ●       ●
    200 ┤               
    100 ┤               
      0 └─┬─┬─┬─┬─┬─┬─
        07/14 07/16 07/18 07/20

統計情報:
  最大粗利:        500 (07/15)
  最小粗利:        300 (07/17)
  平均粗利:        400
  合計粗利:       2800
  データ日数: 7日

分析完了!
```

### Slack通知付き実行

```
=== 粗利推移表示プログラム ===
分析期間: 過去7日間
接続先: root:***@tcp(mysql.local:3306)/sample_mysql?parseTime=true
Slack通知: 有効

対象期間: 2024-07-14 から 2024-07-20 まで

データを取得中...
取得データ数: 14件
データを分析中...

=== 分析結果 ===
対象組織数: 2

(詳細チャート表示)

Slack通知を送信中...
Slack通知送信完了

分析完了!
```

### Slackで受信されるメッセージ

```
📊 粗利推移分析結果 (過去7日間)

【全体統計】
• 合計粗利: 245,000円
• 平均粗利: 35,000円
• 最大粗利: 55,000円 (07/18)
• 最小粗利: 18,000円 (07/16)
• 対象組織数: 2

【組織別トップ3】
1. 株式会社A - 東京倉庫: 140,000円
2. 株式会社A - 大阪倉庫: 105,000円

実行日時: 2024-07-20 14:30:25
```

## システム要件

### 実行環境
- **OS**: Linux, macOS, Windows
- **Go**: 1.21以上
- **メモリ**: 最小512MB、推奨1GB以上
- **ディスク**: 50MB以上
- **ネットワーク**: HTTPS通信可能（Slack通知使用時）

### データベース要件
- **RDBMS**: MySQL 8.0以上
- **文字コード**: UTF-8
- **接続方式**: TCP/IP
- **必要権限**: SELECT権限

### 外部サービス要件（オプション）
- **Slack**: Incoming Webhook URL（通知機能使用時）

## セキュリティ

### データベースセキュリティ
- TLS 1.2以上での接続暗号化
- 最小権限の原則（SELECT権限のみ）
- 接続文字列のパスワードマスキング

### Slack通知セキュリティ
- Webhook URLの環境変数管理
- HTTPS通信必須
- 機密情報の通知メッセージ除外

## 自動化とCI/CD統合

### GitHub Actions統合

```yaml
# .github/workflows/daily-profit-report.yml
name: Daily Profit Report

on:
  schedule:
    - cron: '0 8 * * 1-5'  # 平日朝8時

jobs:
  profit-analysis:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    - name: Build and Run
      env:
        DB_DSN: ${{ secrets.DB_DSN }}
        SLACK_HOOK: ${{ secrets.SLACK_HOOK }}
      run: |
        cd cmd/roo-code-profit-trend-display
        go build -o bin/profit-trend-display .
        ./bin/profit-trend-display -slack -days 7 -summary
```

### Docker統合

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN cd cmd/roo-code-profit-trend-display && go build -o bin/profit-trend-display .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/cmd/roo-code-profit-trend-display/bin/profit-trend-display .
ENTRYPOINT ["./profit-trend-display"]
```

```bash
# Docker実行例
docker run --rm \
  -e DB_DSN="user:pass@tcp(host:3306)/db?parseTime=true" \
  -e SLACK_HOOK="https://hooks.slack.com/services/..." \
  your-registry/profit-trend-display -slack -days 7 -summary
```

### Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: profit-analysis
spec:
  schedule: "0 8 * * 1-5"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: profit-trend-display
            image: your-registry/profit-trend-display:latest
            args: ["-slack", "-days", "7", "-summary"]
            env:
            - name: DB_DSN
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: dsn
            - name: SLACK_HOOK
              valueFrom:
                secretKeyRef:
                  name: slack-secret
                  key: webhook-url
          restartPolicy: OnFailure
```

## 開発・拡張

### アーキテクチャ概要

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Interface                        │
│                      (main.go)                         │
├─────────────────────────────────────────────────────────┤
│                 Business Logic Layer                    │
├─────────────────┬─────────────────┬─────────────────────┤
│   Calculator    │      Chart      │    Notification     │
│  (calculator/)  │    (chart/)     │  (notification/)    │
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

### 主要コンポーネント

- **Database Layer**: MySQL接続とクエリ実行
- **Calculator Layer**: 粗利計算と統計処理
- **Chart Layer**: テキストグラフ描画
- **Notification Layer**: Slack通知送信
- **CLI Layer**: コマンドライン制御

## パフォーマンス

### 推奨制限値
- 最大分析日数: 365日
- 最大組織数: 100組織  
- 最大チャートサイズ: 200×50
- Slack通知タイムアウト: 10秒

### 実行時間目安
- 30日間データ（10組織）: < 2秒
- 90日間データ（10組織）: < 5秒
- Slack通知送信: < 1秒

## トラブルシューティング

よくある問題と解決方法については [`troubleshooting.md`](./troubleshooting.md) を参照してください。

### クイック診断

```bash
# 接続テスト
./bin/profit-trend-display -days 1 -summary

# Slack通知テスト  
export SLACK_HOOK="https://hooks.slack.com/services/..."
./bin/profit-trend-display -slack -days 1 -summary

# ヘルプ表示
./bin/profit-trend-display -help
```

## サポート・貢献

### 詳細ドキュメント

- [技術仕様書](./specification.md) - 詳細な技術仕様
- [API仕様書](./api.md) - CLI API仕様とSlack通知機能
- [データベース設計](./database.md) - スキーマ設計と最適化
- [基本使用例](./examples/basic-usage/simple-run.md) - 実践的な使用例
- [自動化ガイド](./tutorials/automation-guide.md) - Slack統合を含む自動化

### ライセンス

このプロジェクトのライセンス情報については、プロジェクトルートのLICENSEファイルを参照してください。

---

**💡 ヒント**: まずは `./bin/profit-trend-display -help` でヘルプを確認し、`-days 1 -summary` で軽量な実行を試してみることをお勧めします。Slack通知機能を使用する場合は、事前にWebhook URLを準備してください。