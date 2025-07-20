package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

type Sample struct {
	Id      int
	Icol    int
	V16col  string
	V64col  string
	V255col string
	V256col string
}

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
	for i := 1; i < 10000; i++ {
		var sample = Sample{}
		err = db.QueryRow(fmt.Sprintf("SELECT s2.* FROM sample1 s1 INNER JOIN sample2 s2 ON s1.%s = s2.%s WHERE s1.%s = ?",
			*f, *f, *f), strconv.Itoa(i)).Scan(&sample.Id, &sample.Icol, &sample.V16col, &sample.V64col, &sample.V255col, &sample.V256col)
		if err != nil {
			fmt.Println("Query error.")
			panic(err)
		}
	}
}
