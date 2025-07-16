package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	f := flag.String("table", "sample1", "target table")
	flag.Parse()
	fmt.Printf("table:%s\n", *f)

	db, err := sql.Open("mysql", "root:mypass@(go.mysql.local:3307)/test")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	del1, err := db.Prepare(fmt.Sprintf("DELETE FROM %s_ko", *f))
	if err != nil {
		panic(err)
	}
	stmt1 := tx.Stmt(del1)
	stmt1.Exec()
	del2, err := db.Prepare(fmt.Sprintf("DELETE FROM %s_oya", *f))
	if err != nil {
		panic(err)
	}
	stmt2 := tx.Stmt(del2)
	stmt2.Exec()
	tx.Commit()
}
