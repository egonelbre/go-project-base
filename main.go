package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

var (
	addr     = flag.String("listen", ":80", "http server `address`")
	database = flag.String("database", "user='unsafe' password='unsafe' dbname='unsafe' sslmode='disable' host='localhost' port='5432'", "database")

	assets    = flag.String("assets", "assets", "assets directory")
	templates = flag.String("templates", "templates", "templates directory")
)

func AmazonRDS() string {
	user := os.Getenv("RDS_USERNAME")
	pass := os.Getenv("RDS_PASSWORD")

	dbname := os.Getenv("RDS_DB_NAME")
	host := os.Getenv("RDS_HOSTNAME")
	port := os.Getenv("RDS_PORT")

	if user == "" || pass == "" || dbname == "" || host == "" || port == "" {
		return ""
	}

	return fmt.Sprintf("user='%s' password='%s' dbname='%s' host='%s' port='%s'", user, pass, dbname, host, port)
}

func initFlags() {
	host, port := os.Getenv("HOST"), os.Getenv("PORT")
	if host != "" || port != "" {
		*addr = host + ":" + port
	}
	if os.Getenv("DATABASE") != "" {
		*database = os.Getenv("DATABASE")
	}
	if rds := AmazonRDS(); rds != "" {
		*database = rds
	}
}

var T *template.Template
var DB *sql.DB

func main() {
	flag.Parse()
	initFlags()

	T = template.Must(template.ParseGlob(filepath.Join(*templates, "**")))
	SetupDB()

	log.Printf("Starting on %s", *addr)

	http.Handle("/assets/",
		http.StripPrefix("/assets/", http.FileServer(http.Dir(*assets))))

	http.HandleFunc("/", index)
	http.ListenAndServe(*addr, nil)
}

func SetupDB() {
	log.Println("Using DB:", *database)

	db, err := sql.Open("postgres", *database)
	if err != nil {
		log.Fatal(err)
	}

	result, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS
		RequestLog (
			ID SERIAL,
			Message TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", result)

	DB = db
}

func index(rw http.ResponseWriter, req *http.Request) {
	var errors []error

	_, err := DB.Exec("INSERT INTO RequestLog (Message) VALUES ($1)", req.UserAgent())
	if err != nil {
		log.Println(err)
		errors = append(errors, err)
	}

	type LogEntry struct {
		ID      int
		Message string
	}

	errors = append(errors, fmt.Errorf("hello world"))

	entries := make([]LogEntry, 0, 20)
	rows, err := DB.Query("SELECT * FROM RequestLog ORDER BY Id DESC LIMIT 20")
	if err != nil {
		log.Println(err)
		errors = append(errors, err)
	}
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(&entry.ID, &entry.Message)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		errors = append(errors, err)
	}

	err = T.ExecuteTemplate(rw, "index.html", map[string]interface{}{
		"Time":    time.Now(),
		"Entries": entries,
		"Errors":  errors,
	})
}
