package payrollapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const reportDateFormat = "2006-01-02"

// CustomDate wraps time.Time and support custom format
// during json marshal and unmarshal
type CustomDate struct {
	time.Time
}

func (d *CustomDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		d.Time = time.Time{}
		return
	}
	d.Time, err = time.Parse(reportDateFormat, s)
	return
}

func (d *CustomDate) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format(reportDateFormat))), nil
}

// PayrollReport holds response contract for payroll report API
type PayrollReport struct {
	EmpReports []EmployeeReport `json:"employeeReports"`
}

// EmployeeReport holds employee level payroll report
type EmployeeReport struct {
	EmpID        int           `json:"employeeId"`
	Period       PayPeriod     `json:"payPeriod"`
	Amount       string        `json:"amountPaid"`
	AmountPaid   float64       `json:"-"`
	WorkDuration time.Duration `json:"-"`
	Group        JobGroup      `json:"-"`
}

// PayPeriod holds details of a specific pay period
type PayPeriod struct {
	StartDate CustomDate `json:"startDate"`
	EndDate   CustomDate `json:"endDate"`
}

// ServeHTTP handles HTTP requests to fetch payroll report
func (h *PayrollReport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h = new(PayrollReport)
	// Only GET requests allowed
	if r.Method != http.MethodGet {
		writeErrResponse(w, http.StatusMethodNotAllowed, nil, "Invalid method: "+r.Method)
		return
	}
	log.Println("Processing Payroll Report request")

	// Fetch all unique year-month combinations for existing time-reports
	var workPeriods []time.Time
	var err error
	if workPeriods, err = GetWorkPeriods(); err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err, "")
		return
	}
	// Fetch aggregate workHours for all unique year-month combinations
	for _, t := range workPeriods {
		// First pay period
		p := getFirstPayPeriodOfMonth(t)
		log.Println("Fetching report for period: ", p)
		if err = GetPayrollReport(h, p); err != nil {
			writeErrResponse(w, http.StatusInternalServerError, err, "")
			return
		}
		//log.Println("Payroll Report: ", h.EmpReports)

		// Second pay period
		p = getLastPayPeriodOfMonth(t)
		log.Println("Fetching report for period: ", p)
		if err = GetPayrollReport(h, p); err != nil {
			writeErrResponse(w, http.StatusInternalServerError, err, "")
			return
		}
		//log.Println("Payroll Report: ", h.EmpReports)
	}
	h.CalculatePayment()
	var payload []byte
	if payload, err = json.Marshal(h); err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err, "")
		return
	}
	// Form the HTTP Response
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// CalculatePayment takes work hours and job groups and computes amount to be paid for entire payroll report
func (h *PayrollReport) CalculatePayment() error {
	for i, r := range h.EmpReports {
		h.EmpReports[i].AmountPaid = r.WorkDuration.Hours() * float64(getWageForJobGroup(r.Group))
		h.EmpReports[i].Amount = fmt.Sprintf("$%.2f", h.EmpReports[i].AmountPaid)
		log.Println("Processed emp record: ", r)
	}
	return nil
}

// getFirstPayPeriodOfMonth gives time frame for 1-15 of the given month
// Ref: https://stackoverflow.com/a/55215625/6632880
func getFirstPayPeriodOfMonth(t time.Time) (p PayPeriod) {
	p.StartDate = CustomDate{t.AddDate(0, 0, -t.Day()+1)}
	p.EndDate = CustomDate{p.StartDate.AddDate(0, 0, 14)}
	return
}

// getLastPayPeriodOfMonth gives time frame for 16-last-day-of-month for given month
func getLastPayPeriodOfMonth(t time.Time) (p PayPeriod) {
	p.StartDate = CustomDate{t.AddDate(0, 0, -t.Day()+16)}
	p.EndDate = CustomDate{p.StartDate.AddDate(0, 1, -p.StartDate.Day())}
	return
}
