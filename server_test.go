package main

import (
	"io"
	"net/http"
	"net/http/httptest"
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
