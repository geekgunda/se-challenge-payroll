package payrollapi

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

const dbName = "payroll"
const dbUser = "root"
const dbPass = "brutepass"
const timeReportTable = "timereport"
const timeReportItemTable = "timereportitem"
const mysqlDateFormat = "2006-01-02"
const workPeriodDateFormat = "2006-01"

func InitDBConn() (err error) {
	connStr := dbUser + ":" + dbPass + "@tcp(127.0.0.1:3306)/" + dbName
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		return fmt.Errorf("Failed to connect to db [%s] for user [%s]: %v", dbName, dbUser, err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("Failed to ping DB: %v", err)
	}
	return nil
}

// InsertTimeReport inserts a new entry for time-report
// It also reports whether the report ID has already been processed before
func InsertTimeReport(reportID string) (bool, error) {
	var isDuplicate bool
	_, err := db.Exec("INSERT INTO "+timeReportTable+"(report_id) VALUES(?)", reportID)
	if err != nil {
		if merr, ok := err.(*mysql.MySQLError); ok {
			if merr.Number == 1062 {
				isDuplicate = true
			}
			log.Println("MySQL Error number: ", merr.Number, " msg: ", merr.Message)
			return isDuplicate, merr
		}
		return isDuplicate, err
	}
	return isDuplicate, nil
}

// InsertTimeReportItem adds a new entry from time-report
func InsertTimeReportItem(reportID string, item *TimeReportItem) error {
	stmt := "INSERT INTO " + timeReportItemTable
	stmt += "(report_id, emp_id, work_date, work_hours, job_group) "
	stmt += "VALUES(?,?,?,?,?)"
	_, err := db.Exec(stmt, reportID, item.EmpID, item.Date.Format(mysqlDateFormat), item.Hours.String(), item.Group)
	return err
}

// GetWorkPeriods fetches all unique year-month pairs across all time-reports
func GetWorkPeriods() ([]time.Time, error) {
	// Get ordered distinct year and months to be processed
	stmt := `select distinct(date_format(work_date, "%Y-%m")) as input from timereportitem order by input asc`
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	var res []time.Time
	defer rows.Close()
	for rows.Next() {
		var input string
		if err = rows.Scan(&input); err != nil {
			return nil, err
		}
		var t time.Time
		if t, err = time.Parse(workPeriodDateFormat, input); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// GetPayrollReport fetches details of time report for given period
// It appends these details to the passed PayrollReport instance
// Results are ordered by employee ID
func GetPayrollReport(r *PayrollReport, p PayPeriod) error {
	stmt := `select emp_id, work_date, work_hours, job_group from ` + timeReportItemTable
	stmt += ` where work_date >= '` + p.StartDate.Format(mysqlDateFormat) + `' and `
	stmt += `work_date <= '` + p.EndDate.Format(mysqlDateFormat) + `' order by emp_id`
	log.Println("Stmt: ", stmt)
	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}
	var emp EmployeeReport
	defer rows.Close()
	for rows.Next() {
		var empID int
		var workDate, workHours, jobGroup string
		if err = rows.Scan(&empID, &workDate, &workHours, &jobGroup); err != nil {
			return err
		}
		log.Println("Scanned row: ", empID, workDate, workHours, jobGroup)
		// If we are seeing this employee for the first time
		if emp.EmpID == 0 || emp.EmpID != empID {
			// Add details of previous employee in results first
			if emp.EmpID != 0 {
				r.EmpReports = append(r.EmpReports, emp)
			}
			emp = EmployeeReport{}
			emp.EmpID = empID
			emp.Period = p
			if emp.Group, err = getJobGroup(jobGroup); err != nil {
				return err
			}
			if emp.WorkDuration, err = time.ParseDuration(workHours); err != nil {
				return err
			}
			log.Println("Found new row: ", emp)
			continue
		}
		// If this is an existing employee, just add hours
		var hours time.Duration
		if hours, err = time.ParseDuration(workHours); err != nil {
			return err
		}
		emp.WorkDuration += hours
		log.Println("Updated row: ", emp)
	}
	// Append the last entry
	if emp.EmpID != 0 {
		r.EmpReports = append(r.EmpReports, emp)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}
