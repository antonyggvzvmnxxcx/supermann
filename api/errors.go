package api

import (
	"fmt"
	"net/http"
	"net/url"
)

type apiErr struct {
	Status           int        `json:"status,omitempty"`
	Code             string     `json:"code,omitempty"`
	Desc             string     `json:"desc,omitempty"`
	ValidationErrors url.Values `json:"validation_errors,omitempty"`
}

// ErrUnsupportedMethod ...
var ErrUnsupportedMethod = apiErr{Status: http.StatusMethodNotAllowed,
	Code: "unsupported_method", Desc: "Unsupported HTTP method"}

func newNotFoundErr(format string, args ...interface{}) *apiErr {
	return &apiErr{Status: http.StatusNotFound, Code: "not_found", Desc: fmt.Sprintf(format, args...)}
}

func newInvalidArgumentErr(errors url.Values) *apiErr {
	return &apiErr{Status: http.StatusBadRequest, Code: "invalid_arguments", ValidationErrors: errors}
}

func newInternalServerErr(err error) *apiErr {
	return &apiErr{Status: http.StatusInternalServerError, Code: "internal_error", Desc: err.Error()}
}

func newNotImplemented(method, uri string) *apiErr {
	desc := fmt.Sprintf("%s Not Implemented on %s", method, uri)
	return &apiErr{Status: http.StatusNotImplemented, Code: "not_implemented", Desc: desc}
}
