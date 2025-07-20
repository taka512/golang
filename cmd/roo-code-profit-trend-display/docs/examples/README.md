# 使用例とサンプルコード

このディレクトリには `roo-code-profit-trend-display` の実践的な使用例とサンプルコードが含まれています。

## 📁 ディレクトリ構成

```
examples/
├── README.md                 # このファイル
├── basic-usage/             # 基本的な使用例
│   ├── simple-run.md        # 基本実行例
│   ├── custom-settings.md   # カスタム設定例
│   └── output-samples/      # 出力サンプル
├── advanced-usage/          # 高度な使用例
│   ├── batch-processing.md  # バッチ処理例
│   ├── automated-reports.md # 自動レポート生成
│   └── integration.md       # システム統合例
├── scripts/                 # 実用的なスクリプト
│   ├── daily-report.sh      # 日次レポート生成
│   ├── weekly-summary.sh    # 週次サマリー
│   └── alert-system.sh      # アラートシステム
├── data-samples/            # サンプルデータ
│   ├── sample-data.sql      # テストデータ投入
│   └── realistic-data.sql   # より現実的なデータ
└── troubleshooting/         # トラブルシューティング例
    ├── common-errors.md     # よくあるエラーと対処
    ├── performance-tips.md  # パフォーマンス改善
    └── debugging.md         # デバッグ手法
```

## 🚀 クイックスタート例

### 基本実行
```bash
# 最もシンプルな実行（過去30日間）
./bin/profit-trend-display

# 過去1週間の分析
./bin/profit-trend-display 7

# サマリーのみ表示
./bin/profit-trend-display -summary
```

### よく使われる組み合わせ
```bash
# 大きなグラフで詳細分析
./bin/profit-trend-display -days 14 -width 100 -height 25

# シンプルなサマリーレポート
./bin/profit-trend-display -days 30 -summary -grid=false

# 本番環境での実行例
./bin/profit-trend-display \
  -dsn "prod_user:${DB_PASSWORD}@tcp(prod-db:3306)/prod_db?parseTime=true" \
  -days 7 \
  -summary
```

## 📋 使用例カテゴリ

### 1. 基本的な使用例
- **Simple Run**: 最も基本的な実行方法
- **Custom Settings**: 設定をカスタマイズした実行
- **Output Formats**: 異なる出力形式の例

### 2. 高度な使用例
- **Batch Processing**: 複数期間の一括処理
- **Automated Reports**: 定期的な自動レポート生成
- **System Integration**: 他システムとの連携

### 3. 実用的なスクリプト
- **Daily Reports**: 日次業務で使えるスクリプト
- **Monitoring**: 監視・アラート用スクリプト
- **Data Management**: データ管理支援スクリプト

### 4. トラブルシューティング
- **Error Handling**: エラー対処の実例
- **Performance**: パフォーマンス最適化
- **Debugging**: 問題調査の手法

## 🎯 推奨使用パターン

### 日次業務
```bash
# 毎朝の業績確認
./bin/profit-trend-display -days 7 -summary

# 週次レビュー用詳細分析
./bin/profit-trend-display -days 30 -width 80 -height 20
```

### 月次レポート
```bash
# 月次サマリー（前月分析）
./bin/profit-trend-display -days 30 -summary > monthly_report.txt

# 四半期比較用データ
./bin/profit-trend-display -days 90 -width 120 -height 30
```

### 問題調査
```bash
# 特定期間の詳細分析
./bin/profit-trend-display -days 14 -grid=true -stats=true

# パフォーマンス問題調査
time ./bin/profit-trend-display -days 365
```

## 🔧 環境別設定例

### 開発環境
```bash
export DB_DSN="root:devpass@tcp(localhost:3306)/dev_db?parseTime=true"
./bin/profit-trend-display -dsn "$DB_DSN" -days 7
```

### ステージング環境
```bash
export DB_DSN="stage_user:${STAGE_PASSWORD}@tcp(stage-db:3306)/stage_db?parseTime=true"
./bin/profit-trend-display -dsn "$DB_DSN" -days 30 -summary
```

### 本番環境
```bash
export DB_DSN="prod_readonly:${PROD_PASSWORD}@tcp(prod-db.internal:3306)/production?parseTime=true"
./bin/profit-trend-display -dsn "$DB_DSN" -days 7 -summary 2>&1 | tee /var/log/profit-analysis.log
```

## 📊 出力例プレビュー

### 標準出力
```
=== 粗利推移表示プログラム ===
分析期間: 過去7日間
接続先: root:***@tcp(mysql.local:3306)/sample_mysql?parseTime=true

対象期間: 2024-07-14 から 2024-07-20 まで

=== 分析結果 ===
対象組織数: 3

(1/3) [株式会社A - 東京倉庫] 粗利推移 (過去7日間)
```

### サマリー出力
```
=== 粗利推移サマリー ===

全体統計:
  合計粗利:     850000
  平均粗利:      12143
  最大粗利:      45000 (07/18)
  最小粗利:        500 (07/16)
  対象組織数: 3

組織別サマリー:
  株式会社A - 東京倉庫: 合計=450000, 平均=64286
  株式会社A - 大阪倉庫: 合計=250000, 平均=35714
  株式会社B - 福岡倉庫: 合計=150000, 平均=21429
```

## 🚨 注意事項

### データベース接続
- 本番環境では読み取り専用ユーザーを使用
- パスワードは環境変数で管理
- 接続タイムアウトを考慮した設定

### パフォーマンス
- 大量データ分析時はメモリ使用量に注意
- 長期間分析（90日以上）は実行時間を考慮
- 本番環境では業務時間外の実行を推奨

### セキュリティ
- ログファイルにDB接続情報が含まれないよう注意
- 実行権限の適切な設定
- 出力結果の機密性を考慮した取り扱い

---

各ディレクトリの詳細については、それぞれのREADME.mdファイルを参照してください。