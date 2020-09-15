package main

import (
	"log"
	"net/http"

	payrollapi "github.com/geekgunda/se-challenge-payroll"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	payrollapi.InitTimeReportHandler()
	if err := payrollapi.InitDBConn(); err != nil {
		log.Fatalf("Error initiating DB conn: %v", err)
	}
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
