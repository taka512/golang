# トラブルシューティングガイド

このドキュメントでは、`roo-code-profit-trend-display` の使用中に発生する可能性のある問題と解決方法を説明します。

## 1. よくある問題と解決方法

### 1.1 データベース関連の問題

#### 問題: データベース接続エラー

**症状**:
```
データベース接続エラー: dial tcp 127.0.0.1:3306: connect: connection refused
```

**原因と解決方法**:

1. **MySQLサーバーが起動していない**
   ```bash
   # サービス状態確認
   sudo systemctl status mysql
   
   # サービス開始
   sudo systemctl start mysql
   ```

2. **接続情報が間違っている**
   ```bash
   # 接続テスト
   mysql -h mysql.local -u root -p
   
   # 正しいDSNを指定
   ./bin/profit-trend-display -dsn "correct_user:password@tcp(correct_host:3306)/correct_db?parseTime=true"
   ```

3. **ファイアウォールによるブロック**
   ```bash
   # ポート3306の確認
   telnet mysql.local 3306
   
   # ファイアウォール設定確認
   sudo ufw status
   ```

4. **権限不足**
   ```sql
   -- 権限確認
   SHOW GRANTS FOR 'username'@'hostname';
   
   -- 必要な権限付与
   GRANT SELECT ON database.* TO 'username'@'hostname';
   ```

#### 問題: SQLエラー

**症状**:
```
データ取得エラー: Error 1146: Table 'database.companies' doesn't exist
```

**解決方法**:

1. **テーブル存在確認**
   ```sql
   SHOW TABLES LIKE 'companies';
   SHOW TABLES LIKE 'sales_daily_reports';
   SHOW TABLES LIKE 'cost_daily_reports';
   ```

2. **スキーマ確認**
   ```sql
   DESCRIBE companies;
   DESCRIBE sales_daily_reports;
   DESCRIBE cost_daily_reports;
   ```

3. **マイグレーション実行**
   ```bash
   # データベース移行スクリプトの確認
   ls database/migrations/
   
   # 必要なテーブルの作成
   mysql -u root -p database_name < database/migrations/000001_create_companies_table.up.sql
   ```

### 1.2 データ関連の問題

#### 問題: データが見つからない

**症状**:
```
指定された期間にデータが見つかりませんでした。
```

**診断手順**:

1. **データ存在確認**
   ```sql
   -- 売上データの確認
   SELECT COUNT(*) FROM sales_daily_reports 
   WHERE target_date >= CURDATE() - INTERVAL 30 DAY;
   
   -- 原価データの確認
   SELECT COUNT(*) FROM cost_daily_reports 
   WHERE target_date >= CURDATE() - INTERVAL 30 DAY;
   
   -- 明細データの確認
   SELECT COUNT(*) FROM sales_daily_report_items;
   SELECT COUNT(*) FROM cost_daily_report_items;
   ```

2. **日付範囲の確認**
   ```sql
   -- データの日付範囲確認
   SELECT 
       MIN(target_date) as min_date, 
       MAX(target_date) as max_date 
   FROM sales_daily_reports;
   ```

3. **会社・倉庫データの確認**
   ```sql
   -- マスターデータの確認
   SELECT COUNT(*) FROM companies;
   SELECT COUNT(*) FROM warehouse_bases;
   
   -- 組み合わせの確認
   SELECT DISTINCT company_id, warehouse_base_id 
   FROM sales_daily_reports;
   ```

**解決方法**:

1. **期間を長くする**
   ```bash
   # より長い期間で試行
   ./bin/profit-trend-display -days 90
   ./bin/profit-trend-display -days 365
   ```

2. **サンプルデータの投入**
   ```sql
   -- サンプルデータ投入例
   INSERT INTO sales_daily_reports (company_id, warehouse_base_id, target_date, sales_account_title_id) 
   VALUES (1, 1, CURDATE() - INTERVAL 1 DAY, 1);
   
   INSERT INTO sales_daily_report_items (sales_daily_report_id, quantity, price, amount) 
   VALUES (LAST_INSERT_ID(), 10, 100.0, 1000.0);
   ```

#### 問題: 異常なデータ値

**症状**:
```
統計情報:
  最大粗利:   999999999
  最小粗利:  -999999999
```

**診断と修正**:

1. **異常値の特定**
   ```sql
   -- 異常に大きな売上
   SELECT * FROM sales_daily_report_items 
   WHERE amount > 1000000;
   
   -- 異常に大きな原価
   SELECT * FROM cost_daily_report_items 
   WHERE cost_amount > 1000000;
   
   -- 負の値の確認
   SELECT * FROM sales_daily_report_items WHERE amount < 0;
   SELECT * FROM cost_daily_report_items WHERE cost_amount < 0;
   ```

2. **データ修正**
   ```sql
   -- 異常値の修正
   UPDATE sales_daily_report_items 
   SET amount = LEAST(amount, 100000) 
   WHERE amount > 100000;
   
   -- 負の値の修正
   UPDATE sales_daily_report_items 
   SET amount = 0 
   WHERE amount < 0;
   ```

### 1.3 パフォーマンス関連の問題

#### 問題: 実行が遅い

**症状**:
```bash
time ./bin/profit-trend-display -days 365
# real    2m30.123s
```

**診断手順**:

1. **実行時間の分析**
   ```bash
   # より詳細な時間測定
   time -v ./bin/profit-trend-display -days 30
   ```

2. **データ量の確認**
   ```sql
   -- レコード数確認
   SELECT 
       (SELECT COUNT(*) FROM sales_daily_reports) as sales_reports,
       (SELECT COUNT(*) FROM sales_daily_report_items) as sales_items,
       (SELECT COUNT(*) FROM cost_daily_reports) as cost_reports,
       (SELECT COUNT(*) FROM cost_daily_report_items) as cost_items;
   ```

3. **インデックス使用状況**
   ```sql
   -- クエリ実行計画の確認
   EXPLAIN SELECT 
       c.name, wb.name, sdr.target_date,
       SUM(sdri.amount) as sales_amount
   FROM companies c
   JOIN warehouse_bases wb 
   JOIN sales_daily_reports sdr ON c.id = sdr.company_id 
   JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
   WHERE sdr.target_date >= CURDATE() - INTERVAL 30 DAY
   GROUP BY c.id, wb.id, sdr.target_date;
   ```

**解決方法**:

1. **インデックスの追加**
   ```sql
   -- 日付インデックス
   CREATE INDEX idx_sales_reports_date ON sales_daily_reports(target_date);
   CREATE INDEX idx_cost_reports_date ON cost_daily_reports(target_date);
   
   -- 複合インデックス
   CREATE INDEX idx_sales_reports_comp ON sales_daily_reports(company_id, warehouse_base_id, target_date);
   ```

2. **データベース設定の最適化**
   ```sql
   -- インデックス統計情報の更新
   ANALYZE TABLE sales_daily_reports;
   ANALYZE TABLE cost_daily_reports;
   ANALYZE TABLE sales_daily_report_items;
   ANALYZE TABLE cost_daily_report_items;
   ```

3. **期間を短くする**
   ```bash
   # より短い期間で実行
   ./bin/profit-trend-display -days 7
   ./bin/profit-trend-display -days 14
   ```

#### 問題: メモリ不足

**症状**:
```
fatal error: runtime: out of memory
```

**解決方法**:

1. **メモリ使用量の確認**
   ```bash
   # メモリ使用量監視
   top -p $(pgrep profit-trend-display)
   
   # より詳細な情報
   /usr/bin/time -v ./bin/profit-trend-display -days 30
   ```

2. **データ量の削減**
   ```bash
   # より小さなチャートサイズ
   ./bin/profit-trend-display -width 40 -height 10
   
   # サマリーのみ
   ./bin/profit-trend-display -summary
   ```

3. **バッチ処理での対応**
   ```bash
   # 期間を分割して実行
   ./bin/profit-trend-display -days 7  > week1.txt
   ./bin/profit-trend-display -days 14 > week2.txt
   ```

### 1.4 表示関連の問題

#### 問題: 文字化け

**症状**:
```
[æ ªå¼ä¼ç¤¾A - æ±äº¬å庫] ç²å©æ¨ç§» (éå»30æ¥é)
```

**解決方法**:

1. **ターミナルの文字コード設定**
   ```bash
   # 現在の設定確認
   locale
   
   # UTF-8に設定
   export LANG=ja_JP.UTF-8
   export LC_ALL=ja_JP.UTF-8
   ```

2. **データベースの文字コード確認**
   ```sql
   SHOW VARIABLES LIKE 'character_set%';
   SHOW VARIABLES LIKE 'collation%';
   ```

3. **接続文字列の修正**
   ```bash
   ./bin/profit-trend-display -dsn "user:pass@tcp(host:3306)/db?parseTime=true&charset=utf8mb4"
   ```

#### 問題: チャートが崩れる

**症状**:
```
500 ┬     ●     ●     ●
400 ┤   ●   ●       ●   ●     ●
300 ┤ ●       ●   ●     ●
```

**解決方法**:

1. **ターミナルサイズの確認**
   ```bash
   # ターミナルサイズ確認
   tput cols
   tput lines
   
   # 適切なサイズに調整
   ./bin/profit-trend-display -width 60 -height 15
   ```

2. **フォント設定の確認**
   - 等幅フォントを使用
   - Unicode対応フォントを使用

3. **チャートサイズの調整**
   ```bash
   # 小さめのチャート
   ./bin/profit-trend-display -width 40 -height 10
   
   # 大きめのチャート（広いターミナル用）
   ./bin/profit-trend-display -width 120 -height 30
   ```

## 2. デバッグ手法

### 2.1 ログ出力の活用

#### 詳細ログの有効化

```bash
# 標準エラー出力をファイルに保存
./bin/profit-trend-display -days 30 2>debug.log

# 全出力をファイルに保存
./bin/profit-trend-display -days 30 >output.log 2>&1
```

#### ログ分析

```bash
# エラーメッセージの抽出
grep -i error debug.log

# データベース関連のログ
grep -i "database\|mysql\|sql" debug.log

# メモリ関連のログ
grep -i "memory\|allocation" debug.log
```

### 2.2 SQL クエリのデバッグ

#### データベース直接確認

```sql
-- アプリケーションが実行するクエリをテスト
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
    AND sdr.target_date BETWEEN '2024-06-21' AND '2024-07-20'
LEFT JOIN sales_daily_report_items sdri ON sdr.id = sdri.sales_daily_report_id
LEFT JOIN cost_daily_reports cdr ON c.id = cdr.company_id 
    AND wb.id = cdr.warehouse_base_id 
    AND cdr.target_date BETWEEN '2024-06-21' AND '2024-07-20'
LEFT JOIN cost_daily_report_items cdri ON cdr.id = cdri.cost_daily_report_id
WHERE (sdr.target_date IS NOT NULL OR cdr.target_date IS NOT NULL)
GROUP BY c.id, c.name, wb.id, wb.name, DATE(COALESCE(sdr.target_date, cdr.target_date))
ORDER BY c.name, wb.name, DATE(COALESCE(sdr.target_date, cdr.target_date))
LIMIT 10;
```

#### パフォーマンス分析

```sql
-- スロークエリログの確認
SET SESSION long_query_time = 0;
SET SESSION slow_query_log = 1;

-- 実行時間測定
SELECT BENCHMARK(1, (
    -- 上記のクエリをここに挿入
));
```

### 2.3 Goアプリケーションのデバッグ

#### デバッグビルド

```bash
# デバッグ情報付きでビルド
cd cmd/roo-code-profit-trend-display
go build -gcflags="all=-N -l" -o bin/profit-trend-display-debug .
```

#### プロファイリング

```go
// main.go にプロファイリングコードを追加（開発時のみ）
import (
    _ "net/http/pprof"
    "net/http"
    "log"
)

func main() {
    // プロファイリングサーバー開始（バックグラウンド）
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // 既存のmain処理
    // ...
}
```

```bash
# プロファイリング実行
go tool pprof http://localhost:6060/debug/pprof/profile

# メモリプロファイル
go tool pprof http://localhost:6060/debug/pprof/heap
```

### 2.4 環境別の問題診断

#### 開発環境

```bash
# 開発環境での詳細デバッグ
export DEBUG=true
./bin/profit-trend-display -days 7 -summary
```

#### 本番環境

```bash
# 本番環境での安全な診断
./bin/profit-trend-display -days 1 -summary -dsn "${PROD_READ_ONLY_DSN}"

# リソース使用量監視
htop &
./bin/profit-trend-display -days 30
```

## 3. パフォーマンス最適化

### 3.1 データベース最適化

#### インデックス戦略

```sql
-- 推奨インデックス
CREATE INDEX idx_sales_daily_reports_lookup 
ON sales_daily_reports(company_id, warehouse_base_id, target_date);

CREATE INDEX idx_cost_daily_reports_lookup 
ON cost_daily_reports(company_id, warehouse_base_id, target_date);

CREATE INDEX idx_sales_items_report 
ON sales_daily_report_items(sales_daily_report_id, amount);

CREATE INDEX idx_cost_items_report 
ON cost_daily_report_items(cost_daily_report_id, cost_amount);
```

#### データベース設定調整

```sql
-- MySQLパフォーマンス設定
SET GLOBAL innodb_buffer_pool_size = 1073741824;  -- 1GB
SET GLOBAL query_cache_type = 1;
SET GLOBAL query_cache_size = 67108864;  -- 64MB
```

### 3.2 アプリケーション最適化

#### メモリ使用量削減

```bash
# より小さなチャートサイズ
./bin/profit-trend-display -width 40 -height 10

# サマリーのみで実行
./bin/profit-trend-display -summary

# 短期間での実行
./bin/profit-trend-display -days 7
```

#### 並列処理の検討

将来的な改善として並列処理を検討：

```go
// 将来的な実装例（現在は未実装）
func (r *ProfitRepository) GetProfitTrendsParallel(startDate, endDate time.Time) {
    // 会社別に並列でデータ取得
    // チャンネルを使用した並行処理
}
```

## 4. 監視とアラート

### 4.1 ヘルスチェック

#### 基本ヘルスチェック

```bash
#!/bin/bash
# health-check.sh

# データベース接続確認
timeout 10 ./bin/profit-trend-display -days 1 -summary >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "OK: Application is healthy"
    exit 0
else
    echo "ERROR: Application health check failed"
    exit 1
fi
```

#### 詳細ヘルスチェック

```bash
#!/bin/bash
# detailed-health-check.sh

TEMP_FILE="/tmp/health-check-$$.txt"
EXIT_CODE=0

# 実行時間チェック
timeout 30 time ./bin/profit-trend-display -days 7 -summary > "${TEMP_FILE}" 2>&1
APP_EXIT_CODE=$?

if [ ${APP_EXIT_CODE} -ne 0 ]; then
    echo "ERROR: Application failed with exit code ${APP_EXIT_CODE}"
    EXIT_CODE=1
fi

# 出力内容チェック
if ! grep -q "分析完了" "${TEMP_FILE}"; then
    echo "ERROR: Application did not complete successfully"
    EXIT_CODE=1
fi

# データ数チェック
DATA_COUNT=$(grep "取得データ数:" "${TEMP_FILE}" | awk '{print $2}' | sed 's/件//')
if [ "${DATA_COUNT}" -eq 0 ]; then
    echo "WARNING: No data found"
fi

rm -f "${TEMP_FILE}"
exit ${EXIT_CODE}
```

### 4.2 自動復旧

#### サービス再起動スクリプト

```bash
#!/bin/bash
# auto-recovery.sh

MAX_RETRIES=3
RETRY_COUNT=0

while [ ${RETRY_COUNT} -lt ${MAX_RETRIES} ]; do
    if ./health-check.sh; then
        echo "Service is healthy"
        exit 0
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "Health check failed. Retry ${RETRY_COUNT}/${MAX_RETRIES}"
    
    # データベース接続プールのリセット等
    sleep 10
done

echo "Service recovery failed after ${MAX_RETRIES} retries"
exit 1
```

## 5. 緊急対応手順

### 5.1 緊急時の診断

#### 緊急診断チェックリスト

1. **システム基盤の確認**
   ```bash
   # システムリソース確認
   free -h
   df -h
   top
   
   # プロセス確認
   ps aux | grep profit-trend-display
   ```

2. **データベース状態確認**
   ```bash
   # MySQL接続確認
   mysqladmin ping
   
   # レプリケーション状態（該当する場合）
   mysql -e "SHOW SLAVE STATUS\G"
   ```

3. **ネットワーク確認**
   ```bash
   # データベースサーバーへの疎通確認
   ping database-server
   telnet database-server 3306
   ```

### 5.2 緊急回避策

#### データベース問題の場合

```bash
# 読み取り専用レプリカへの切り替え
./bin/profit-trend-display -dsn "${BACKUP_DB_DSN}" -days 7 -summary

# キャッシュされた結果の使用
cat /var/cache/profit-trends/last-successful-report.txt
```

#### アプリケーション問題の場合

```bash
# 最小限の実行
./bin/profit-trend-display -days 1 -summary

# 前回成功時の結果を表示
ls -la /var/reports/profit-trends/ | tail -5
```

### 5.3 エスカレーション

#### サポートチームへの報告テンプレート

```
件名: [緊急] 粗利推移分析システム障害

発生時刻: YYYY-MM-DD HH:MM:SS
影響範囲: [説明]
症状: [具体的なエラーメッセージや現象]

実行環境:
- OS: [OS情報]
- アプリケーションバージョン: [バージョン]
- データベース: [MySQL バージョン]

実行したコマンド:
./bin/profit-trend-display [オプション]

エラーメッセージ:
[エラーメッセージ全文]

実施した対応:
1. [実施内容1]
2. [実施内容2]

現在の状況:
[現在の状況説明]

添付ファイル:
- ログファイル
- 設定ファイル
- エラースクリーンショット
```

## 6. 予防保守

### 6.1 定期メンテナンス

#### 週次メンテナンス

```bash
#!/bin/bash
# weekly-maintenance.sh

# ログローテーション
logrotate /etc/logrotate.d/profit-trends

# 古いレポートファイルの削除
find /var/reports/profit-trends -name "*.txt" -mtime +30 -delete

# データベース統計情報更新
mysql -e "ANALYZE TABLE sales_daily_reports, cost_daily_reports, sales_daily_report_items, cost_daily_report_items;"

# ヘルスチェック実行
./health-check.sh
```

#### 月次メンテナンス

```bash
#!/bin/bash
# monthly-maintenance.sh

# データベース最適化
mysql -e "OPTIMIZE TABLE sales_daily_reports, cost_daily_reports;"

# インデックス再構築
mysql -e "ALTER TABLE sales_daily_reports ENGINE=InnoDB;"

# パフォーマンステスト
time ./bin/profit-trend-display -days 90 > /dev/null
```

### 6.2 監視メトリクス

#### 重要な監視項目

| メトリクス | 正常範囲 | 警告閾値 | 緊急閾値 |
|------------|----------|----------|----------|
| 実行時間 | < 10秒 | 30秒 | 60秒 |
| メモリ使用量 | < 512MB | 1GB | 2GB |
| エラー率 | 0% | 1% | 5% |
| データ取得件数 | > 0 | = 0 | N/A |

#### 監視スクリプトの例

```bash
#!/bin/bash
# monitoring.sh

METRICS_FILE="/var/log/profit-trends/metrics.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# 実行時間測定
START_TIME=$(date +%s)
./bin/profit-trend-display -days 7 -summary > /dev/null 2>&1
EXIT_CODE=$?
END_TIME=$(date +%s)
EXECUTION_TIME=$((END_TIME - START_TIME))

# メトリクス記録
echo "${TIMESTAMP},execution_time,${EXECUTION_TIME}" >> "${METRICS_FILE}"
echo "${TIMESTAMP},exit_code,${EXIT_CODE}" >> "${METRICS_FILE}"

# 閾値チェック
if [ ${EXECUTION_TIME} -gt 30 ]; then
    echo "WARNING: Execution time exceeded 30 seconds: ${EXECUTION_TIME}s"
fi

if [ ${EXIT_CODE} -ne 0 ]; then
    echo "ERROR: Application failed with exit code: ${EXIT_CODE}"
fi
```

---

このトラブルシューティングガイドを活用して、問題の迅速な解決と安定した運用を実現してください。