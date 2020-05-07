package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type configuration struct {
	ServerPort int    `json:"server_port"`
	Username   string `json:"mysql_username"`
	Password   string `json:"mysql_password"`
	Server     string `json:"mysql_server"`
	Database   string `json:"mysql_database"`
	Port       int    `json:"mysql_port"`
}

type tableinfo struct {
	Table  string   `json:"Table"`
	Fields []string `json:"Fields"`
	Alias  []string `json:"Alias"`
}

var tables map[string]json.RawMessage
var serverstring string
var port string
var verbose bool

const version = "1.0.0"
const author = "Michel Noel"

func init() {
	// Parse flags
	verbosePtr := flag.Bool("verbose", false, "Show more information on the webserver")
	helpPtr := flag.Bool("help", false, "Show help")
	versionPtr := flag.Bool("version", false, "Show version")
	flag.Parse()

	// Show help
	if *helpPtr {

		help := ` 
jsonToMySQL - This is a simple webserver that takes a preformatted http get request and parses the JSON and puts it into a preconfigured MySQL database.

Arguments:
`
		fmt.Println(help)
		flag.PrintDefaults()
		fmt.Println(" ")
		os.Exit(0)
	}

	if *versionPtr {
		fmt.Println(version, "- Written by", author)
		os.Exit(0)
	}

	fmt.Println("jsonToMySQL Server Running")
	verbose = *verbosePtr

	// Get script working directory
	_, filename, _, _ := runtime.Caller(0)

	// load configruation file
	mysqlfullpath := path.Join(path.Dir(filename), "./config.json")
	mysqlfile, err := ioutil.ReadFile(mysqlfullpath)
	errorcheck(err)

	data := configuration{}
	err = json.Unmarshal([]byte(mysqlfile), &data)
	errorcheck(err)

	serverstring = data.Username + ":" + data.Password + "@tcp(" + data.Server + ":" + strconv.Itoa(data.Port) + ")/" + data.Database
	port = strconv.Itoa(data.ServerPort)

	// load tables.json file
	tablesfullpath := path.Join(path.Dir(filename), "./tables.json")
	tablesdata, err := ioutil.ReadFile(tablesfullpath)
	errorcheck(err)

	err = json.Unmarshal([]byte(tablesdata), &tables)
	errorcheck(err)
}

func main() {
	http.HandleFunc("/", handleServer)
	http.ListenAndServe(":"+port, nil)
}

func handleServer(w http.ResponseWriter, r *http.Request) {
	result := "Failed"
	auth := r.Header.Get("Authorization")
	body, err := ioutil.ReadAll(r.Body)
	errorcheck(err)

	// Return result to end user (always assume failed)
	defer func() {
		if err := recover(); err != nil && verbose {
			fmt.Println(err)
		}
		fmt.Fprintf(w, result)
	}()

	if auth != "" && len(body) > 0 {
		if table, fields, values, err := passBody(auth, body); err != nil {
			errorcheck(err)
		} else {
			if err := passToMySQL(table, fields, values); err != nil {
				errorcheck(err)
			} else {
				result = "Ok"
			}
		}
	}
}

func passBody(auth string, body []byte) (table string, fields []string, values string, _ error) {
	// Get relevant table
	tableGUID := tables[auth]
	if len(tableGUID) == 0 {
		return "", nil, "", errors.New("no such table")
	}

	tableinfo := tableinfo{}
	if err := json.Unmarshal(tableGUID, &tableinfo); err != nil {
		return "", nil, "", errors.New("table.json issue")
	}

	// Get table name / alias / fields
	table = tableinfo.Table
	alias := tableinfo.Alias
	fields = tableinfo.Fields

	if len(table) == 0 || len(alias) == 0 || len(fields) == 0 {
		return "", nil, "", errors.New("issue with table values")
	}

	var objmap map[string]interface{}
	if err := json.Unmarshal([]byte(body), &objmap); err != nil {
		return "", nil, "", errors.New("load body issue")
	}

	// Pull data from body and put into comma delimited values
	for i := 0; i < len(alias); i++ {
		comma := ","
		if i+1 == len(alias) {
			comma = ""
		}
		tempval := objmap[string(alias[i])]

		var str string
		switch tempval.(type) {
		case int:
			str = strconv.Itoa(tempval.(int))
		case float64:
			str = fmt.Sprintf("%f", tempval)
		case string:
			str = "\"" + tempval.(string) + "\""
		default:
			str = ""
		}
		values = values + str + comma
	}
	return table, fields, values, nil
}

func passToMySQL(table string, fields []string, values string) (err error) {

	db, err := sql.Open("mysql", serverstring)
	errorcheck(err)

	// close database after all work is done
	defer db.Close()
	pingDB(db)

	// prepare statement
	stmt, err := db.Prepare("Replace into " + table + " (" + strings.Join(fields, ",") + ") values (" + values + " )")
	errorcheck(err)

	//execute
	res, err := stmt.Exec()
	errorcheck(err)

	id, err := res.LastInsertId()
	errorcheck(err)

	if verbose {
		fmt.Println("Insert id", id)
	}
	return err
}

func errorcheck(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func pingDB(db *sql.DB) {
	err := db.Ping()
	errorcheck(err)
}
