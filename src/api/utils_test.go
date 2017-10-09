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
	PanicStack = false
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
	statusCode := 12345
	message := "This message is okay in tests!"

	realHandler := func(w http.ResponseWriter, r *http.Request) HttpError {
		return HttpErrorf(statusCode, message)
	}
	errHandler := ErrorHandler(realHandler)
	w := httptest.NewRecorder()
	errHandler(w, nil)

	// check status code is set
	if w.Code != statusCode {
		t.Errorf("Expected status code %v != %v", statusCode, w.Code)
	}

	// check message is set
	actualMessage := strings.TrimSpace(w.Body.String())
	if actualMessage != message {
		t.Errorf("Expected error message %q != %q", message, actualMessage)
	}
}

func TestHandlesThrownError(t *testing.T) {
	realHandler := func(w http.ResponseWriter, r *http.Request) HttpError {
		panic("this message is okay in tests!")
	}
	errHandler := ErrorHandler(realHandler)
	w := FakeResponse()
	errHandler(w, nil)

	// check status code is set
	if w.Code != ServerError.Status() {
		t.Errorf("Expected status code %v != %v", ServerError.Status(), w.Code)
	}

	// check message is set
	actualMessage := strings.TrimSpace(w.Body.String())
	if actualMessage != ServerError.Error() {
		t.Errorf("Expected error message %q != %q", ServerError.Error(), actualMessage)
	}
}

func TestGetsProxiedAddr(t *testing.T) {
	testProxy := func(clientAddr string,
		headerAddrs []string, serverAddr string) (string, error) {
		r, _ := http.NewRequest("GET", "/", nil)
		r.RemoteAddr = clientAddr + ":0" // net.SplitHostPort requires a port
		for _, headerAddr := range headerAddrs {
			r.Header.Add("X_HTTP_REAL_IP", headerAddr)
		}
		srv := &Server{AtcApiOptions: AtcApiOptions{ProxyAddr: serverAddr}}
		addr, err := getProxiedClientAddr(srv, r)
		if addr != nil {
			return addr.String(), err
		}
		return "", err
	}

	// Neither the server nor the client are proxied.
	addr, err := testProxy("1.1.1.1", []string{}, "")
	if err != nil {
		t.Error(err)
	} else if addr != "1.1.1.1" {
		t.Errorf("Wrong proxy address: %q", addr)
	}

	// Both the client and the server are proxied.
	// There is multiple X_HTTP_REAL_IP headers so we fail.
	addr, err = testProxy("1.1.1.1", []string{"2.2.2.2", "3.3.3.3"}, "1.1.1.1")
	if err == nil {
		t.Errorf("Proxy address should be invalid: %q", addr)
	}

	// Both the client and the server are proxied.
	addr, err = testProxy("1.1.1.1", []string{"2.2.2.2"}, "1.1.1.1")
	if err != nil {
		t.Error(err)
	} else if addr != "2.2.2.2" {
		t.Errorf("Wrong proxy address: %q", addr)
	}

	// Server expects a proxy, but client doesn't send one
	addr, err = testProxy("this.message.ok.in.tests", []string{}, "2.2.2.2")
	if err == nil {
		t.Errorf("Proxy address should be invalid: %q", addr)
	}

	// Client sends a proxy, but the server doesn't expect it
	addr, err = testProxy("this.message.ok.in.tests", []string{"2.2.2.2"}, "")
	if err == nil {
		t.Errorf("Proxy address should be invalid: %q", addr)
	}
}
