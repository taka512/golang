# claude-code-profit-report

売上・コスト・粗利のトレンドを表示するコマンドラインツール

## 概要

指定期間の売上・コスト・粗利のトレンドを日別に集計して表示します。
出力先は標準出力またはSlackを選択できます。

## 使用方法

```bash
./claude-code-profit-report --company <会社ID> --warehouse <倉庫ID> --start <開始日> --end <終了日> [--slack]
```

### 必須パラメータ

- `--start, -s`: 開始日 (YYYY-MM-DD形式)
- `--end, -e`: 終了日 (YYYY-MM-DD形式)

### オプションパラメータ

- `--company, -c`: 会社ID（未指定時は全社のデータを集計）
- `--warehouse, -w`: 倉庫ID（未指定時は全倉庫のデータを集計）
- `--slack`: Slackに出力する（環境変数`SLACK_HOOK`の設定が必要）

## 環境変数

### データベース接続
- `DB_HOST`: データベースホスト（デフォルト: localhost）
- `DB_PORT`: データベースポート（デフォルト: 3306）
- `DB_USER`: データベースユーザー（デフォルト: root）
- `DB_PASSWORD`: データベースパスワード
- `DB_NAME`: データベース名（デフォルト: test）

### Slack連携
- `SLACK_HOOK`: Slack Webhook URL（--slackオプション使用時に必須）

## ビルド方法

```bash
cd cmd/claude-code-profit-report
go mod download
go build -o claude-code-profit-report
```

## 実行例

```bash
# 標準出力に表示（特定の会社・倉庫）
./claude-code-profit-report -c 1 -w 1 -s 2024-01-01 -e 2024-01-31

# 全社・全倉庫の集計
./claude-code-profit-report -s 2024-01-01 -e 2024-01-31

# 特定会社の全倉庫集計
./claude-code-profit-report -c 1 -s 2024-01-01 -e 2024-01-31

# Slackにも送信
export SLACK_HOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
./claude-code-profit-report -c 1 -w 1 -s 2024-01-01 -e 2024-01-31 --slack
```