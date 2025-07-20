# 自動化ガイド

このガイドでは、`roo-code-profit-trend-display` を使った自動化の実装方法を説明します。

## 1. 日次レポート自動生成

### 1.1 日次レポートスクリプト

以下のシェルスクリプトを `daily-report.sh` として保存し、cron等で定期実行してください：

```bash
#!/bin/bash

# 日次粗利レポート自動生成スクリプト
# 使用方法: ./daily-report.sh [日数]

# 設定
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${SCRIPT_DIR}/../../../cmd/roo-code-profit-trend-display"
REPORT_DIR="/var/reports/profit-trends"
LOG_DIR="/var/log/profit-trends"
DATE=$(date +%Y%m%d)
DAYS=${1:-7}

# ログファイル設定
LOG_FILE="${LOG_DIR}/daily-report-${DATE}.log"
ERROR_LOG="${LOG_DIR}/daily-report-error-${DATE}.log"

# ディレクトリ作成
mkdir -p "${REPORT_DIR}" "${LOG_DIR}"

# 実行開始ログ
echo "[$(date '+%Y-%m-%d %H:%M:%S')] 日次レポート生成開始 (過去${DAYS}日間)" | tee -a "${LOG_FILE}"

# データベース接続情報（環境変数から取得）
if [ -z "${DB_DSN}" ]; then
    echo "[ERROR] DB_DSN環境変数が設定されていません" | tee -a "${ERROR_LOG}"
    exit 1
fi

# レポート生成
cd "${APP_DIR}"

# サマリーレポート生成
echo "[$(date '+%Y-%m-%d %H:%M:%S')] サマリーレポート生成中..." | tee -a "${LOG_FILE}"
./bin/profit-trend-display \
    -dsn "${DB_DSN}" \
    -days "${DAYS}" \
    -summary \
    > "${REPORT_DIR}/summary-${DATE}.txt" 2>> "${ERROR_LOG}"

if [ $? -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] サマリーレポート生成完了" | tee -a "${LOG_FILE}"
else
    echo "[ERROR] サマリーレポート生成失敗" | tee -a "${ERROR_LOG}"
    exit 1
fi

# 詳細レポート生成
echo "[$(date '+%Y-%m-%d %H:%M:%S')] 詳細レポート生成中..." | tee -a "${LOG_FILE}"
./bin/profit-trend-display \
    -dsn "${DB_DSN}" \
    -days "${DAYS}" \
    -width 80 \
    -height 20 \
    > "${REPORT_DIR}/detailed-${DATE}.txt" 2>> "${ERROR_LOG}"

if [ $? -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 詳細レポート生成完了" | tee -a "${LOG_FILE}"
else
    echo "[ERROR] 詳細レポート生成失敗" | tee -a "${ERROR_LOG}"
    exit 1
fi

# 古いレポートファイルの削除（30日以上前）
find "${REPORT_DIR}" -name "*.txt" -mtime +30 -delete
find "${LOG_DIR}" -name "*.log" -mtime +30 -delete

echo "[$(date '+%Y-%m-%d %H:%M:%S')] 日次レポート生成完了" | tee -a "${LOG_FILE}"

# 成功時の通知（オプション）
if [ -n "${SLACK_WEBHOOK}" ]; then
    curl -X POST -H 'Content-type: application/json' \
         --data "{\"text\":\"📊 日次粗利レポート生成完了 (${DATE})\"}" \
         "${SLACK_WEBHOOK}"
fi
```

### 1.2 実行権限の設定と配置

```bash
# スクリプトを配置
sudo cp daily-report.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/daily-report.sh

# ログディレクトリ作成
sudo mkdir -p /var/log/profit-trends
sudo chown $(whoami):$(whoami) /var/log/profit-trends

# レポートディレクトリ作成
sudo mkdir -p /var/reports/profit-trends  
sudo chown $(whoami):$(whoami) /var/reports/profit-trends
```

### 1.3 環境変数設定

```bash
# ~/.bashrc または /etc/environment に追加
export DB_DSN="prod_user:${DB_PASSWORD}@tcp(prod-db:3306)/production?parseTime=true"
export SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
```

### 1.4 cron設定

```bash
# crontabを編集
crontab -e

# 平日朝8時に実行（過去7日間）
0 8 * * 1-5 /usr/local/bin/daily-report.sh 7

# 毎月1日朝9時に月次レポート実行（過去30日間）
0 9 1 * * /usr/local/bin/daily-report.sh 30
```

## 2. 週次サマリー自動生成

### 2.1 週次サマリースクリプト

以下のスクリプトを `weekly-summary.sh` として作成：

```bash
#!/bin/bash

# 週次粗利サマリー自動生成スクリプト
# 毎週月曜日に前週分析を実行

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${SCRIPT_DIR}/../../../cmd/roo-code-profit-trend-display"
REPORT_DIR="/var/reports/profit-trends/weekly"
DATE=$(date +%Y%m%d)
WEEK=$(date +%V)

# ディレクトリ作成
mkdir -p "${REPORT_DIR}"

echo "=== 週次粗利サマリー (第${WEEK}週) ===" > "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
echo "生成日時: $(date)" >> "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
echo "" >> "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"

# 過去7日、14日、30日の比較
for period in 7 14 30; do
    echo "=== 過去${period}日間の分析 ===" >> "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
    
    cd "${APP_DIR}"
    ./bin/profit-trend-display \
        -dsn "${DB_DSN}" \
        -days ${period} \
        -summary \
        >> "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
    
    echo "" >> "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
    echo "----------------------------------------" >> "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
    echo "" >> "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
done

# メール送信（optonal）
if [ -n "${REPORT_EMAIL}" ]; then
    mail -s "週次粗利サマリー (第${WEEK}週)" "${REPORT_EMAIL}" < "${REPORT_DIR}/week-${WEEK}-${DATE}.txt"
fi
```

### 2.2 週次実行のcron設定

```bash
# 毎週月曜日 朝9時に実行
0 9 * * 1 /usr/local/bin/weekly-summary.sh
```

## 3. アラートシステム

### 3.1 粗利アラートスクリプト

閾値を下回った場合にアラートを送信するスクリプト：

```bash
#!/bin/bash

# 粗利アラートシステム
# 粗利が設定閾値を下回った場合にアラートを送信

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${SCRIPT_DIR}/../../../cmd/roo-code-profit-trend-display"
ALERT_LOG="/var/log/profit-trends/alerts.log"

# 閾値設定
MIN_DAILY_PROFIT=${MIN_DAILY_PROFIT:-10000}    # 1日最低粗利
MIN_WEEKLY_AVG=${MIN_WEEKLY_AVG:-15000}        # 週平均最低粗利

# 一時ファイル
TEMP_FILE="/tmp/profit-analysis-$$.txt"

# 過去7日間の分析実行
cd "${APP_DIR}"
./bin/profit-trend-display -dsn "${DB_DSN}" -days 7 -summary > "${TEMP_FILE}"

if [ $? -ne 0 ]; then
    echo "[$(date)] ERROR: 粗利分析の実行に失敗" >> "${ALERT_LOG}"
    exit 1
fi

# 平均粗利の抽出
AVG_PROFIT=$(grep "平均粗利:" "${TEMP_FILE}" | awk '{print $2}' | head -1)

if [ -z "${AVG_PROFIT}" ]; then
    echo "[$(date)] ERROR: 平均粗利の取得に失敗" >> "${ALERT_LOG}"
    exit 1
fi

# 閾値チェック
if (( $(echo "${AVG_PROFIT} < ${MIN_WEEKLY_AVG}" | bc -l) )); then
    ALERT_MSG="⚠️ 粗利アラート: 週平均粗利が閾値を下回りました
    
現在の週平均粗利: ${AVG_PROFIT}
設定閾値: ${MIN_WEEKLY_AVG}
    
詳細は添付のレポートを確認してください。"

    echo "[$(date)] ALERT: 週平均粗利が閾値を下回りました (${AVG_PROFIT} < ${MIN_WEEKLY_AVG})" >> "${ALERT_LOG}"
    
    # Slack通知
    if [ -n "${SLACK_WEBHOOK}" ]; then
        curl -X POST -H 'Content-type: application/json' \
             --data "{\"text\":\"${ALERT_MSG}\"}" \
             "${SLACK_WEBHOOK}"
    fi
    
    # メール通知
    if [ -n "${ALERT_EMAIL}" ]; then
        echo "${ALERT_MSG}" | mail -s "粗利アラート - 閾値下回り" "${ALERT_EMAIL}"
    fi
else
    echo "[$(date)] INFO: 粗利正常 (週平均: ${AVG_PROFIT})" >> "${ALERT_LOG}"
fi

# 一時ファイル削除
rm -f "${TEMP_FILE}"
```

### 3.2 アラート設定

```bash
# 環境変数設定
export MIN_DAILY_PROFIT=10000
export MIN_WEEKLY_AVG=15000
export ALERT_EMAIL="manager@company.com"
export SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

# cron設定（平日 1時間おきにチェック）
0 9-18 * * 1-5 /usr/local/bin/profit-alert.sh
```

## 4. バッチ処理による複数期間分析

### 4.1 複数期間比較スクリプト

```bash
#!/bin/bash

# 複数期間の粗利比較分析スクリプト

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${SCRIPT_DIR}/../../../cmd/roo-code-profit-trend-display"
OUTPUT_DIR="/var/reports/profit-trends/batch"
DATE=$(date +%Y%m%d)

mkdir -p "${OUTPUT_DIR}"

# 分析期間の配列
PERIODS=(7 14 30 60 90)

echo "=== 複数期間粗利比較分析 ===" > "${OUTPUT_DIR}/comparison-${DATE}.txt"
echo "分析日時: $(date)" >> "${OUTPUT_DIR}/comparison-${DATE}.txt"
echo "" >> "${OUTPUT_DIR}/comparison-${DATE}.txt"

cd "${APP_DIR}"

for period in "${PERIODS[@]}"; do
    echo "=== 過去${period}日間 ===" >> "${OUTPUT_DIR}/comparison-${DATE}.txt"
    
    # サマリーのみを取得
    ./bin/profit-trend-display \
        -dsn "${DB_DSN}" \
        -days "${period}" \
        -summary | \
        grep -A 10 "全体統計:" >> "${OUTPUT_DIR}/comparison-${DATE}.txt"
    
    echo "" >> "${OUTPUT_DIR}/comparison-${DATE}.txt"
    echo "----------------------------------------" >> "${OUTPUT_DIR}/comparison-${DATE}.txt"
    echo "" >> "${OUTPUT_DIR}/comparison-${DATE}.txt"
done

# Excel形式でも出力（CSV）
echo "期間,合計粗利,平均粗利,最大粗利,最小粗利,対象組織数" > "${OUTPUT_DIR}/comparison-${DATE}.csv"

for period in "${PERIODS[@]}"; do
    SUMMARY=$(./bin/profit-trend-display -dsn "${DB_DSN}" -days "${period}" -summary | grep -A 5 "全体統計:")
    
    TOTAL=$(echo "${SUMMARY}" | grep "合計粗利:" | awk '{print $2}')
    AVG=$(echo "${SUMMARY}" | grep "平均粗利:" | awk '{print $2}')
    MAX=$(echo "${SUMMARY}" | grep "最大粗利:" | awk '{print $2}')
    MIN=$(echo "${SUMMARY}" | grep "最小粗利:" | awk '{print $2}')
    ORGS=$(echo "${SUMMARY}" | grep "対象組織数:" | awk '{print $2}')
    
    echo "${period},${TOTAL},${AVG},${MAX},${MIN},${ORGS}" >> "${OUTPUT_DIR}/comparison-${DATE}.csv"
done

echo "比較分析完了: ${OUTPUT_DIR}/comparison-${DATE}.txt"
echo "CSV出力: ${OUTPUT_DIR}/comparison-${DATE}.csv"
```

### 4.2 月次バッチ実行

```bash
# 月末最終営業日に実行
0 18 * * 5 [ $(date -d "+3 days" +\%m) != $(date +\%m) ] && /usr/local/bin/batch-comparison.sh
```

## 5. システム統合

### 5.1 CI/CDパイプライン統合

GitHub Actionsでの自動レポート生成例：

```yaml
# .github/workflows/daily-profit-report.yml
name: Daily Profit Report

on:
  schedule:
    - cron: '0 8 * * 1-5'  # 平日朝8時
  workflow_dispatch:       # 手動実行も可能

jobs:
  generate-report:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Build Application
      run: |
        cd cmd/roo-code-profit-trend-display
        go build -o bin/profit-trend-display .
    
    - name: Generate Report
      env:
        DB_DSN: ${{ secrets.DB_DSN }}
      run: |
        cd cmd/roo-code-profit-trend-display
        ./bin/profit-trend-display -days 7 -summary > daily-report.txt
    
    - name: Upload Report
      uses: actions/upload-artifact@v3
      with:
        name: daily-profit-report
        path: cmd/roo-code-profit-trend-display/daily-report.txt
    
    - name: Notify Slack
      if: always()
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: Daily profit report generated
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### 5.2 Dockerでの実行環境

```dockerfile
# Dockerfile.reporting
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN cd cmd/roo-code-profit-trend-display && \
    go build -o bin/profit-trend-display .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/cmd/roo-code-profit-trend-display/bin/profit-trend-display .
COPY scripts/docker-entrypoint.sh .

RUN chmod +x docker-entrypoint.sh

ENTRYPOINT ["./docker-entrypoint.sh"]
```

Docker実行用エントリーポイント：

```bash
#!/bin/sh
# docker-entrypoint.sh

# 環境変数のデフォルト値設定
DAYS=${DAYS:-7}
FORMAT=${FORMAT:-summary}
OUTPUT_DIR=${OUTPUT_DIR:-/reports}

# 出力ディレクトリ作成
mkdir -p "${OUTPUT_DIR}"

# レポート生成
case "${FORMAT}" in
    "summary")
        ./profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" -summary > "${OUTPUT_DIR}/report-$(date +%Y%m%d).txt"
        ;;
    "detailed")
        ./profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" -width 100 -height 25 > "${OUTPUT_DIR}/detailed-$(date +%Y%m%d).txt"
        ;;
    *)
        ./profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" "${@}"
        ;;
esac
```

### 5.3 Kubernetes CronJobとして実行

```yaml
# k8s-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: daily-profit-report
spec:
  schedule: "0 8 * * 1-5"  # 平日朝8時
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: profit-trend-display
            image: your-registry/profit-trend-display:latest
            env:
            - name: DB_DSN
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: dsn
            - name: DAYS
              value: "7"
            - name: FORMAT
              value: "summary"
            volumeMounts:
            - name: report-storage
              mountPath: /reports
          volumes:
          - name: report-storage
            persistentVolumeClaim:
              claimName: report-pvc
          restartPolicy: OnFailure
```

## 6. 監視とログ管理

### 6.1 ログ管理設定

```bash
# logrotate設定ファイル: /etc/logrotate.d/profit-trends
/var/log/profit-trends/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 profit profit
    postrotate
        # 必要に応じて追加処理
    endscript
}
```

### 6.2 監視スクリプト

```bash
#!/bin/bash

# 自動化プロセス監視スクリプト

LOG_DIR="/var/log/profit-trends"
TODAY=$(date +%Y%m%d)

# 今日のレポート生成確認
if [ ! -f "/var/reports/profit-trends/summary-${TODAY}.txt" ]; then
    echo "WARNING: 今日の日次レポートが生成されていません"
    # アラート送信
fi

# エラーログチェック
ERROR_COUNT=$(grep -c "ERROR" "${LOG_DIR}/daily-report-${TODAY}.log" 2>/dev/null || echo 0)
if [ "${ERROR_COUNT}" -gt 0 ]; then
    echo "WARNING: 今日のレポート生成で${ERROR_COUNT}件のエラーが発生"
    # アラート送信
fi

# ディスク容量チェック
DISK_USAGE=$(df /var/reports/profit-trends | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "${DISK_USAGE}" -gt 80 ]; then
    echo "WARNING: レポートディスクの使用率が${DISK_USAGE}%です"
    # アラート送信
fi
```

## まとめ

このガイドで紹介した自動化手法により、以下が実現できます：

1. **定期的なレポート生成**: 日次・週次・月次での自動実行
2. **アラートシステム**: 閾値監視と即座の通知
3. **バッチ処理**: 複数期間の比較分析
4. **システム統合**: CI/CD、Docker、Kubernetesとの連携
5. **監視・ログ管理**: 自動化プロセスの健全性監視

これらの仕組みにより、手作業を最小限に抑えつつ、継続的な粗利分析が可能になります。