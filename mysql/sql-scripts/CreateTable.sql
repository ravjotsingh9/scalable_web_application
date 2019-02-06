CREATE DATABASE
IF NOT EXISTS payroll_report;
use payroll_report;
create table record
(
    date Date,
    hours_worked FLOAT,
    employee_id VARCHAR(20),
    job_group CHAR(1),
    report_id INTEGER,
    payday Date
);
create table job_group
(
    job_group CHAR(1),
    sal FLOAT
);

insert into job_group
values('A', 20);
insert into job_group
values('B', 30);