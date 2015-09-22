package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func init() {
	PANIC_STACK = false
}

func FakeResponse() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
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
		return HttpErrorf(status_code, message)
	}
	err_handler := ErrorHandler(real_handler)
	w := httptest.NewRecorder()
	err_handler(w, nil)

	// check status code is set
	if w.Code != status_code {
		t.Errorf("Expected status code %v != %v", status_code, w.Code)
	}

	// check message is set
	actual_message := strings.TrimSpace(w.Body.String())
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
	if w.Code != ServerError.Status() {
		t.Errorf("Expected status code %v != %v", ServerError.Status(), w.Code)
	}

	// check message is set
	actual_message := strings.TrimSpace(w.Body.String())
	if actual_message != ServerError.Error() {
		t.Errorf("Expected error message %q != %q", ServerError.Error(), actual_message)
	}
}
