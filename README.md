# Wave Software Development Challenge

## Documentation:

### Instructions on how to build and run the app:

Dependencies: `docker`, `mysql-client-core-8.0`, `go`. On an Ubuntu 20.04 machine, you can install these via:
- `sudo snap install docker`
- `sudo apt install mysql-client-core-8.0`
- Instructions for downloading and installing Go: https://golang.org/doc/install
- setup `$GOPATH` env variable and add the Go binary to `$PATH` env variable

Preparation
- Create the necessary directory structure: `mkdir -p $GOPATH/src/github.com/geekgunda/`
- Extract the git bundle into the directory above

Installation (using Makefile):
- `make setup`: download and setup mysql-server 8.0 docker image (using docker-compose)
- `make build`: compile go executable binary for this app (using go binary)
- `make run`  : start the app
- `make clean`: stop docker containers and remove them

The server will be available at http://127.0.0.1:8081/ with following endpoints:
1. Time-Report Ingestion: `curl --request POST 'http://127.0.0.1:8081/timereport' --form 'timereport=@time-report-42.csv'`
1. Payroll-Report: `curl --request GET 'http://127.0.0.1:8081/payrollreport'`


##### Points to note:
- Ensure nothing is running on ports 3306 and 8081 on local machine
- Ensure the app base directory is: `$GOPATH/src/github.com/geekgunda/se-challenge-payroll`
- `make setup` might need to run with `sudo`, if docker is running as `root` user
- `make setup` might fail as docker container might not have started by then. Retry the command in that case

-----

### Design

Time Report Ingestion API: 
- The API extracts, parses and validates data, before archiving it in DB
- Strict Go types (`time.Time` for date, `time.Duration` for work hours) are used to simplify validation and computation of payroll reports

Payroll Report API:
- Since this report needs to be across all uploaded time reports, first all unique year-month pairs are extracted from DB
- Next we iterate over these pairs, splitting them into the two specified pay periods per month and fetching employee records
- CustomDate type is used for PayPeriod date fields, so that requested response formatting can be supported (`"YYYY-MM-DD"`)

Misc:
- `model.go` deals with database operations (fetching or storing data in DB)
- Standard library's `database/sql` package along with `go-sql-driver/mysql` driver has been used
- DB schema is documented in `db.sql` file
- `main.go` is kept in a separate package inside `cmd/` directory in accordance with [Go standard project layout](https://github.com/golang-standards/project-layout)
- Rest of the files are not broken down into separate packages to simplify the basic implementation

MySQL is used as the data store to keep time-report info, and also process payroll-report.
There are two tables used:
1. `timereport`: Metadata about each uploaded time-report (ID). It is also used as a locking mechanism to handle race conditions and duplicate updates.
1. `timereportitem`: Granular data extracted from uploaded time-report (employee ID, hours, job group and date) stored for reporting and archival purposes.

-----

### Questionnaire

#### How was the app tested:

- Manual tests using Postman and the input data shared in examples
- Automated end to end test using sample shared below

#### Improvements for making the app production ready:

- Move payroll report processing into an async job, and change the API to just serve pre-computed stats.
- Add mandatory query param for pay period to payroll report API. Full table scans in production are not recommended!
- Enhancing time-report ingestion API to take more details (like company or entity, time-report name etc)
- Instrumentation: Integrating with corresponding library to track HTTP response code, latency and other metrics
- Containerization: Converting entire app into a container for seamless build and deploy pipeline
- Profiling: Identifying any bottlenecks and resolving them (Ex: DB throughput)
- SPOF: Identifying and addressing single points of failure in the app (Ex: Add retries and circuit breaker for DB queries)
- Add authentication to these APIs (assuming they are publicly exposed)
- Using go modules for dependency management
- Connection pooling within DB library

#### Pending updates (compromises due to time constraints):

- More automated tests (both unit and end to end)
- Assumption: An employee will work in only one job group per pay cycle (Ex: EmpID=1 worked only within JobGroup=1 from 1 Sep to 15 Sep)
- Better error handling and reporting (Currently it is mostly restricted to setting correct HTTP Response Code in error cases)
- Using Status field in timereport table (Idea is to set it to 'processing' at the beginning, and 'processed' at the end. This way better errors can be shown during duplicate requests)
- Creating separate packages for handlers, model, contracts etc.
- Having a config file (for DB variables, log level, etc)
- Resetting DB state within automated end to end test (Currently it is throwing some errors, so it needs to be done manually before each test run)

-----
### Original Project README:

Applicants for the Full-stack Developer role at Wave must
complete the following challenge, and submit a solution prior to the onsite
interview.

The purpose of this exercise is to create something that we can work on
together during the onsite. We do this so that you get a chance to collaborate
with Wavers during the interview in a situation where you know something better
than us (it's your code, after all!)

There isn't a hard deadline for this exercise; take as long as you need to
complete it. However, in terms of total time spent actively working on the
challenge, we ask that you not spend more than a few hours, as we value your
time and are happy to leave things open to discussion in the on-site interview.

Please use whatever programming language and framework you feel the most
comfortable with.

Feel free to email [dev.careers@waveapps.com](dev.careers@waveapps.com) if you
have any questions.

## Project Description

Imagine that this is the early days of Wave's history, and that we are prototyping a new payroll system API. A front end (that hasn't been developed yet, but will likely be a single page application) is going to use our API to achieve two goals:

1. Upload a CSV file containing data on the number of hours worked per day per employee
1. Retrieve a report detailing how much each employee should be paid in each _pay period_

All employees are paid by the hour (there are no salaried employees.) Employees belong to one of two _job groups_ which determine their wages; job group A is paid $20/hr, and job group B is paid $30/hr. Each employee is identified by a string called an "employee id" that is globally unique in our system.

Hours are tracked per employee, per day in comma-separated value files (CSV).
Each individual CSV file is known as a "time report", and will contain:

1. A header, denoting the columns in the sheet (`date`, `hours worked`,
   `employee id`, `job group`)
1. 0 or more data rows

In addition, the file name should be of the format `time-report-x.csv`,
where `x` is the ID of the time report represented as an integer. For example, `time-report-42.csv` would represent a report with an ID of `42`.

You can assume that:

1. Columns will always be in that order.
1. There will always be data in each column and the number of hours worked will always be greater than 0.
1. There will always be a well-formed header line.
1. There will always be a well-formed file name.

A sample input file named `time-report-42.csv` is included in this repo.

### What your API must do:

We've agreed to build an API with the following endpoints to serve HTTP requests:

1. An endpoint for uploading a file.

   - This file will conform to the CSV specifications outlined in the previous section.
   - Upon upload, the timekeeping information within the file must be stored to a database for archival purposes.
   - If an attempt is made to upload a file with the same report ID as a previously uploaded file, this upload should fail with an error message indicating that this is not allowed.

1. An endpoint for retrieving a payroll report structured in the following way:

   _NOTE:_ It is not the responsibility of the API to return HTML, as we will delegate the visual layout and redering to the front end. The expectation is that this API will only return JSON data.

   - Return a JSON object `payrollReport`.
   - `payrollReport` will have a single field, `employeeReports`, containing a list of objects with fields `employeeId`, `payPeriod`, and `amountPaid`.
   - The `payPeriod` field is an object containing a date interval that is roughly biweekly. Each month has two pay periods; the _first half_ is from the 1st to the 15th inclusive, and the _second half_ is from the 16th to the end of the month, inclusive. `payPeriod` will have two fields to represent this interval: `startDate` and `endDate`.
   - Each employee should have a single object in `employeeReports` for each pay period that they have recorded hours worked. The `amountPaid` field should contain the sum of the hours worked in that pay period multiplied by the hourly rate for their job group.
   - If an employee was not paid in a specific pay period, there should not be an object in `employeeReports` for that employee + pay period combination.
   - The report should be sorted in some sensical order (e.g. sorted by employee id and then pay period start.)
   - The report should be based on all _of the data_ across _all of the uploaded time reports_, for all time.

   As an example, given the upload of a sample file with the following data:

    <table>
    <tr>
      <th>
        date
      </th>
      <th>
        hours worked
      </th>
      <th>
        employee id
      </th>
      <th>
        job group
      </th>
    </tr>
    <tr>
      <td>
        2020-01-04
      </td>
      <td>
        10
      </td>
      <td>
        1
      </td>
      <td>
        A
      </td>
    </tr>
    <tr>
      <td>
        2020-01-14
      </td>
      <td>
        5
      </td>
      <td>
        1
      </td>
      <td>
        A
      </td>
    </tr>
    <tr>
      <td>
        2020-01-20
      </td>
      <td>
        3
      </td>
      <td>
        2
      </td>
      <td>
        B
      </td>
    </tr>
    <tr>
      <td>
        2020-01-20
      </td>
      <td>
        4
      </td>
      <td>
        1
      </td>
      <td>
        A
      </td>
    </tr>
    </table>

   A request to the report endpoint should return the following JSON response:

   ```javascript
   {
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
   }
   ```

We consider ourselves to be language agnostic here at Wave, so feel free to use any combination of technologies you see fit to both meet the requirements and showcase your skills. We only ask that your submission:

- Is easy to set up
- Can run on either a Linux or Mac OS X developer machine
- Does not require any non open-source software

### Documentation:

Please commit the following to this `README.md`:

1. Instructions on how to build/run your application
1. Answers to the following questions:
   - How did you test that your implementation was correct?
   - If this application was destined for a production environment, what would you add or change?
   - What compromises did you have to make as a result of the time constraints of this challenge?

## Submission Instructions

1. Clone the repository.
1. Complete your project as described above within your local repository.
1. Ensure everything you want to commit is committed.
1. Create a git bundle: `git bundle create your_name.bundle --all`
1. Email the bundle file to [dev.careers@waveapps.com](dev.careers@waveapps.com) and CC the recruiter you have been in contact with.

## Evaluation

Evaluation of your submission will be based on the following criteria.

1. Did you follow the instructions for submission?
1. Did you complete the steps outlined in the _Documentation_ section?
1. Were models/entities and other components easily identifiable to the
   reviewer?
1. What design decisions did you make when designing your models/entities? Are
   they explained?
1. Did you separate any concerns in your application? Why or why not?
1. Does your solution use appropriate data types for the problem as described?
