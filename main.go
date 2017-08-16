package main

import (
	"database/sql"
	"fmt"

	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host        = "postgres"
	port        = 5432
	user        = "postgres"
	password    = "password"
	dbname      = "postgres"
	createtable = `CREATE TABLE users (
  id serial NOT NULL,
  first_name TEXT,
  last_name TEXT,
  age INT,
  email TEXT UNIQUE NOT NULL,
  avatar TEXT,
  CONSTRAINT userinfo_pkey PRIMARY KEY (id)
)`
)

type User struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Age     int    `json:"age"`
	Email   string `json:"email"`
	Avatar  string `json:"avatar"`
}

var dbGlobal *sql.DB

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/users", handleUsers).Methods("GET")
	router.HandleFunc("/user", handleUser).Methods("POST")
	router.HandleFunc("/delete", handleDelete).Methods("POST")
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

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

	http.ListenAndServe(":8080", handler)

}

func handleDelete(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "POST":
		log.Println("Delete received ")
		user := new(User)
		decoder := json.NewDecoder(req.Body)
		error := decoder.Decode(&user)
		if error != nil {
			log.Println(error.Error())
			http.Error(res, error.Error(), http.StatusInternalServerError)
			return
		}
		log.Println(user)
		sqlStatement := `  
DELETE FROM users  
WHERE id = $1;`
		_, err := dbGlobal.Exec(sqlStatement, user.Id)
		if err != nil {
			log.Println(err.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return

		}

		res.WriteHeader(http.StatusOK)
	}

}

func handleUser(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "POST":
		log.Println("POST received %s", req.Body)
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
		err := dbGlobal.QueryRow("INSERT INTO users(first_name,last_name,age,email,avatar) VALUES($1,$2,$3,$4,$5) returning id;", user.Name, user.Surname, user.Age, user.Email, user.Avatar).Scan(&lastInsertId)

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

	var users []User

	for rows.Next() {
		var usr User

		err = rows.Scan(&usr.Id, &usr.Name, &usr.Surname, &usr.Age, &usr.Email, &usr.Avatar)
		if err != nil {
			fmt.Println("error:", err)
		}

		users = append(users, usr)
	}

	if len(users) > 0 {
		b, err := json.Marshal(users)
		if err != nil {
			fmt.Println("error:", err)
		}

		fmt.Fprintf(res, string(b))
	}

}
