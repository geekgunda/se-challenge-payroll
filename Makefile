all: setup build run

setup: 
		docker-compose up -d
		sleep 10 # wait for mysql to start
		mysql -uroot -pbrutepass -h 127.0.0.1 < ${CURDIR}/db.sql

build:
		cd ${CURDIR}/cmd/payroll; go get -d; go clean -r; go build;

run:
		cd ${CURDIR}/cmd/payroll/; ./payroll

clean:
		docker-compose down

test:
		mysql -uroot -pbrutepass -h 127.0.0.1 < ${CURDIR}/db.sql
		go test -v
