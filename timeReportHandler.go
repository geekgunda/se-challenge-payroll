package payrollapi

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// timeReportDateFormat used in time-report
const timeReportDateFormat = "2/1/2006"

var timeReportFileRegex *regexp.Regexp

func InitTimeReportHandler() {
	timeReportFileRegex = regexp.MustCompile(`time-report-(\d+).csv`)
}

// JobGroup classifies jobs
type JobGroup string

const (
	JobGroupA = "A"
	JobGroupB = "B"
	JobGroupUnknown
)

func getJobGroup(s string) (JobGroup, error) {
	switch s {
	case "A":
		return JobGroupA, nil
	case "B":
		return JobGroupB, nil
	}
	return JobGroupUnknown, fmt.Errorf("Unknown job group")
}

// GetWageForJobGroup determines the hourly wage rate for given job group
func getWageForJobGroup(j JobGroup) int {
	switch j {
	case JobGroupA:
		return 20
	case JobGroupB:
		return 30
	}
	return 0
}

// TimeReport holds parsed details of a time report
type TimeReport struct {
	ID    string
	Items []*TimeReportItem
}

// TimeReportItem holds parsed details of a single line in time report
type TimeReportItem struct {
	Date  time.Time
	Hours time.Duration
	EmpID int
	Group JobGroup
}

// ServerHTTP serves as http handler for TimeReport ingestion request
func (h *TimeReport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h = new(TimeReport)
	// Only POST requests allowed
	if r.Method != http.MethodPost {
		writeErrResponse(w, http.StatusMethodNotAllowed, nil, "Invalid method: "+r.Method)
		return
	}
	log.Println("Processing TimeReportIngest request")
	// Limit max input file length
	// Ref: https://stackoverflow.com/a/40699578/6632880
	r.ParseMultipartForm(10000000) // ~ 10 MB

	// Extract and validate the file from form data
	file, header, err := r.FormFile("timereport")
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, err, "Error reading timereport file")
		return
	}
	defer file.Close()

	// Extract and validate the time-report ID from filename
	if h.ID, err = parseTimeReportID(header.Filename); err != nil {
		writeErrResponse(w, http.StatusBadRequest, err, "Time Report file format is incorrect")
		return
	}
	log.Println("Processing time-report ID: ", h.ID)
	if isDuplicate, err := InsertTimeReport(h.ID); err != nil {
		if isDuplicate {
			writeErrResponse(w, http.StatusBadRequest, err, "Time Report is already processed")
		} else {
			writeErrResponse(w, http.StatusInternalServerError, err, "DB error while processing time report")
		}
		return
	}

	// Read the csv file
	csvFile := csv.NewReader(file)
	// Ignore the header line
	csvFile.Read()
	// parse the contents
	for {
		var line []string
		if line, err = csvFile.Read(); err != nil {
			if err == io.EOF {
				err = nil // Reset error
			}
			break
		}
		var item *TimeReportItem
		if item, err = parseTimeReportItem(line); err != nil {
			break
		}
		log.Printf("Time-Report ID: %s | Read line: %v\n", h.ID, item)
		InsertTimeReportItem(h.ID, item)
		h.Items = append(h.Items, item)
	}
	// If there was an error reading the file, handle it now
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, err, "Failed processing time report file")
		return
	}
	// Form the HTTP Response
	fmt.Fprintf(w, "Time Report ID [%s] processed successfully", h.ID)
}

func parseTimeReportID(filename string) (string, error) {
	var timeReportID string
	matches := timeReportFileRegex.FindStringSubmatch(filename)
	if matches == nil || len(matches) != 2 {
		return timeReportID, fmt.Errorf("Filename format incorrect for time-report")
	}
	return matches[1], nil
}

func parseTimeReportItem(line []string) (*TimeReportItem, error) {
	if line == nil || len(line) != 4 {
		return nil, fmt.Errorf("Inconsistent columns")
	}
	item := new(TimeReportItem)
	var err error
	if item.Date, err = time.Parse(timeReportDateFormat, line[0]); err != nil {
		return nil, fmt.Errorf("Date parse error: %v", err)
	}
	if item.Hours, err = time.ParseDuration(line[1] + "h"); err != nil {
		return nil, fmt.Errorf("Hours parse error: %v", err)
	}
	if item.EmpID, err = strconv.Atoi(line[2]); err != nil {
		return nil, err
	}
	if item.Group, err = getJobGroup(line[3]); err != nil {
		return nil, err
	}
	return item, nil
}
