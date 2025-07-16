package main

import (
	"database/sql"
	"flag"
	"fmt"
    "time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	f := flag.String("table", "test_reports", "target table")
	flag.Parse()
	fmt.Printf("table:%s\n", *f)

	db, err := sql.Open("mysql", "root:mypass@(go.mysql.local:3307)/test?autocommit=0")
	if err != nil {
		panic(err)
	}

	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	ins, _ := db.Prepare(fmt.Sprintf("INSERT INTO %s(target_date, user_id, quantity, amount, created) VALUES(?, ?, ?, ?, now())", *f))
	stmt := tx.Stmt(ins)
	for month := 1; month <= 12; month++ {
    	for day := 1; day <= 28; day++ {
           t := time.Date(2023, time.Month(month), day, 0, 0, 0, 0, time.Local)
    		for id := 1; id <= 1000; id++ {
    	    	_, err := stmt.Exec(t.Format("2006/01/02"), id, day, day * month)
                if err != nil {
                 	panic(err)
    	    	}
            }
		}
	}
	tx.Commit()
}
