# 自動化ガイド（Slack通知統合版）

このガイドでは、`roo-code-profit-trend-display` を使った自動化の実装方法を説明します。Slack通知機能を活用して、分析結果をリアルタイムでチームに共有する方法も含まれています。

## 1. Slack通知付き日次レポート自動生成

### 1.1 基本的な日次レポートスクリプト

以下のシェルスクリプトを `daily-report-slack.sh` として保存し、cron等で定期実行してください：

```bash
#!/bin/bash

# Slack通知付き日次粗利レポート自動生成スクリプト
# 使用方法: ./daily-report-slack.sh [日数]

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
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Slack通知付き日次レポート生成開始 (過去${DAYS}日間)" | tee -a "${LOG_FILE}"

# 必須環境変数チェック
if [ -z "${DB_DSN}" ]; then
    echo "[ERROR] DB_DSN環境変数が設定されていません" | tee -a "${ERROR_LOG}"
    exit 1
fi

if [ -z "${SLACK_HOOK}" ]; then
    echo "[WARNING] SLACK_HOOK環境変数が未設定 - Slack通知なしで実行します" | tee -a "${LOG_FILE}"
    SLACK_ENABLED=false
else
    echo "[INFO] Slack通知が有効です" | tee -a "${LOG_FILE}"
    SLACK_ENABLED=true
fi

# レポート生成
cd "${APP_DIR}"

# Slack通知付きサマリーレポート生成
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Slack通知付きサマリーレポート生成中..." | tee -a "${LOG_FILE}"
if [ "${SLACK_ENABLED}" = true ]; then
    ./bin/profit-trend-display \
        -dsn "${DB_DSN}" \
        -days "${DAYS}" \
        -summary \
        -slack \
        > "${REPORT_DIR}/summary-${DATE}.txt" 2>> "${ERROR_LOG}"
else
    ./bin/profit-trend-display \
        -dsn "${DB_DSN}" \
        -days "${DAYS}" \
        -summary \
        > "${REPORT_DIR}/summary-${DATE}.txt" 2>> "${ERROR_LOG}"
fi

if [ $? -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] サマリーレポート生成完了" | tee -a "${LOG_FILE}"
else
    echo "[ERROR] サマリーレポート生成失敗" | tee -a "${ERROR_LOG}"
    
    # エラー時のSlack通知
    if [ "${SLACK_ENABLED}" = true ]; then
        curl -X POST -H 'Content-type: application/json' \
             --data "{
                 \"text\": \"❌ 日次粗利レポート生成エラー\",
                 \"attachments\": [{
                     \"color\": \"danger\",
                     \"title\": \"エラー詳細\",
                     \"text\": \"日次レポート生成に失敗しました。ログを確認してください。\",
                     \"fields\": [{
                         \"title\": \"日付\",
                         \"value\": \"${DATE}\",
                         \"short\": true
                     }, {
                         \"title\": \"対象期間\",
                         \"value\": \"過去${DAYS}日間\",
                         \"short\": true
                     }]
                 }]
             }" \
             "${SLACK_HOOK}" 2>/dev/null
    fi
    exit 1
fi

# 詳細レポート生成（Slack通知なし）
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

# 成功時の追加Slack通知（詳細情報付き）
if [ "${SLACK_ENABLED}" = true ]; then
    # レポートファイルから統計情報を抽出
    TOTAL_PROFIT=$(grep "合計粗利:" "${REPORT_DIR}/summary-${DATE}.txt" | head -1 | awk '{print $2}')
    AVG_PROFIT=$(grep "平均粗利:" "${REPORT_DIR}/summary-${DATE}.txt" | head -1 | awk '{print $2}')
    ORG_COUNT=$(grep "対象組織数:" "${REPORT_DIR}/summary-${DATE}.txt" | head -1 | awk '{print $2}')
    
    curl -X POST -H 'Content-type: application/json' \
         --data "{
             \"text\": \"✅ 日次粗利レポート生成完了\",
             \"attachments\": [{
                 \"color\": \"good\",
                 \"title\": \"レポート概要 (過去${DAYS}日間)\",
                 \"fields\": [{
                     \"title\": \"合計粗利\",
                     \"value\": \"${TOTAL_PROFIT:-'N/A'}円\",
                     \"short\": true
                 }, {
                     \"title\": \"平均粗利\",
                     \"value\": \"${AVG_PROFIT:-'N/A'}円\",
                     \"short\": true
                 }, {
                     \"title\": \"対象組織数\",
                     \"value\": \"${ORG_COUNT:-'N/A'}\",
                     \"short\": true
                 }, {
                     \"title\": \"生成日時\",
                     \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                     \"short\": true
                 }]
             }]
         }" \
         "${SLACK_HOOK}" 2>/dev/null
fi
```

### 1.2 環境変数設定

```bash
# ~/.bashrc または /etc/environment に追加
export DB_DSN="prod_user:${DB_PASSWORD}@tcp(prod-db:3306)/production?parseTime=true"
export SLACK_HOOK="https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

# オプション: アラート用の追加設定
export ALERT_EMAIL="manager@company.com"
export MIN_WEEKLY_AVG=15000
```

### 1.3 実行権限の設定と配置

```bash
# スクリプトを配置
sudo cp daily-report-slack.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/daily-report-slack.sh

# ログディレクトリ作成
sudo mkdir -p /var/log/profit-trends
sudo chown $(whoami):$(whoami) /var/log/profit-trends

# レポートディレクトリ作成
sudo mkdir -p /var/reports/profit-trends  
sudo chown $(whoami):$(whoami) /var/reports/profit-trends

# テスト実行
/usr/local/bin/daily-report-slack.sh 1
```

### 1.4 cron設定

```bash
# crontabを編集
crontab -e

# 平日朝8時に実行（過去7日間、Slack通知付き）
0 8 * * 1-5 /usr/local/bin/daily-report-slack.sh 7

# 毎月1日朝9時に月次レポート実行（過去30日間）
0 9 1 * * /usr/local/bin/daily-report-slack.sh 30

# 緊急用: 毎時サマリーチェック（営業時間のみ）
0 9-17 * * 1-5 /usr/local/bin/hourly-check.sh
```

## 2. 高度なSlack通知システム

### 2.1 状況別通知スクリプト

```bash
#!/bin/bash

# 状況別Slack通知スクリプト
# 使用方法: ./smart-notify.sh [days] [notification_type]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${SCRIPT_DIR}/../../../cmd/roo-code-profit-trend-display"
DAYS=${1:-7}
NOTIFICATION_TYPE=${2:-"auto"}  # auto, success, warning, error

# 必須環境変数チェック
if [ -z "${SLACK_HOOK}" ]; then
    echo "ERROR: SLACK_HOOK環境変数が設定されていません"
    exit 1
fi

# 一時ファイル
TEMP_FILE="/tmp/profit-analysis-$$.txt"

# 分析実行
cd "${APP_DIR}"
./bin/profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" -summary > "${TEMP_FILE}" 2>&1
ANALYSIS_RESULT=$?

# 結果の解析
if [ ${ANALYSIS_RESULT} -eq 0 ]; then
    # 成功時の詳細分析
    TOTAL_PROFIT=$(grep "合計粗利:" "${TEMP_FILE}" | head -1 | awk '{print $2}')
    AVG_PROFIT=$(grep "平均粗利:" "${TEMP_FILE}" | head -1 | awk '{print $2}')
    MAX_PROFIT=$(grep "最大粗利:" "${TEMP_FILE}" | head -1 | awk '{print $2}')
    MIN_PROFIT=$(grep "最小粗利:" "${TEMP_FILE}" | head -1 | awk '{print $2}')
    ORG_COUNT=$(grep "対象組織数:" "${TEMP_FILE}" | head -1 | awk '{print $2}')
    
    # 閾値チェック
    THRESHOLD=${MIN_WEEKLY_AVG:-15000}
    if (( $(echo "${AVG_PROFIT} < ${THRESHOLD}" | bc -l) )); then
        NOTIFICATION_TYPE="warning"
        COLOR="warning"
        ICON="⚠️"
        TITLE="粗利低下警告"
    else
        NOTIFICATION_TYPE="success"
        COLOR="good"
        ICON="📊"
        TITLE="粗利分析結果"
    fi
else
    # エラー時
    NOTIFICATION_TYPE="error"
    COLOR="danger"
    ICON="❌"
    TITLE="分析エラー"
fi

# 通知メッセージ作成
case "${NOTIFICATION_TYPE}" in
    "success")
        MESSAGE="{
            \"text\": \"${ICON} ${TITLE} (過去${DAYS}日間)\",
            \"attachments\": [{
                \"color\": \"${COLOR}\",
                \"title\": \"全体統計\",
                \"fields\": [{
                    \"title\": \"合計粗利\",
                    \"value\": \"$(printf \"%'d\" ${TOTAL_PROFIT})円\",
                    \"short\": true
                }, {
                    \"title\": \"平均粗利\",
                    \"value\": \"$(printf \"%'d\" ${AVG_PROFIT})円\",
                    \"short\": true
                }, {
                    \"title\": \"最大粗利\",
                    \"value\": \"$(printf \"%'d\" ${MAX_PROFIT})円\",
                    \"short\": true
                }, {
                    \"title\": \"最小粗利\",
                    \"value\": \"$(printf \"%'d\" ${MIN_PROFIT})円\",
                    \"short\": true
                }, {
                    \"title\": \"対象組織数\",
                    \"value\": \"${ORG_COUNT}\",
                    \"short\": true
                }, {
                    \"title\": \"分析日時\",
                    \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                    \"short\": true
                }]
            }]
        }"
        ;;
    "warning")
        MESSAGE="{
            \"text\": \"${ICON} ${TITLE} (過去${DAYS}日間)\",
            \"attachments\": [{
                \"color\": \"${COLOR}\",
                \"title\": \"警告: 平均粗利が閾値を下回りました\",
                \"fields\": [{
                    \"title\": \"現在の平均粗利\",
                    \"value\": \"$(printf \"%'d\" ${AVG_PROFIT})円\",
                    \"short\": true
                }, {
                    \"title\": \"設定閾値\",
                    \"value\": \"$(printf \"%'d\" ${THRESHOLD})円\",
                    \"short\": true
                }, {
                    \"title\": \"対象組織数\",
                    \"value\": \"${ORG_COUNT}\",
                    \"short\": true
                }, {
                    \"title\": \"確認時刻\",
                    \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                    \"short\": true
                }],
                \"text\": \"詳細な分析と対策を検討してください。\"
            }]
        }"
        ;;
    "error")
        ERROR_MSG=$(cat "${TEMP_FILE}" | tail -5 | tr '\n' ' ')
        MESSAGE="{
            \"text\": \"${ICON} ${TITLE}\",
            \"attachments\": [{
                \"color\": \"${COLOR}\",
                \"title\": \"粗利分析の実行に失敗しました\",
                \"fields\": [{
                    \"title\": \"対象期間\",
                    \"value\": \"過去${DAYS}日間\",
                    \"short\": true
                }, {
                    \"title\": \"発生時刻\",
                    \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                    \"short\": true
                }],
                \"text\": \"エラー詳細: ${ERROR_MSG}\"
            }]
        }"
        ;;
esac

# Slack通知送信
curl -X POST \
     -H 'Content-type: application/json' \
     --data "${MESSAGE}" \
     "${SLACK_HOOK}" \
     --max-time 10 \
     --silent

# 一時ファイル削除
rm -f "${TEMP_FILE}"

echo "Slack通知送信完了: ${NOTIFICATION_TYPE}"
```

### 2.2 週次比較レポート（Slack通知付き）

```bash
#!/bin/bash

# 週次比較レポート（Slack通知付き）
# 毎週月曜日に前週との比較分析を実行

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${SCRIPT_DIR}/../../../cmd/roo-code-profit-trend-display"
REPORT_DIR="/var/reports/profit-trends/weekly"
DATE=$(date +%Y%m%d)
WEEK=$(date +%V)

# ディレクトリ作成
mkdir -p "${REPORT_DIR}"

# 今週と先週のデータを比較
CURRENT_WEEK_FILE="/tmp/current-week-$$.txt"
LAST_WEEK_FILE="/tmp/last-week-$$.txt"

cd "${APP_DIR}"

# 今週のデータ（過去7日間）
./bin/profit-trend-display -dsn "${DB_DSN}" -days 7 -summary > "${CURRENT_WEEK_FILE}"
CURRENT_AVG=$(grep "平均粗利:" "${CURRENT_WEEK_FILE}" | awk '{print $2}')
CURRENT_TOTAL=$(grep "合計粗利:" "${CURRENT_WEEK_FILE}" | awk '{print $2}')

# 先週のデータ（8-14日前）
# 注意: この実装は簡略化されています。実際には日付範囲を指定する機能が必要です
./bin/profit-trend-display -dsn "${DB_DSN}" -days 14 -summary > "${LAST_WEEK_FILE}"
# 簡易的な先週データ推定（実装要改善）
LAST_AVG=$(grep "平均粗利:" "${LAST_WEEK_FILE}" | awk '{print $2}')
LAST_TOTAL=$(grep "合計粗利:" "${LAST_WEEK_FILE}" | awk '{print $2}')

# 変化率計算
if [ -n "${CURRENT_AVG}" ] && [ -n "${LAST_AVG}" ] && [ "${LAST_AVG}" != "0" ]; then
    CHANGE_RATE=$(echo "scale=1; (${CURRENT_AVG} - ${LAST_AVG}) / ${LAST_AVG} * 100" | bc)
    
    if (( $(echo "${CHANGE_RATE} > 0" | bc -l) )); then
        TREND_ICON="📈"
        TREND_COLOR="good"
        TREND_TEXT="改善"
    elif (( $(echo "${CHANGE_RATE} < -5" | bc -l) )); then
        TREND_ICON="📉"
        TREND_COLOR="danger"
        TREND_TEXT="悪化"
    else
        TREND_ICON="📊"
        TREND_COLOR="warning"
        TREND_TEXT="横ばい"
    fi
else
    CHANGE_RATE="N/A"
    TREND_ICON="📊"
    TREND_COLOR="good"
    TREND_TEXT="データ不足"
fi

# Slack通知送信
if [ -n "${SLACK_HOOK}" ]; then
    curl -X POST -H 'Content-type: application/json' \
         --data "{
             \"text\": \"${TREND_ICON} 週次粗利比較レポート (第${WEEK}週)\",
             \"attachments\": [{
                 \"color\": \"${TREND_COLOR}\",
                 \"title\": \"先週比較\",
                 \"fields\": [{
                     \"title\": \"今週平均粗利\",
                     \"value\": \"$(printf \"%'d\" ${CURRENT_AVG})円\",
                     \"short\": true
                 }, {
                     \"title\": \"変化率\",
                     \"value\": \"${CHANGE_RATE}% (${TREND_TEXT})\",
                     \"short\": true
                 }, {
                     \"title\": \"今週合計\",
                     \"value\": \"$(printf \"%'d\" ${CURRENT_TOTAL})円\",
                     \"short\": true
                 }, {
                     \"title\": \"分析日時\",
                     \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                     \"short\": true
                 }]
             }]
         }" \
         "${SLACK_HOOK}"
fi

# 一時ファイル削除
rm -f "${CURRENT_WEEK_FILE}" "${LAST_WEEK_FILE}"
```

### 2.3 リアルタイムアラートシステム

```bash
#!/bin/bash

# リアルタイム粗利アラートシステム
# 閾値を下回った場合に即座にSlack通知

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${SCRIPT_DIR}/../../../cmd/roo-code-profit-trend-display"
ALERT_LOG="/var/log/profit-trends/alerts.log"

# 閾値設定（環境変数または設定ファイルから取得）
MIN_DAILY_PROFIT=${MIN_DAILY_PROFIT:-10000}    # 1日最低粗利
MIN_WEEKLY_AVG=${MIN_WEEKLY_AVG:-15000}        # 週平均最低粗利
CRITICAL_THRESHOLD=${CRITICAL_THRESHOLD:-5000} # 緊急閾値

# 一時ファイル
TEMP_FILE="/tmp/profit-alert-$$.txt"

# 過去7日間の分析実行（Slack通知付き）
cd "${APP_DIR}"
./bin/profit-trend-display -dsn "${DB_DSN}" -days 7 -summary -slack > "${TEMP_FILE}" 2>&1
ANALYSIS_RESULT=$?

if [ ${ANALYSIS_RESULT} -ne 0 ]; then
    echo "[$(date)] ERROR: 粗利分析の実行に失敗" >> "${ALERT_LOG}"
    
    # 分析失敗時のSlack通知
    if [ -n "${SLACK_HOOK}" ]; then
        curl -X POST -H 'Content-type: application/json' \
             --data "{
                 \"text\": \"🚨 緊急: 粗利分析システムエラー\",
                 \"attachments\": [{
                     \"color\": \"danger\",
                     \"title\": \"システム障害発生\",
                     \"text\": \"粗利分析の実行に失敗しました。システム管理者に連絡してください。\",
                     \"fields\": [{
                         \"title\": \"発生時刻\",
                         \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                         \"short\": true
                     }]
                 }]
             }" \
             "${SLACK_HOOK}"
    fi
    exit 1
fi

# 平均粗利の抽出
AVG_PROFIT=$(grep "平均粗利:" "${TEMP_FILE}" | awk '{print $2}' | head -1)
MIN_PROFIT=$(grep "最小粗利:" "${TEMP_FILE}" | awk '{print $2}' | head -1)

if [ -z "${AVG_PROFIT}" ] || [ -z "${MIN_PROFIT}" ]; then
    echo "[$(date)] ERROR: 粗利データの取得に失敗" >> "${ALERT_LOG}"
    exit 1
fi

# 緊急レベルのアラート判定
if (( $(echo "${AVG_PROFIT} < ${CRITICAL_THRESHOLD}" | bc -l) )); then
    ALERT_LEVEL="CRITICAL"
    ALERT_COLOR="danger"
    ALERT_ICON="🚨"
    ALERT_TITLE="緊急: 粗利が危険レベルまで低下"
elif (( $(echo "${AVG_PROFIT} < ${MIN_WEEKLY_AVG}" | bc -l) )); then
    ALERT_LEVEL="WARNING"
    ALERT_COLOR="warning"
    ALERT_ICON="⚠️"
    ALERT_TITLE="警告: 粗利が閾値を下回りました"
elif (( $(echo "${MIN_PROFIT} < ${MIN_DAILY_PROFIT}" | bc -l) )); then
    ALERT_LEVEL="INFO"
    ALERT_COLOR="warning"
    ALERT_ICON="ℹ️"
    ALERT_TITLE="情報: 最低日次粗利が閾値を下回りました"
else
    echo "[$(date)] INFO: 粗利正常 (週平均: ${AVG_PROFIT})" >> "${ALERT_LOG}"
    rm -f "${TEMP_FILE}"
    exit 0
fi

# アラートログ記録
echo "[$(date)] ${ALERT_LEVEL}: ${ALERT_TITLE} (平均: ${AVG_PROFIT}, 最小: ${MIN_PROFIT})" >> "${ALERT_LOG}"

# Slack緊急通知送信
if [ -n "${SLACK_HOOK}" ]; then
    curl -X POST -H 'Content-type: application/json' \
         --data "{
             \"text\": \"${ALERT_ICON} ${ALERT_TITLE}\",
             \"attachments\": [{
                 \"color\": \"${ALERT_COLOR}\",
                 \"title\": \"粗利アラート詳細\",
                 \"fields\": [{
                     \"title\": \"週平均粗利\",
                     \"value\": \"$(printf \"%'d\" ${AVG_PROFIT})円\",
                     \"short\": true
                 }, {
                     \"title\": \"最小日次粗利\",
                     \"value\": \"$(printf \"%'d\" ${MIN_PROFIT})円\",
                     \"short\": true
                 }, {
                     \"title\": \"設定閾値 (週平均)\",
                     \"value\": \"$(printf \"%'d\" ${MIN_WEEKLY_AVG})円\",
                     \"short\": true
                 }, {
                     \"title\": \"緊急閾値\",
                     \"value\": \"$(printf \"%'d\" ${CRITICAL_THRESHOLD})円\",
                     \"short\": true
                 }, {
                     \"title\": \"アラートレベル\",
                     \"value\": \"${ALERT_LEVEL}\",
                     \"short\": true
                 }, {
                     \"title\": \"発生時刻\",
                     \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                     \"short\": true
                 }],
                 \"text\": \"緊急の対応が必要な場合があります。詳細な分析と対策を検討してください。\"
             }]
         }" \
         "${SLACK_HOOK}"
fi

# 緊急レベルの場合は追加通知
if [ "${ALERT_LEVEL}" = "CRITICAL" ] && [ -n "${ALERT_EMAIL}" ]; then
    echo "緊急粗利アラート: 平均粗利が${AVG_PROFIT}円まで低下しました。即座の対応が必要です。" | \
    mail -s "【緊急】粗利危険レベル到達" "${ALERT_EMAIL}"
fi

# 一時ファイル削除
rm -f "${TEMP_FILE}"
```

## 3. CI/CD統合とSlack通知

### 3.1 GitHub Actions with Slack

```yaml
# .github/workflows/profit-analysis-slack.yml
name: Daily Profit Analysis with Slack

on:
  schedule:
    - cron: '0 8 * * 1-5'  # 平日朝8時
  workflow_dispatch:       # 手動実行も可能

jobs:
  profit-analysis:
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
    
    - name: Run Profit Analysis with Slack
      env:
        DB_DSN: ${{ secrets.DB_DSN }}
        SLACK_HOOK: ${{ secrets.SLACK_HOOK }}
      run: |
        cd cmd/roo-code-profit-trend-display
        ./bin/profit-trend-display -days 7 -summary -slack
    
    - name: Generate Detailed Report
      env:
        DB_DSN: ${{ secrets.DB_DSN }}
      run: |
        cd cmd/roo-code-profit-trend-display
        ./bin/profit-trend-display -days 7 -width 100 -height 25 > detailed-report.txt
    
    - name: Upload Report Artifact
      uses: actions/upload-artifact@v3
      with:
        name: profit-analysis-report
        path: cmd/roo-code-profit-trend-display/detailed-report.txt
    
    - name: Notify Slack on Failure
      if: failure()
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: "🚨 GitHub Actions: 粗利分析ワークフローが失敗しました"
        webhook_url: ${{ secrets.SLACK_HOOK }}
        
    - name: Notify Slack on Success
      if: success()
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: "✅ GitHub Actions: 粗利分析ワークフローが正常に完了しました"
        webhook_url: ${{ secrets.SLACK_HOOK }}
```

### 3.2 Docker環境でのSlack統合

```dockerfile
# Dockerfile.slack-enabled
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN cd cmd/roo-code-profit-trend-display && \
    go build -o bin/profit-trend-display .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata curl bc
WORKDIR /root/

COPY --from=builder /app/cmd/roo-code-profit-trend-display/bin/profit-trend-display .
COPY scripts/docker-slack-entrypoint.sh .

RUN chmod +x docker-slack-entrypoint.sh

ENTRYPOINT ["./docker-slack-entrypoint.sh"]
```

Docker実行用エントリーポイント（Slack対応版）：

```bash
#!/bin/sh
# docker-slack-entrypoint.sh

# 環境変数のデフォルト値設定
DAYS=${DAYS:-7}
FORMAT=${FORMAT:-summary}
OUTPUT_DIR=${OUTPUT_DIR:-/reports}
ENABLE_SLACK=${ENABLE_SLACK:-true}

# 出力ディレクトリ作成
mkdir -p "${OUTPUT_DIR}"

# Slack設定チェック
if [ "${ENABLE_SLACK}" = "true" ] && [ -z "${SLACK_HOOK}" ]; then
    echo "WARNING: SLACK_HOOK environment variable is not set. Slack notifications disabled."
    ENABLE_SLACK=false
fi

# レポート生成
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
OUTPUT_FILE="${OUTPUT_DIR}/report-${TIMESTAMP}.txt"

case "${FORMAT}" in
    "summary")
        if [ "${ENABLE_SLACK}" = "true" ]; then
            ./profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" -summary -slack > "${OUTPUT_FILE}"
        else
            ./profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" -summary > "${OUTPUT_FILE}"
        fi
        ;;
    "detailed")
        ./profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" -width 100 -height 25 > "${OUTPUT_FILE}"
        
        # 詳細レポート完了のSlack通知
        if [ "${ENABLE_SLACK}" = "true" ] && [ $? -eq 0 ]; then
            curl -X POST -H 'Content-type: application/json' \
                 --data "{
                     \"text\": \"📄 詳細粗利レポート生成完了\",
                     \"attachments\": [{
                         \"color\": \"good\",
                         \"title\": \"レポート情報\",
                         \"fields\": [{
                             \"title\": \"対象期間\",
                             \"value\": \"過去${DAYS}日間\",
                             \"short\": true
                         }, {
                             \"title\": \"生成時刻\",
                             \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                             \"short\": true
                         }, {
                             \"title\": \"ファイル名\",
                             \"value\": \"report-${TIMESTAMP}.txt\",
                             \"short\": false
                         }]
                     }]
                 }" \
                 "${SLACK_HOOK}" 2>/dev/null
        fi
        ;;
    *)
        # カスタム実行
        ./profit-trend-display -dsn "${DB_DSN}" -days "${DAYS}" "${@}" > "${OUTPUT_FILE}"
        ;;
esac

RESULT=$?

# 実行結果の通知
if [ "${ENABLE_SLACK}" = "true" ]; then
    if [ ${RESULT} -eq 0 ]; then
        echo "✅ Dockerコンテナでの粗利分析が正常に完了しました"
    else
        curl -X POST -H 'Content-type: application/json' \
             --data "{
                 \"text\": \"❌ Dockerコンテナでの粗利分析が失敗しました\",
                 \"attachments\": [{
                     \"color\": \"danger\",
                     \"title\": \"エラー情報\",
                     \"fields\": [{
                         \"title\": \"終了コード\",
                         \"value\": \"${RESULT}\",
                         \"short\": true
                     }, {
                         \"title\": \"実行時刻\",
                         \"value\": \"$(date '+%Y-%m-%d %H:%M:%S')\",
                         \"short\": true
                     }]
                 }]
             }" \
             "${SLACK_HOOK}" 2>/dev/null
    fi
fi

exit ${RESULT}
```

## 4. Kubernetes CronJob with Slack

```yaml
# k8s-cronjob-slack.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: profit-analysis-slack
spec:
  schedule: "0 8 * * 1-5"  # 平日朝8時
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: profit-trend-display
            image: your-registry/profit-trend-display:slack-enabled
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
            - name: DAYS
              value: "7"
            - name: FORMAT
              value: "summary"
            - name: ENABLE_SLACK
              value: "true"
            volumeMounts:
            - name: report-storage
              mountPath: /reports
          volumes:
          - name: report-storage
            persistentVolumeClaim:
              claimName: report-pvc
          restartPolicy: OnFailure

---
# Slack Webhook Secret
apiVersion: v1
kind: Secret
metadata:
  name: slack-secret
type: Opaque
data:
  webhook-url: <base64-encoded-slack-webhook-url>
```

## 5. 監視とログ管理（Slack統合版）

### 5.1 ログ監視とSlack通知

```bash
#!/bin/bash

# ログ監視とSlack通知スクリプト

LOG_DIR="/var/log/profit-trends"
TODAY=$(date +%Y%m%d)
ALERT_THRESHOLD=5  # エラー件数の閾値

# 今日のエラーログチェック
ERROR_COUNT=$(find "${LOG_DIR}" -name "*-${TODAY}.log" -exec grep -c "ERROR" {} \; 2>/dev/null | awk '{sum+=$1} END {print sum+0}')

if [ "${ERROR_COUNT}" -gt "${ALERT_THRESHOLD}" ]; then
    # エラー多発時のSlack通知
    if [ -n "${SLACK_HOOK}" ]; then
        curl -X POST -H 'Content-type: application/json' \
             --data "{
                 \"text\": \"⚠️ 粗利分析システム: エラー多発警告\",
                 \"attachments\": [{
                     \"color\": \"warning\",
                     \"title\": \"ログ監視アラート\",
                     \"fields\": [{
                         \"title\": \"エラー件数\",
                         \"value\": \"${ERROR_COUNT}件\",
                         \"short\": true
                     }, {
                         \"title\": \"閾値\",
                         \"value\": \"${ALERT_THRESHOLD}件\",
                         \"short\": true
                     }, {
                         \"title\": \"監視日\",
                         \"value\": \"${TODAY}\",
                         \"short\": true
                     }],
                     \"text\": \"システム管理者による確認が必要です。\"
                 }]
             }" \
             "${SLACK_HOOK}"
    fi
fi

# ディスク容量チェック
DISK_USAGE=$(df /var/reports/profit-trends | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "${DISK_USAGE}" -gt 85 ]; then
    if [ -n "${SLACK_HOOK}" ]; then
        curl -X POST -H 'Content-type: application/json' \
             --data "{
                 \"text\": \"💾 ディスク容量警告\",
                 \"attachments\": [{
                     \"color\": \"warning\",
                     \"title\": \"ストレージ監視アラート\",
                     \"fields\": [{
                         \"title\": \"使用率\",
                         \"value\": \"${DISK_USAGE}%\",
                         \"short\": true
                     }, {
                         \"title\": \"パス\",
                         \"value\": \"/var/reports/profit-trends\",
                         \"short\": true
                     }],
                     \"text\": \"古いファイルの削除や容量増設を検討してください。\"
                 }]
             }" \
             "${SLACK_HOOK}"
    fi
fi
```

## 6. まとめ

このSlack統合版自動化ガイドにより、以下が実現できます：

### 6.1 実現される機能

1. **リアルタイム通知**: 分析結果の即座の共有
2. **状況別通知**: 正常・警告・エラーに応じた適切な通知
3. **チーム連携**: Slackチャンネルでの情報共有
4. **自動アラート**: 閾値監視と緊急時通知
5. **運用監視**: システム状態のリアルタイム把握

### 6.2 運用メリット

- **迅速な対応**: 問題の早期発見と即座の通知
- **可視性向上**: チーム全体での状況共有
- **自動化**: 手作業の削減と確実な実行
- **履歴管理**: Slackでの通知履歴保持
- **拡張性**: 必要に応じた通知ルールの追加

### 6.3 セキュリティ考慮事項

- **Webhook URL保護**: 環境変数やSecretでの管理
- **機密情報除外**: 通知メッセージからの個人情報排除  
- **アクセス制御**: 適切なSlackチャンネル権限設定
- **ログ管理**: 通知履歴の適切な保管

これらの仕組みにより、手作業を最小限に抑えつつ、チーム全体で効率的な粗利分析と監視が可能になります。