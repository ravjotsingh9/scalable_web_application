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

	// Read environment variables
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare conn string and connect to DB
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, cfg.DBAddress, dbname)
	_, err = a.DB.DBConnection(connectionString)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Initialize API router
	a.Router = mux.NewRouter()
	a.initializeRoutes()

	// Initialize connection to Cache
	a.RedisClient.NewClient(cfg.REDISAddress)

	// Initialize report ID set
	a.Set = make(map[string]bool)

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
	if ret == false {
		var res response
		res.Status = http.StatusBadRequest
		res.Message = err.Error()
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return

	} else {
		var res response
		res.Status = http.StatusCreated
		res.Message = "Successfully uploaded"
		util.RespondWithJSON(w, http.StatusCreated, res)
		return
	}
}

const (
	ReportStr = "REPORT ID"
	Date      = "DATE"
)

func uploadToDB(filePath string, a *App) (bool, error) {

	ReportID, tmpFile, err := cleanAndReWriteCSV(filePath)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// check if it is recently processed
	if a.Set[ReportID] {
		err := errors.New("Already processed report, numbered " + ReportID)
		log.Println(err)
		return false, err
	}

	ret, err := a.DB.IfReportIDExists(ReportID)
	if err != nil {
		err := errors.New("Already processed report, numbered " + ReportID)
		log.Println(err)
		return false, err
	}
	if ret {
		err := errors.New("Already processed report, numbered " + ReportID)
		log.Println(err)
		return false, err
	}

	ret, err = a.DB.UploadCsv(tmpFile, ReportID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// Turn the cache to invalid
	a.RedisClient.SetToCache("valid", "false")

	// add to the mem
	a.Set[ReportID] = true

	return true, nil
}

func cleanAndReWriteCSV(filePath string) (string, string, error) {

	var ReportID string
	tmpFile := "tmp_" + filePath

	// open the given file
	file, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return ReportID, tmpFile, err
	}
	defer file.Close()

	//reader for given file
	scanner := bufio.NewScanner(file)

	//create a tmp file
	newfile, err := os.Create(tmpFile)
	if err != nil {
		log.Println(err)
		return ReportID, tmpFile, err
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
		return ReportID, tmpFile, err
	}
	return ReportID, tmpFile, nil
}
