package mysqlDB

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	mysql "github.com/go-sql-driver/mysql"
)

const (
	ReportStr = "REPORT ID"
	Date      = "DATE"
)

type Idb interface {
	DBConnetion(connStr string) (*Mysqldb, error)
	GetReport() error
	UploadCsv(string) (bool, error)
	IfReportIDExists(string) (bool, error)
	Close()
}

type Mysqldb struct {
	DB *sql.DB
}

type Record struct {
	Date        string  `json:"date"`
	HoursWorked float64 `json:"hoursWorked"`
	EmployeeID  int     `json:"employeeID"`
	JobGroup    string  `json:"jobGroup"`
	PayDay      string  `json:"payDate"`
	Salary      float64 `json:"salary,omitempty"`
}

func (m *Mysqldb) DBConnection(connStr string) (*Mysqldb, error) {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}
	m.DB = db
	return &Mysqldb{
		m.DB,
	}, nil
}

func (m *Mysqldb) Close() {
	m.DB.Close()
}

func (m *Mysqldb) GetReport() ([]Record, error) {
	statement := fmt.Sprintf("(select employee_id, payday,  sum(record.pay) as salary from (select date, hours_worked, employee_id, job_group, payday, hours_worked*(select sal from job_group as salTable where salTable.job_group = rec.job_group) as pay from  record as rec) as record  group by employee_id, payday) ")
	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	records := []Record{}

	for rows.Next() {
		var record Record
		if err := rows.Scan(&record.EmployeeID, &record.PayDay, &record.Salary); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (m *Mysqldb) UploadCsv(cleanedCsvfilePath string, ReportID string) (bool, error) {
	//register the file
	mysql.RegisterLocalFile(cleanedCsvfilePath)

	//load the file into database
	//execCmd := "LOAD DATA LOCAL INFILE '" + tmpFile + "' INTO TABLE RECORD fields terminated BY ',' lines terminated BY '\n' IGNORE 1 LINES (@vdate, hours_worked, employee_id, job_group, @vpayday, report_id)  SET date = STR_TO_DATE(@vdate,'%d/%m/%Y'), report_id =" + ReportID + ", payday = STR_TO_DATE(@vpayday,'%Y-%m-%d') ;"
	execCmd := "LOAD DATA LOCAL INFILE '" + cleanedCsvfilePath + "' INTO TABLE record fields terminated BY ',' lines terminated BY '\n' IGNORE 1 LINES (@vdate, hours_worked, employee_id, job_group, @vpayday, report_id)  SET date = STR_TO_DATE(@vdate,'%d/%m/%Y'), report_id =" + ReportID + ", payday = STR_TO_DATE(@vpayday,'%Y-%m-%d') ;"

	_, err := m.DB.Exec(execCmd)
	if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

func (m *Mysqldb) IfReportIDExists(ReportID string) (bool, error) {
	reportID, err := strconv.Atoi(ReportID)
	if err != nil {
		log.Println(err)
		return false, err
	}
	statement := fmt.Sprintf("SELECT report_id FROM record WHERE report_id=%d", reportID)
	rows, err := m.DB.Query(statement)
	if err != nil {
		log.Println(err)
		return false, err
	}

	for rows.Next() {
		err := errors.New("Already processed report, numbered " + ReportID)
		log.Println(err)
		return true, err
	}

	return false, nil
}
