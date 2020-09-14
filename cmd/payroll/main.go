package main

import (
	"database/sql"
	"log"
	"net/http"

	payrollapi "github.com/geekgunda/se-challenge-payroll"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	payrollapi.InitTimeReportHandler()
	dbUser, dbPass := "root", "brutepass"
	initDBConn(payrollapi.DBName, dbUser, dbPass)
	startServer()
}

func startServer() {
	var timeReportHandler payrollapi.TimeReport
	var payrollReportHandler payrollapi.PayrollReport
	http.Handle("/timereport", &timeReportHandler)
	http.Handle("/payrollreport", &payrollReportHandler)
	log.Println("Starting HTTP Server")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Println("Error while listenAndServe: ", err)
	}
}

func initDBConn(dbName, dbUser, dbPass string) {
	connStr := dbUser + ":" + dbPass + "@tcp(127.0.0.1:3306)/" + dbName
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to db [%s] for user [%s]: %v", dbName, dbUser, err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping db: ", err)
	}
	payrollapi.InitDBConn(db)
}
