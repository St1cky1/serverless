package main

import (
	"fmt"
	"net/http"
	"os"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "Hello, Serverless!")
}

func main() {
	http.HandleFunc("/", helloHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Сервер запущен на порту %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
