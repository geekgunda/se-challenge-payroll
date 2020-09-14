DROP DATABASE IF EXISTS payroll;

CREATE DATABASE payroll;

USE payroll;

CREATE TABLE `timereport`(
    `report_id` varchar(64) NOT NULL PRIMARY KEY,
    `status` varchar(128) DEFAULT NULL
) ENGINE=InnoDB CHARACTER SET utf8;

CREATE TABLE `timereportitem`(
    `report_id` varchar(64) NOT NULL,
    `emp_id` int NOT NULL,
    `work_date` date NOT NULL,
    `work_hours` varchar(64) NOT NULL,
    `job_group` varchar(64) NOT NULL,
    UNIQUE(`work_date`, `emp_id`)
) ENGINE=InnoDB CHARACTER SET utf8;
