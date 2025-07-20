# 出荷粗利計算プログラム Makefile

.PHONY: run build clean help deps

# デフォルトターゲット
all: deps run

# 依存関係をダウンロード
deps:
	@echo "=== 依存関係をダウンロード中 ==="
	go mod tidy
	@echo "依存関係のダウンロード完了"

# プログラムを実行
run: deps
	@echo "=== 出荷粗利計算プログラムを実行中 ==="
	go run main.go

# プログラムをビルド
build: deps
	@echo "=== プログラムをビルド中 ==="
	go build -o profit-calculator main.go
	@echo "ビルド完了: profit-calculator"

# バイナリファイルを削除
clean:
	@echo "=== クリーンアップ中 ==="
	rm -f profit-calculator
	@echo "クリーンアップ完了"

# ヘルプを表示
help:
	@echo "利用可能なコマンド:"
	@echo "  make deps   - 依存関係をダウンロード"
	@echo "  make run    - プログラムを実行"
	@echo "  make build  - プログラムをビルド"
	@echo "  make clean  - バイナリファイルを削除"
	@echo "  make help   - このヘルプを表示"

# テストを実行
test: deps
	@echo "=== テストを実行中 ==="
	go test ./...

# フォーマット
fmt:
	@echo "=== コードをフォーマット中 ==="
	go fmt .

# リント
lint: deps
	@echo "=== リントを実行中 ==="
	golangci-lint run 