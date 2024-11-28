package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Configurations
const (
	dbUser     = "go_user" // Change to your MySQL username
	dbPassword = "2995"    // Change to your MySQL password
	dbName     = "toronto_time"
	dbHost     = "localhost"
	dbPort     = "3306"
)

// Global database connection
var db *sql.DB

// TimeResponse structure for JSON response
type TimeResponse struct {
	CurrentTime string `json:"current_time"`
}

func main() {
	var err error

	// Initialize the database connection
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName))
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Set up HTTP handlers
	http.HandleFunc("/current-time", currentTimeHandler)
	http.HandleFunc("/all-times", allTimesHandler)

	log.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// Handler for /current-time
func currentTimeHandler(w http.ResponseWriter, r *http.Request) {
	loc, err := time.LoadLocation("America/Toronto")
	if err != nil {
		http.Error(w, "Failed to load timezone", http.StatusInternalServerError)
		log.Printf("Timezone error: %v", err)
		return
	}

	// Get current time in Toronto
	torontoTime := time.Now().In(loc)

	// Insert into database
	_, err = db.Exec("INSERT INTO time_log (timestamp) VALUES (?)", torontoTime)
	if err != nil {
		http.Error(w, "Database insert error", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	// Send JSON response
	response := TimeResponse{CurrentTime: torontoTime.Format(time.RFC3339)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler for /all-times
func allTimesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT timestamp FROM time_log")
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		log.Printf("Database query error: %v", err)
		return
	}
	defer rows.Close()

	var times []string
	for rows.Next() {
		var timestamp time.Time
		if err := rows.Scan(&timestamp); err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			log.Printf("Row scan error: %v", err)
			return
		}
		times = append(times, timestamp.Format(time.RFC3339))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(times)
}
