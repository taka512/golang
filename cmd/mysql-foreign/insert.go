package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	f := flag.String("table", "sample1", "target table")
	flag.Parse()
	fmt.Printf("table:%s\n", *f)

	db, err := sql.Open("mysql", "root:mypass@(mysql.local:3306)/test?autocommit=0")
	if err != nil {
		panic(err)
	}

	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	ins1, _ := db.Prepare(fmt.Sprintf("INSERT INTO %s_oya(id, name) VALUES(?,?)", *f))
	stmt1 := tx.Stmt(ins1)
	ins2, _ := db.Prepare(fmt.Sprintf("INSERT INTO %s_ko(%s_oya_id, name) VALUES(?,?)", *f, *f))
	stmt2 := tx.Stmt(ins2)
	for i := 1; i <= 10000; i++ {
		_, err := stmt1.Exec(i, fmt.Sprintf("name:%s", strconv.Itoa(i)))
		if err != nil {
			panic(err)
		}
		for j := 1; j <= 10; j++ {
			_, err := stmt2.Exec(i, fmt.Sprintf("name:%s-%s", strconv.Itoa(i), strconv.Itoa(j)))
			if err != nil {
				panic(err)
			}

		}
	}
	tx.Commit()
}
