package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MasterData struct {
	Companies          []IDName
	Warehouses        []IDName
	SalesAccountTitles []IDName
	CostAccountTitles  []IDName
}

type IDName struct {
	ID   int
	Name string
}

func main() {
	// データベース接続
	db, err := sql.Open("mysql", "root:mypass@(mysql.local:3306)/sample_mysql?parseTime=true")
	if err != nil {
		log.Fatal("データベース接続エラー:", err)
	}
	defer db.Close()

	// マスターデータ取得
	masterData, err := getMasterData(db)
	if err != nil {
		log.Fatal("マスターデータ取得エラー:", err)
	}

	// 1週間分の日付を生成（今日から7日前まで）
	dates := generateWeekDates()

	fmt.Printf("=== データ挿入開始 ===\n")
	fmt.Printf("対象期間: %s ～ %s\n", dates[0].Format("2006-01-02"), dates[len(dates)-1].Format("2006-01-02"))
	fmt.Printf("会社数: %d, 倉庫数: %d, 売上科目数: %d, 原価科目数: %d\n", 
		len(masterData.Companies), len(masterData.Warehouses), 
		len(masterData.SalesAccountTitles), len(masterData.CostAccountTitles))

	// 売上データ挿入
	salesCount, err := insertSalesData(db, masterData, dates)
	if err != nil {
		log.Fatal("売上データ挿入エラー:", err)
	}
	fmt.Printf("売上データ挿入完了: %d件のレポート\n", salesCount)

	// 原価データ挿入
	costCount, err := insertCostData(db, masterData, dates)
	if err != nil {
		log.Fatal("原価データ挿入エラー:", err)
	}
	fmt.Printf("原価データ挿入完了: %d件のレポート\n", costCount)

	fmt.Printf("=== データ挿入完了 ===\n")
}

func getMasterData(db *sql.DB) (*MasterData, error) {
	masterData := &MasterData{}

	// 会社データ取得
	rows, err := db.Query("SELECT id, name FROM companies ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item IDName
		err := rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, err
		}
		masterData.Companies = append(masterData.Companies, item)
	}

	// 倉庫データ取得
	rows, err = db.Query("SELECT id, name FROM warehouse_bases ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item IDName
		err := rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, err
		}
		masterData.Warehouses = append(masterData.Warehouses, item)
	}

	// 売上科目データ取得
	rows, err = db.Query("SELECT id, name FROM sales_account_titles ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item IDName
		err := rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, err
		}
		masterData.SalesAccountTitles = append(masterData.SalesAccountTitles, item)
	}

	// 原価科目データ取得
	rows, err = db.Query("SELECT id, name FROM cost_account_titles ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item IDName
		err := rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, err
		}
		masterData.CostAccountTitles = append(masterData.CostAccountTitles, item)
	}

	return masterData, nil
}

func generateWeekDates() []time.Time {
	var dates []time.Time
	now := time.Now()
	
	// 今日から7日前まで
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dates = append(dates, date)
	}
	
	return dates
}

func insertSalesData(db *sql.DB, masterData *MasterData, dates []time.Time) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 売上レポート挿入準備
	reportStmt, err := tx.Prepare(`
		INSERT INTO sales_daily_reports (company_id, warehouse_base_id, target_date, sales_account_title_id) 
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer reportStmt.Close()

	// 売上レポート明細挿入準備
	itemStmt, err := tx.Prepare(`
		INSERT INTO sales_daily_report_items (sales_daily_report_id, size, quantity, price, amount) 
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer itemStmt.Close()

	reportCount := 0
	
	for _, date := range dates {
		for _, company := range masterData.Companies {
			for _, warehouse := range masterData.Warehouses {
				for _, accountTitle := range masterData.SalesAccountTitles {
					// 売上レポート挿入
					result, err := reportStmt.Exec(company.ID, warehouse.ID, date.Format("2006-01-02"), accountTitle.ID)
					if err != nil {
						return 0, err
					}
					
					reportID, err := result.LastInsertId()
					if err != nil {
						return 0, err
					}
					
					// 各レポートに対して2-4個の明細を挿入
					itemCount := rand.Intn(3) + 2 // 2-4個
					sizes := []string{"S", "M", "L", "XL"}
					
					for i := 0; i < itemCount; i++ {
						size := sizes[i%len(sizes)]
						quantity := rand.Intn(100) + 10 // 10-109個
						price := float64(rand.Intn(5000)+1000) / 100.0 // 10.00-59.99円
						amount := float64(quantity) * price
						
						_, err := itemStmt.Exec(reportID, size, quantity, price, amount)
						if err != nil {
							return 0, err
						}
					}
					
					reportCount++
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return reportCount, nil
}

func insertCostData(db *sql.DB, masterData *MasterData, dates []time.Time) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 原価レポート挿入準備
	reportStmt, err := tx.Prepare(`
		INSERT INTO cost_daily_reports (company_id, warehouse_base_id, target_date, cost_account_title_id) 
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer reportStmt.Close()

	// 原価レポート明細挿入準備
	itemStmt, err := tx.Prepare(`
		INSERT INTO cost_daily_report_items (cost_daily_report_id, size, quantity, cost_price, cost_amount) 
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer itemStmt.Close()

	reportCount := 0
	
	for _, date := range dates {
		for _, company := range masterData.Companies {
			for _, warehouse := range masterData.Warehouses {
				for _, accountTitle := range masterData.CostAccountTitles {
					// 原価レポート挿入
					result, err := reportStmt.Exec(company.ID, warehouse.ID, date.Format("2006-01-02"), accountTitle.ID)
					if err != nil {
						return 0, err
					}
					
					reportID, err := result.LastInsertId()
					if err != nil {
						return 0, err
					}
					
					// 各レポートに対して2-4個の明細を挿入
					itemCount := rand.Intn(3) + 2 // 2-4個
					sizes := []string{"S", "M", "L", "XL"}
					
					for i := 0; i < itemCount; i++ {
						size := sizes[i%len(sizes)]
						quantity := rand.Intn(100) + 10 // 10-109個
						costPrice := float64(rand.Intn(3000)+500) / 100.0 // 5.00-34.99円（売上より安く）
						costAmount := float64(quantity) * costPrice
						
						_, err := itemStmt.Exec(reportID, size, quantity, costPrice, costAmount)
						if err != nil {
							return 0, err
						}
					}
					
					reportCount++
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return reportCount, nil
}
