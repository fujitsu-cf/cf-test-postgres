package main

import (
	"database/sql"
	"fmt"

	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 32757
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

type User struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Age     string `json:"age"`
	Email   string `json:"email"`
}

var dbGlobal *sql.DB

func main() {

	createtable := `CREATE TABLE users (
  id serial NOT NULL,
  first_name TEXT,
  last_name TEXT,
  age INT,
  email TEXT UNIQUE NOT NULL,
  CONSTRAINT userinfo_pkey PRIMARY KEY (id)
)`

	router := mux.NewRouter()
	router.HandleFunc("/users", handleUsers).Methods("GET")
	router.HandleFunc("/user", handleUser).Methods("POST")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	dbGlobal = db
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	stmt, err := db.Prepare(createtable)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}

	http.ListenAndServe(":9999", router)

}

func handleUser(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "POST":
		log.Println("POST received")
		user := new(User)
		decoder := json.NewDecoder(req.Body)
		error := decoder.Decode(&user)
		if error != nil {
			log.Println(error.Error())
			http.Error(res, error.Error(), http.StatusInternalServerError)
			return
		}
		log.Println(user)
		log.Println(dbGlobal)
		var lastInsertId int
		err := dbGlobal.QueryRow("INSERT INTO users(first_name,last_name,age,email) VALUES($1,$2,$3,$4) returning id;", user.Name, user.Surname, user.Age, user.Email).Scan(&lastInsertId)

		if err != nil {
			log.Println(err.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("last inserted id =", lastInsertId)

		res.WriteHeader(http.StatusCreated)
	}
}

func handleUsers(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	rows, err := dbGlobal.Query("SELECT * FROM users")
	if err != nil {
		log.Println(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	// Make a slice for the values
	values := make([]interface{}, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		// Print data
		for i, value := range values {
			switch value.(type) {
			case nil:
				fmt.Println(columns[i], ": NULL")

			case []byte:
				fmt.Println(columns[i], ": ", string(value.([]byte)))
				fmt.Fprintf(res, " %s ", string(value.([]byte)))

			default:
				fmt.Println(columns[i], ": ", value)
				fmt.Fprintf(res, " %v ", value)
			}
		}
		fmt.Fprintf(res, "\n ")
		fmt.Println("-----------------------------------")
	}

}
