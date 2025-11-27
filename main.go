package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// RequestData структура для входящих данных
type RequestData struct {
	User   string `json:"user"`
	Action string `json:"action"`
}

// ResponseData структура для ответа
type ResponseData struct {
	User      string `json:"user"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
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

	respData := ResponseData{
		User:      reqData.User,
		Action:    reqData.Action,
		Timestamp: fmt.Sprintf("%v", "2024"),
		Message:   fmt.Sprintf("Received: user=%s, action=%s", reqData.User, reqData.Action),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respData)
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/echo", echoHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Сервер запущен на порту %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
