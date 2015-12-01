package api

import (
	"fmt"
	"net/http"
	"regexp"
)

var (
	// Shared constant between ATCD and ATC_API
	// Make sure to change in both places
	NoSuchItemRegex = regexp.MustCompile(`Internal error processing [a-zA-Z0-9_]*: NO_SUCH_ITEM`)
)

var (
	ServerError = HttpErrorf(http.StatusInternalServerError, "Internal Server Error")

	// Used to indicate to the error handler that the request already wrote the
	// HTTP status
	NoStatus = HttpErrorf(-1, "Not a real error")
)

type HttpError interface {
	error
	Status() int
}

type httpError struct {
	status  int
	message string
}

func (e httpError) Error() string {
	return e.message
}

func (e httpError) Status() int {
	return e.status
}

func HttpErrorf(status int, f string, things ...interface{}) HttpError {
	return httpError{status, fmt.Sprintf(f, things...)}
}

func IsNoSuchItem(err error) bool {
	return NoSuchItemRegex.MatchString(err.Error())
}

func InvalidMethod(r *http.Request) HttpError {
	return HttpErrorf(http.StatusMethodNotAllowed, "Method not allowed: %s %v", r.Method, r.URL)
}
