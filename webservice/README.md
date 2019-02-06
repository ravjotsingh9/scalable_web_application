# payroll_app_API

Steps to run the server:
1. Install mysql
2. Start mysql and log into the mysql teminal.
3. Set up root user for mysql with password as "root".
4. Create two tables and insert some record with following commands:
    create database payroll_report;
    create table RECORD (date Date, hours_worked FLOAT, employee_id VARCHAR(20), job_group CHAR(1), report_id INTEGER, payday Date);
    create table job_group (job_group CHAR(1), sal FLOAT);
    insert into job_group values('A',20);
    insert into job_group values('B',30);
5. Set up Go environment variables. Make sure GOPATH and GOBIN are set correctly.
6. Compile code with `go build`.
7. Run the generated binary. The following two rest end points will be available:
    GET  /getReport
        - retreive the required report in json format
    POST /uploadReport
        - expects file and return sucess or error in json format
8. Once server is running run the frontend server.
