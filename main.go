package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

// RequestData структура для входящих данных
type RequestData struct {
	User   string `json:"user"`
	Action string `json:"action"`
}

// ResponseData структура для ответа
type ResponseData struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// Record структура для получения данных из БД
type Record struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

func initDB() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		return err
	}

	// Создаём таблицу если её нет
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS requests (
		id SERIAL PRIMARY KEY,
		"user" VARCHAR(255) NOT NULL,
		action VARCHAR(255) NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return err
	}

	fmt.Println("✅ Подключение к БД успешно, таблица готова")
	return nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "Hello, Serverless!")
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only POST method is allowed"})
		return
	}

	var reqData RequestData
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Сохраняем в БД
	var id int
	timestamp := time.Now()
	insertQuery := `
	INSERT INTO requests ("user", action, timestamp) 
	VALUES ($1, $2, $3) 
	RETURNING id;
	`

	err = db.QueryRow(insertQuery, reqData.User, reqData.Action, timestamp).Scan(&id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save data"})
		log.Println("Database error:", err)
		return
	}

	respData := ResponseData{
		ID:        id,
		User:      reqData.User,
		Action:    reqData.Action,
		Timestamp: timestamp,
		Message:   fmt.Sprintf("Saved: user=%s, action=%s", reqData.User, reqData.Action),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respData)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only GET method is allowed"})
		return
	}

	query := `SELECT id, "user", action, timestamp FROM requests ORDER BY id DESC LIMIT 100;`
	rows, err := db.Query(query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch data"})
		log.Println("Database error:", err)
		return
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.ID, &record.User, &record.Action, &record.Timestamp)
		if err != nil {
			log.Println("Scan error:", err)
			continue
		}
		records = append(records, record)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":   len(records),
		"records": records,
	})
}

func main() {
	// Инициализируем БД
	err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/list", listHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Сервер запущен на порту %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
