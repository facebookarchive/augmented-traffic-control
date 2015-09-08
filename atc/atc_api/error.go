package main

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
	ServerError   = httpError{http.StatusInternalServerError, "Internal Server Error"}
	InvalidMethod = Errorf(http.StatusMethodNotAllowed, "Method not allowed")

	// Used to indicate to the error handler that the request already wrote the
	// HTTP status
	NoStatus = Errorf(-1, "Not a real error")
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

func Error(status int, message string) HttpError {
	return httpError{
		message: message,
		status:  status,
	}
}

func Errorf(status int, f string, things ...interface{}) HttpError {
	return httpError{
		message: fmt.Sprintf(f, things...),
		status:  status,
	}
}

func IsNoSuchItem(err error) bool {
	return NoSuchItemRegex.MatchString(err.Error())
}
