package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/go-chi/chi"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var StoragePath = "/tmp"
var data = map[string]string{}

func main() {
	const VARIABLE_PORT = "PORT"
	port := "8080"
	if fromEnvironment := os.Getenv(VARIABLE_PORT); fromEnvironment != "" {
		port = fromEnvironment
	}

	log.Printf("Starting up on http://localhost:%s", port)

	router := chi.NewRouter()

	// Handler for /
	router.Get("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		JSON(responseWriter, map[string]string{"hello": "world"})
	})

	// Save a value
	router.Post("/key/{key}", func(responseWriter http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")

		body, err := io.ReadAll(request.Body) // TODO: Set max size before reading this to memory
		if err != nil {
			responseWriter.WriteHeader(http.StatusInternalServerError)
			JSON(responseWriter, map[string]string{"error": err.Error()})
			return
		}

		err = Set(request.Context(), key, string(body))
		if err != nil {
			responseWriter.WriteHeader(http.StatusInternalServerError)
			JSON(responseWriter, map[string]string{"error": err.Error()})
			return
		}

		JSON(responseWriter, map[string]string{"status": "success"})
	})

	// handler for Getting a value by key
	router.Get("/key/{key}", func(responseWriter http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")

		data, err := Get(request.Context(), key)
		if err != nil {
			responseWriter.WriteHeader(http.StatusInternalServerError)
			JSON(responseWriter, map[string]string{"error": err.Error()})
			return
		}
		responseWriter.Write([]byte(data))
	})

	// Delete an Item by it's key
	router.Delete("/key/{key}", func(responseWriter http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")

		err := Delete(request.Context(), key)
		if err != nil {
			responseWriter.WriteHeader(http.StatusInternalServerError)
			JSON(responseWriter, map[string]string{"error": err.Error()})
			return
		}

		JSON(responseWriter, map[string]string{"status": "success"})
	})

	log.Fatal(http.ListenAndServe(":"+port, router))
}

func Get(ctx context.Context, key string) (string, error) {
	data, err := loadData(ctx)
	if err != nil {
		return "", err
	}

	return data[key], nil
}

func Set(ctx context.Context, key, value string) error {
	data, err := loadData(ctx)

	if err != nil {
		return err
	}

	data[key] = value
	if err := saveData(ctx, data); err != nil {
		return err
	}

	return nil
}

func Delete(ctx context.Context, key string) error {
	data, err := loadData(ctx)

	if err != nil {
		return err
	}

	delete(data, key)
	if err := saveData(ctx, data); err != nil {
		return err
	}

	return nil
}

// JSON encodes data to json and writes it to the http response.
func JSON(writer http.ResponseWriter, data interface{}) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, err := json.Marshal(data)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		JSON(writer, map[string]string{"error": err.Error()})
		return
	}
	writer.Write(body)
}

func dataPath() string {
	return filepath.Join(StoragePath, "data.json")
}

func loadData(ctx context.Context) (map[string]string, error) {
	empty := map[string]string{}
	emptyData, err := encode(map[string]string{})
	if err != nil {
		return empty, err
	}

	// First check if the folder exists and create it if it is missing.
	if _, err := os.Stat(StoragePath); os.IsNotExist(err) {
		err := os.MkdirAll(StoragePath, 0755)
		if err != nil {
			return empty, err
		}
	}

	// Then check if the file exists and create it if it is missing.
	if _, err := os.Stat(dataPath()); os.IsNotExist(err) {
		err := os.WriteFile(dataPath(), emptyData, 0644)
		if err != nil {
			return empty, err
		}
	}

	content, err := os.ReadFile(dataPath())
	if err != nil {
		return empty, err
	}

	return decode(content)
}

func saveData(ctx context.Context, data map[string]string) error {

	if _, err := os.Stat(StoragePath); os.IsNotExist(err) {
		err = os.MkdirAll(StoragePath, 0755)
		if err != nil {
			return err
		}
	}

	encodedData, err := encode(data)
	if err != nil {
		return err
	}

	return os.WriteFile(dataPath(), encodedData, 0644)
}

func encode(data map[string]string) ([]byte, error) {
	encodedData := map[string]string{}
	for k, v := range data {
		ek := base64.URLEncoding.EncodeToString([]byte(k))
		ev := base64.URLEncoding.EncodeToString([]byte(v))
		encodedData[ek] = ev
	}

	return json.Marshal(encodedData)
}

func decode(data []byte) (map[string]string, error) {
	var jsonData map[string]string

	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	returnData := map[string]string{}
	for k, v := range jsonData {
		dk, err := base64.URLEncoding.DecodeString(k)
		if err != nil {
			return nil, err
		}

		dv, err := base64.URLEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}

		returnData[string(dk)] = string(dv)
	}

	return returnData, nil
}
