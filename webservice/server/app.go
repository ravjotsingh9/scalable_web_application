// app.go

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cache_redis "github.com/ravjotsingh9/payroll-application/webservice/server/redis"
	util "github.com/ravjotsingh9/payroll-application/webservice/server/util"

	db "github.com/ravjotsingh9/payroll-application/webservice/server/mysql"

	"github.com/gorilla/mux"
	_ "github.com/jinzhu/now"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBAddress    string `envconfig:"DB_ADDRESS"`
	REDISAddress string `envconfig:"REDIS_ADDRESS"`
}

type App struct {
	Router *mux.Router
	//DB          *sql.DB
	DB          db.Mysqldb
	Set         map[string]bool
	RedisClient cache_redis.RedisClient
}
type Record struct {
	Date        string  `json:"date"`
	HoursWorked float64 `json:"hoursWorked"`
	EmployeeID  int     `json:"employeeID"`
	JobGroup    string  `json:"jobGroup"`
	PayDay      string  `json:"payDate"`
	Salary      float64 `json:"salary,omitempty"`
}

type response struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

func (a *App) Initialize(user, password, dbname string) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, cfg.DBAddress, dbname)
	/*
		a.DB, err = sql.Open("mysql", connectionString)
		if err != nil {
			log.Fatal("Couldn't connect to DB")
		}
	*/

	a.DB.DBConnection(connectionString)

	a.Router = mux.NewRouter()
	a.initializeRoutes()

	// set to maintain report IDs
	a.Set = make(map[string]bool)

	a.RedisClient.NewClient(cfg.REDISAddress)
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/uploadReport", a.uploadReport)
	a.Router.HandleFunc("/getReport", a.getReport)
}

func (a *App) getReport(w http.ResponseWriter, r *http.Request) {

	val, _ := a.RedisClient.GetFromCache("valid")

	if strings.Compare(val, "true") == 0 {
		content, _ := a.RedisClient.GetFromCache("content")
		util.RespondWithJSONFromBytes(w, http.StatusCreated, []byte(content))
		return
	} else {
		/*
			statement := fmt.Sprintf("(select employee_id, payday,  sum(record.pay) as salary from (select date, hours_worked, employee_id, job_group, payday, hours_worked*(select sal from job_group as salTable where salTable.job_group = rec.job_group) as pay from  record as rec) as record  group by employee_id, payday) ")
			rows, err := a.DB.Query(statement)
			if err != nil {
				util.RespondWithError(w, http.StatusInternalServerError, "Failed to query DB")
				return
			}

			defer rows.Close()

			records := []Record{}

			for rows.Next() {
				var record Record
				if err := rows.Scan(&record.EmployeeID, &record.PayDay, &record.Salary); err != nil {
					util.RespondWithError(w, http.StatusInternalServerError, "No rows returned from DB")
					return
				}
				records = append(records, record)
			}
		*/
		records, err := a.DB.GetReport()
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, "Failed to query DB")
			return
		}

		recordsJSONBytes, _ := json.Marshal(records)
		recordsJSON := string(recordsJSONBytes)

		a.RedisClient.SetToCache("content", recordsJSON)

		a.RedisClient.SetToCache("valid", "true")

		util.RespondWithJSON(w, http.StatusCreated, records)
		return
	}

}

func (a *App) uploadReport(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method)

	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("filepond")

	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	ret, err := uploadToDB(handler.Filename, a)
	fmt.Println(ret)
	if ret == false {
		var res response
		res.Status = http.StatusBadRequest
		res.Message = err.Error()
		util.RespondWithJSON(w, http.StatusBadRequest, res)
		return

	} else {
		var res response
		res.Status = http.StatusCreated
		res.Message = "Successfully uploaded"
		util.RespondWithJSON(w, http.StatusCreated, res)
		return
	}
}

/*
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	//fmt.Println(payload)
	response, _ := json.Marshal(payload)
	//fmt.Println(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithJSONFromBytes(w http.ResponseWriter, code int, dataInbytes []byte) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dataInbytes)
}
*/
const (
	ReportStr = "REPORT ID"
	Date      = "DATE"
)

func uploadToDB(filePath string, a *App) (bool, error) {

	var ReportID string
	tmpFile := "tmp_" + filePath
	//// copying a file line by line

	// open the given file
	file, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer file.Close()

	//reader for given file
	scanner := bufio.NewScanner(file)

	//create a tmp file
	newfile, err := os.Create(tmpFile)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer newfile.Close()

	// writer for new file
	writer := bufio.NewWriter(newfile)

	// iterate over the file and write new clean (last line removed) file
	for scanner.Scan() {
		if strings.Contains(strings.ToUpper(scanner.Text()), Date) {
			continue
		}
		if strings.Contains(strings.ToUpper(scanner.Text()), ReportStr) {
			row := scanner.Text()
			cols := strings.Split(row, ",")
			ReportID = cols[1]
			break
		} else {
			var payday string
			//TODO: calculate the payday and append it at the end
			line := scanner.Text()
			words := strings.Split(line, ",")
			splittedStr := strings.Split(words[0], "/")

			year, err := strconv.Atoi(splittedStr[2])
			if err != nil {
				log.Println(err)
			}
			month, err := strconv.Atoi(splittedStr[1])
			if err != nil {
				log.Println(err)
			}
			day, err := strconv.Atoi(splittedStr[0])
			if err != nil {
				log.Println(err)
			}
			if day <= 15 && day >= 1 {
				payday = splittedStr[2] + "-" + splittedStr[1] + "-15"
			} else {
				date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
				lastOfMonth := date.AddDate(0, 1, -1)
				layout := "2006-01-02"
				payday = lastOfMonth.Format(layout)
			}

			_, err = writer.WriteString(scanner.Text() + "," + payday + "\n")
			writer.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
		return false, err
	}

	//// check if report id already exist

	// check if it is recently processed
	if a.Set[ReportID] {
		err := errors.New("Already processed report, numbered " + ReportID)
		log.Println(err)
		return false, err
	}

	// check if is in db
	/*
		reportID, err := strconv.Atoi(ReportID)
		if err != nil {
			log.Println(err)
		}
		statement := fmt.Sprintf("SELECT report_id FROM record WHERE report_id=%d", reportID)
		rows, err := a.DB.Query(statement)
		if err != nil {
			log.Println(err)
			return false, err
		}

		for rows.Next() {
			err := errors.New("Already processed report, numbered " + ReportID)
			log.Println(err)
			return false, err
		}
	*/

	ret, err := a.DB.IfReportIDExists(ReportID)
	if err != nil {
		log.Println(err)
		return false, err
	}
	if ret {
		err := errors.New("Already processed report, numbered " + ReportID)
		log.Println(err)
		return false, err
	}

	/*
		//register the file
		mysql.RegisterLocalFile(tmpFile)

		//load the file into database
		//execCmd := "LOAD DATA LOCAL INFILE '" + tmpFile + "' INTO TABLE RECORD fields terminated BY ',' lines terminated BY '\n' IGNORE 1 LINES (@vdate, hours_worked, employee_id, job_group, @vpayday, report_id)  SET date = STR_TO_DATE(@vdate,'%d/%m/%Y'), report_id =" + ReportID + ", payday = STR_TO_DATE(@vpayday,'%Y-%m-%d') ;"
		execCmd := "LOAD DATA LOCAL INFILE '" + tmpFile + "' INTO TABLE record fields terminated BY ',' lines terminated BY '\n' IGNORE 1 LINES (@vdate, hours_worked, employee_id, job_group, @vpayday, report_id)  SET date = STR_TO_DATE(@vdate,'%d/%m/%Y'), report_id =" + ReportID + ", payday = STR_TO_DATE(@vpayday,'%Y-%m-%d') ;"

		_, err = a.DB.Exec(execCmd)
		if err != nil {
			log.Println(err)
			return false, err
		}
	*/

	ret, err = a.DB.UploadCsv(tmpFile, ReportID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	a.RedisClient.SetToCache("valid", "false")

	// add to the mem
	a.Set[ReportID] = true
	return true, nil
}

// create table RECORD (date Date, hours_worked FLOAT, employee_id VARCHAR(20), job_group CHAR(1), report_id INTEGER, payday Date);
// create table job_group (job_group CHAR(1), sal FLOAT);
//insert into job_group values('A',20);
//insert into job_group values('B',30);
