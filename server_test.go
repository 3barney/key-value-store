package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestJSON(t *testing.T) {

	header := http.Header{}
	headerKey := "Content-Type"
	headerValue := "application/json; charset=utf-8"
	header.Add(headerKey, headerValue)

	// Setup table testing using anonymous struct
	testCases := []struct {
		givenInput     interface{} // map[string]string
		header         http.Header
		expectedOutput string
	}{
		{map[string]string{"hello": "world"}, header, `{"hello":"world"}`},
		{map[string]string{"hello": "tables"}, header, `{"hello":"tables"}`},
		{make(chan bool), header, `{"error":"json: unsupported type: chan bool"}`},
	}

	for _, test := range testCases {
		recorder := httptest.NewRecorder() // Allow writing to ResponseWriter
		JSON(recorder, test.givenInput)

		response := recorder.Result() // Get result from ResponseWriter
		defer response.Body.Close()

		output, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err) // fatalf Fail to continue but
		}

		if string(output) != test.expectedOutput {
			t.Errorf("Got %s, expected %s", string(output), test.expectedOutput)
		}

		// Check for header setup
		if contentType := response.Header.Get(headerKey); contentType != headerValue {
			t.Errorf("Got %s, expected %s", contentType, headerValue)
		}
	}
}

// TODO: implement edge cases using table testing
func TestGet(t *testing.T) {
	makeStorage(t)
	defer cleanupStorage(t)

	key := "key"
	value := "value"
	encodedKey := base64.URLEncoding.EncodeToString([]byte(key))
	encodedValue := base64.URLEncoding.EncodeToString([]byte(value))
	fileContents, _ := json.Marshal(map[string]string{encodedKey: encodedValue})
	err := os.WriteFile(StoragePath+"/data.json", fileContents, 0644)
	if err != nil {
		return
	}

	got, err := Get(context.Background(), key)
	if err != nil {
		t.Errorf("Received unexpected error: %s", err)
	}
	if got != value {
		t.Errorf("Got %s, expected %s", got, value)
	}
}

// Run go test -bench .
func BenchmarkGet(b *testing.B) {
	makeStorage(b)
	defer cleanupStorage(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Get(context.Background(), "key1")
	}
}

// Simple Intergration Test
// GET "key" -> SET -> GET -> DELETE -> GET
func TestGetSetDelete(t *testing.T) {
	// t.Parallel()
	makeStorage(t)
	defer cleanupStorage(t)
	ctx := context.Background()

	key := "key"
	value := "value"

	if out, err := Get(ctx, key); err != nil || out != "" {
		t.Fatalf("First Get returned unexpected result, out: %q, error: %s", out, err)
	}

	if err := Set(ctx, key, value); err != nil {
		t.Fatalf("Set returned unexpected error: %s", err)
	}

	if out, err := Get(ctx, key); err != nil || out != value {
		t.Fatalf("Second Get returned unexpected result, out: %q, error: %s", out, err)
	}

	if err := Delete(ctx, key); err != nil {
		t.Fatalf("Delete returned unexpected error: %s", err)
	}

	if out, err := Get(ctx, key); err != nil || out != "" {
		t.Fatalf("Third Get returned unexpected result, out: %q, error: %s", out, err)
	}
}

// Accect both Testing and Benchmark tests
func makeStorage(tb testing.TB) {
	err := os.Mkdir("testdata", 0755)
	if err != nil && !os.IsExist(err) {
		tb.Fatalf("Couldn't create directory testdata: %s", err)
	}

	StoragePath = "testdata"
}

func cleanupStorage(tb testing.TB) {
	if err := os.RemoveAll(StoragePath); err != nil {
		tb.Errorf("Failed to delete storage path: %s", StoragePath)
	}

	StoragePath = "/tmp/kv"
}

// OLD TEsting only implementation
//func makeStorage(t *testing.T) {
//	err := os.Mkdir("testdata", 0755)
//	if err != nil && !os.IsExist(err) {
//		tb.Fatalf("Couldn't create directory testdata: %s", err)
//	}
//
//	StoragePath = "testdata"
//}
