package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

/*
Fake http.ResponseWriter to pass into tests. Saves any data written in the
http response, and the status code. Disregardes headers
*/
type fakeHttpWriter struct {
	status int
	buf    bytes.Buffer
}

func FakeResponse() *fakeHttpWriter {
	return &fakeHttpWriter{}
}

func (w *fakeHttpWriter) Header() http.Header {
	// can't return nil since it might be used
	return http.Header{}
}

func (w *fakeHttpWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *fakeHttpWriter) WriteHeader(status int) {
	w.status = status
}

/*
Helper to build HTTP requests for tests
*/
func FakeRequest(method, url string, body interface{}) *http.Request {
	var req *http.Request
	if body != nil {
		data := &bytes.Buffer{}
		json.NewEncoder(data).Encode(body)
		req, _ = http.NewRequest(method, url, data)
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	return req
}

func TestHandlesReturnedError(t *testing.T) {
	// Note 12345 isn't a real HTTP status code. It's just for testing.
	status_code := 12345
	message := "This message is okay in tests!"

	real_handler := func(w http.ResponseWriter, r *http.Request) HttpError {
		return Error(status_code, message)
	}
	err_handler := ErrorHandler(real_handler)
	w := FakeResponse()
	err_handler(w, nil)

	// check status code is set
	if w.status != status_code {
		t.Errorf("Expected status code %v != %v", status_code, w.status)
	}

	// check message is set
	actual_message := strings.TrimSpace(w.buf.String())
	if actual_message != message {
		t.Errorf("Expected error message %q != %q", message, actual_message)
	}
}

func TestHandlesThrownError(t *testing.T) {
	real_handler := func(w http.ResponseWriter, r *http.Request) HttpError {
		panic("this message is okay in tests!")
	}
	err_handler := ErrorHandler(real_handler)
	w := FakeResponse()
	err_handler(w, nil)

	// check status code is set
	if w.status != ServerError.status {
		t.Errorf("Expected status code %v != %v", ServerError.status, w.status)
	}

	// check message is set
	actual_message := strings.TrimSpace(w.buf.String())
	if actual_message != ServerError.message {
		t.Errorf("Expected error message %q != %q", ServerError.message, actual_message)
	}
}
