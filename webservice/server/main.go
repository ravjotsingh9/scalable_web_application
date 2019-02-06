package main

// for database connection
const (
	USER   = "root"
	PASS   = "root"
	DBNAME = "payroll_report"
)

func main() {
	a := App{}
	a.Initialize(USER, PASS, DBNAME)

	a.Run(":8080")
}
