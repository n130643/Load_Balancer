package databaseLayer

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

//Have to add required table and db initialization before service starts // Future work.

func initDB() (*sql.DB, error) {

	db, err := sql.Open("mysql", "Gopi:Gopi@tcp(load-balancer-db:3306)/sample_db")

	if err != nil {
		return nil, errors.New(fmt.Sprintln("failed to open DB connection", err))
	}

	err = db.Ping()

	if err != nil {
		fmt.Println("error", err.Error())
		return nil, err
	}
	return db, nil
}

func ExeucteInserQuery(q string) (*sql.Rows, error) {
	rows, er := Db.Query(q)

	if er != nil {
		return nil, er
	}

	return rows, nil
}

func GetMaxOrder() (int, error) {
	var order int
	q := "select max(ll_order) from load_balancers"
	rws, er := ExeucteInserQuery(q)

	if er != nil {
		return 0, errors.New("failed to get max order")
	}

	for rws.Next() {
		rws.Scan(&order)
		return order, nil
	}
	return 0, nil
}

func init() {

	db, er := initDB()

	Db = db

	if er != nil {
		fmt.Println(er)
	}

}
