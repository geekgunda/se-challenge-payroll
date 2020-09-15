package payrollapi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

const timeReportEndpoint = "http://127.0.0.1:8081/timereport"
const payrollReportEndpoint = "http://127.0.0.1:8081/payrollreport"

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func shutdown() {
}

func setup() {
	// Initialize regexp
	InitTimeReportHandler()
	// Initialize DB Conn
	InitDBConn()
}

// TestErrTimeReport tests some error scenarios for timereport API
func TestErrTimeReport(t *testing.T) {
	var cases = []struct {
		name, method, contentType string
		body                      io.Reader
		expStatusCode             int
	}{
		{"GET-Req1", http.MethodGet, "", nil, http.StatusMethodNotAllowed},
		{"POST-Req1", http.MethodPost, "", nil, http.StatusBadRequest},
		{"POST-Req2", http.MethodPost, "multipart/form-data", nil, http.StatusBadRequest},
		{"POST-Req3", http.MethodPost, "multipart/form-data; boundary=xxx", nil, http.StatusBadRequest},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, timeReportEndpoint, tt.body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()
			var h TimeReport
			h.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != tt.expStatusCode {
				t.Errorf("Case %s failed", tt.name)
			}
		})
	}
}

// TestEndtoEndAPI tests end to end scenarios for the API
// First it uploads a file via time-report API.
// Next it validates that uploading same file throws error.
// Finally it fetches a payroll report via the other API
// and validates that the report returns the expected data.
func TestEndtoEndAPI(t *testing.T) {
	var cases = []struct {
		name, filePath string
		expStatusCode  int
	}{
		{"POST-multipart-1", "time-report-10.csv", http.StatusOK},
		{"POST-multipart-2", "time-report-10.csv", http.StatusBadRequest},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// Create multi-part-form HTTP Request object with input file
			// Ref: https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84
			// Ref: https://gist.github.com/andrewmilson/19185aab2347f6ad29f5
			file, err := os.Open(tt.filePath)
			if err != nil {
				t.Errorf("Case [%s] failed: %v", tt.name, err)
			}
			fi, err := file.Stat()
			if err != nil {
				t.Errorf("Case [%s] failed: %v", tt.name, err)
			}
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("timereport", fi.Name())
			if err != nil {
				t.Errorf("Case [%s] failed: %v", tt.name, err)
			}
			io.Copy(part, file)
			file.Close()
			writer.Close()

			// Invoke the handler and test assertions
			req := httptest.NewRequest(http.MethodPost, timeReportEndpoint, body)
			req.Header.Add("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()
			var h TimeReport
			h.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != tt.expStatusCode {
				t.Errorf("TimeReport case failed: %s | Resp: %v", tt.name, resp)
			}
		})
	}
	// Test the Payroll Report
	req := httptest.NewRequest(http.MethodGet, payrollReportEndpoint, nil)
	w := httptest.NewRecorder()
	var h PayrollReport
	h.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("PayrollReport case failed | Resp: %#v", resp)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("PayrollReport case filed | Err: %v", err)
	}
	var parsedResp PayrollReport
	err = json.Unmarshal(body, &parsedResp)
	if err != nil {
		t.Errorf("PayrollReport case filed | Err: %v", err)
	}
	expJsonBody := []byte(`{
		payrollReport: {
	    employeeReports: [
	      {
	        employeeId: 1,
	        payPeriod: {
	          startDate: "2020-01-01",
	          endDate: "2020-01-15"
	        },
	        amountPaid: "$300.00"
	      },
	      {
	        employeeId: 1,
	        payPeriod: {
	          startDate: "2020-01-16",
	          endDate: "2020-01-31"
	        },
	        amountPaid: "$80.00"
	      },
	      {
	        employeeId: 2,
	        payPeriod: {
	          startDate: "2020-01-16",
	          endDate: "2020-01-31"
	        },
	        amountPaid: "$90.00"
	      }
		];
	  }
	}`)
	var expResp PayrollReport
	err = json.Unmarshal(expJsonBody, &expResp)
	if reflect.DeepEqual(parsedResp, expResp) {
		t.Errorf("PayrollReport case failed. | Exp: %#v | Act: %#v", expResp, parsedResp)
	}

}
