package main

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"os"
)

func main() {
	const VARIABLE_PORT = "PORT"
	port := "8080"
	if fromEnvironment := os.Getenv(VARIABLE_PORT); fromEnvironment != "" {
		port = fromEnvironment
	}

	log.Printf("Starting up on http://localhost:%s", port)

	router := chi.NewRouter()
	router.Get("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		JSON(responseWriter, map[string]string{"hello": "world"})
	})

	log.Fatal(http.ListenAndServe(":"+port, router))
}

// JSON encodes data to json and writes it to the http response.
func JSON(writer http.ResponseWriter, data interface{}) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, error := json.Marshal(data)

	if error != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		JSON(writer, map[string]string{"error": error.Error()})
		return
	}
	writer.Write(body)
}
