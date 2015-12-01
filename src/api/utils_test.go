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

func TestGetsProxiedAddr(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "1.1.1.1:1234"
	srv := &Server{proxy_addr: ""}

	// Non-proxied request should work
	addr, err := getProxiedClientAddr(srv, r)
	if err != nil {
		t.Error(err)
	} else if addr != "1.1.1.1" {
		t.Errorf("Wrong proxy address: %q", addr)
	}

	// Simulate a proxied request
	r.Header.Set("X_HTTP_REAL_IP", "2.2.2.2")

	// Works without a proxy address set
	addr, err = getProxiedClientAddr(srv, r)
	if err != nil {
		t.Error(err)
	} else if addr != "2.2.2.2" {
		t.Errorf("Wrong proxy address: %q", addr)
	}

	// Works with correct proxy address set
	srv.proxy_addr = "1.1.1.1"
	addr, err = getProxiedClientAddr(srv, r)
	if err != nil {
		t.Error(err)
	} else if addr != "2.2.2.2" {
		t.Errorf("Wrong proxy address: %q", addr)
	}

	srv.proxy_addr = "3.3.3.3"
	r.RemoteAddr = "message.ok.in.tests:1234"
	// Shouldn't work if proxy address is set wrong
	_, err = getProxiedClientAddr(srv, r)
	if err == nil {
		t.Errorf("Proxy address should be invalid: %q", addr)
	}
}
