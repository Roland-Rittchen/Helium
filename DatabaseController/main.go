package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	// Import viper
	"github.com/spf13/viper"
	// Import Postgres
	_ "github.com/lib/pq"
)

var (
	host     = getEnvVariable("HOST")
	port     = getEnvVariable("PORT")
	user     = getEnvVariable("POSTGRES_USER")
	password = getEnvVariable("POSTGRES_PASSWORD")
	dbname   = getEnvVariable("DBNAME")
)

func main() {
	http.HandleFunc("/", GETHandler)
	http.HandleFunc("/insert", POSTHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Person struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

func getEnvVariable(key string) string {
	viper.SetConfigFile("../.env")
	// load .env file
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	value, ok := viper.Get(key).(string)

	if !ok {
		log.Fatalf("Invalid type assertion")
	}

	return value
}

func OpenConnection() *sql.DB {
	portNumber, err := strconv.Atoi(port)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, portNumber, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func GETHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	rows, err := db.Query("SELECT * FROM person")
	if err != nil {
		log.Fatal(err)
	}

	var people []Person

	for rows.Next() {
		var person Person
		rows.Scan(&person.Name, &person.Nickname)
		people = append(people, person)
	}

	peopleBytes, _ := json.MarshalIndent(people, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(peopleBytes)

	defer rows.Close()
	defer db.Close()
}

func POSTHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	var p Person
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO person (name, nickname) VALUES ($1, $2)`
	_, err = db.Exec(sqlStatement, p.Name, p.Nickname)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	defer db.Close()
}
