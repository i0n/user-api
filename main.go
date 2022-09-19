package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	version "github.com/i0n/user-api/pkg/version"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	log "github.com/sirupsen/logrus"
)

var db *pgxpool.Pool

//Return a paginated list of Users, allowing for filtering by certain criteria (e.g. all Users with the country "UK")
//
//The service must:
//Be well documented

// User The main user struct for storing users in the db. Timestamps are dealt with by Postgres (See db/schema.sql)
type User struct {
	ID        int
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string
	Country   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserJSON for representing users as JSON
type UserJSON struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Country   string `json:"country"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MarshalJSON Override for User in order to format time to RFC3339
//
// Alternative Time formats:
//
// ANSIC	“Mon Jan _2 15:04:05 2006”
// UnixDate	“Mon Jan _2 15:04:05 MST 2006”
// RubyDate	“Mon Jan 02 15:04:05 -0700 2006”
// RFC822	“02 Jan 06 15:04 MST”
// RFC822Z	“02 Jan 06 15:04 -0700”
// RFC850	“Monday, 02-Jan-06 15:04:05 MST”
// RFC1123	“Mon, 02 Jan 2006 15:04:05 MST”
// RFC1123Z	“Mon, 02 Jan 2006 15:04:05 -0700”
// RFC3339	“2006-01-02T15:04:05Z07:00”
// RFC3339Nano	“2006-01-02T15:04:05.999999999Z07:00”
func (u *User) MarshalJSON() ([]byte, error) {
	res := UserJSON{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Password:  u.Password,
		Email:     u.Email,
		Country:   u.Country,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
	return json.Marshal(&res)
}

// HealthCheckHandler returns status 200
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":         true,
		"Version":    version.GetVersion(),
		"Revision":   version.GetRevision(),
		"Branch":     version.GetBranch(),
		"Built By":   version.GetBuildUser(),
		"Build Date": version.GetBuildDate(),
		"Go Version": version.GetGoVersion(),
	})
}

// CreateUserHandler POST a new user. Requires a unique email address. Key value pairs passed using x-www-form-urlencoded.
//
// Available fields:
//
// first_name
// last_name
// nickname
// password
// email
// country
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	if r.FormValue("email") == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "email is required."})
		return
	}
	_, err := db.Exec(context.Background(), "insert into users(first_name, last_name, nickname, password, email, country) values($1,$2,$3,$4,$5,$6)", r.FormValue("first_name"), r.FormValue("last_name"), r.FormValue("nickname"), r.FormValue("password"), r.FormValue("email"), r.FormValue("country"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": fmt.Sprintf("%v", err)})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	UsersChanged("created")
}

// GetUsersHandler GET a list of all users. Filter by query params e.g. /users?country=USA&first_name=Hulk would return all users from the USA with the first name Hulk.
//
// Available query params:
//
// first_name
// last_name
// nickname
// password
// email
// country
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {

	var sql string
	var sqlStrings []string
	var sqlCommand string

	var user User
	var users []User

	r.ParseForm()

	w.Header().Set("Content-Type", "application/json")

	filterKeys := []string{
		"first_name", "last_name", "nickname", "password", "email", "country",
	}
	for _, f := range filterKeys {
		if r.FormValue(f) != "" {
			sql = f + "= " + "'" + r.FormValue(f) + "'"
			sqlStrings = append(sqlStrings, sql)
		}
	}

	if len(sqlStrings) == 0 {
		sqlCommand = "select * from users"
	} else {
		sqlCommand = "select * from users WHERE " + strings.Join(sqlStrings, " AND ")
	}
	rows, _ := db.Query(context.Background(), sqlCommand)

	for rows.Next() {
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Nickname, &user.Password, &user.Email, &user.Country, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": fmt.Sprintf("%v", err)})
			return
		}
		users = append(users, user)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// PatchUserHandler PATCH an existing user. key value pairs passed using x-www-form-urlencoded.
func PatchUserHandler(w http.ResponseWriter, r *http.Request) {
	var sqlStrings []string
	var sql string

	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	vars := mux.Vars(r)
	updateableKeys := []string{
		"first_name", "last_name", "nickname", "password", "email", "country",
	}
	for _, f := range updateableKeys {
		if r.FormValue(f) != "" {
			sql = f + "= " + "'" + r.FormValue(f) + "'"
			sqlStrings = append(sqlStrings, sql)
		}
	}
	if len(sqlStrings) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": fmt.Sprintf("You did not provide any fields to update")})
		return
	}
	_, err := db.Exec(context.Background(), "UPDATE users SET "+strings.Join(sqlStrings, ",")+" WHERE id=$1;", vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": fmt.Sprintf("%v", err)})
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	UsersChanged("updated")
}

// DeleteUserHandler DELETE an existing user.
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	_, err := db.Exec(context.Background(), "delete from users where id=$1", vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": fmt.Sprintf("%v", err)})
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	UsersChanged("deleted")
}

// UsersChanged Intended for hook implementation based on user changes.
func UsersChanged(s string) {
	log.Println("UsersChanged. user " + s)
	// TODO Hooks...
}

func main() {
	// Set vars
	var err error

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresURL := os.Getenv("POSTGRES_URL")
	postgresDB := os.Getenv("POSTGRES_DB")

	databaseURL := "postgres://" + postgresUser + ":" + postgresPassword + "@" + postgresURL + "/" + postgresDB

	wait := time.Second * 15

	// Setup DB

	db, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Setup http server

	r := mux.NewRouter()
	r.HandleFunc("/", HealthCheckHandler)
	r.HandleFunc("/user", CreateUserHandler).Methods("POST")
	r.HandleFunc("/users", GetUsersHandler).Methods("GET")
	r.HandleFunc("/user/{id}", PatchUserHandler).Methods("PATCH")
	r.HandleFunc("/user/{id}", DeleteUserHandler).Methods("DELETE")

	log.Println("################################################")
	log.Println()
	log.Println("User API server starting on port 8080")
	log.Println()
	log.Println("Version: " + version.GetVersion())
	log.Println("Revision: " + version.GetRevision())
	log.Println("Branch: " + version.GetBranch())
	log.Println("Built By: " + version.GetBuildUser())
	log.Println("Build Date: " + version.GetBuildDate())
	log.Println("Go Version: " + version.GetGoVersion())
	log.Println("Graceful shutdown period: " + wait.String())
	log.Println()
	log.Println("################################################")

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via:
	// SIGINT (Ctrl+C)
	// SIGQUIT or SIGTERM (Ctrl+/)
	//
	// SIGKILL isn't caught so will immediately shutdown the server
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGQUIT)
	signal.Notify(c, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
