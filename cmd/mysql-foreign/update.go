package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	f := flag.String("col", "icol", "col")
	flag.Parse()
	fmt.Printf("col:%s\n", *f)

	db, err := sql.Open("mysql", "root:mypass@(mysql.local:3306)/test")
	if err != nil {
		fmt.Println("db error.")
		panic(err)
	}

	defer db.Close()
	upd, err := db.Prepare(fmt.Sprintf("UPDATE sample1 SET %s = ? WHERE %s = ?", *f, *f))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	for i := 1; i < 100000; i++ {
		upd.Exec(i+100000, strconv.Itoa(i))
		upd.Exec(i, strconv.Itoa(i+100000))
	}
}
