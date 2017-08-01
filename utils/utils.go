package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// NewErr combines a custom comment and error into a new formatted error.
func NewErr(input string, err error) error {
	if err != nil {
		return errors.New(input + " " + err.Error())
	}
	return errors.New(input)
}

// ComboErr combines a custom comment and two error messages into one
// formatted error.
func ComboErr(input string, first error, second error) error {
	var res string
	switch {
	case first != nil && second != nil:
		res = fmt.Sprintf("%s %s. %s", input, first.Error(),
			second.Error())
		return errors.New(res)
	case first != nil:
		return NewErr(input, first)
	case second != nil:
		return NewErr(input, second)
	}
	return errors.New(input)
}

// GetDirPath gets a directory path name from the URL and formats it
func GetDirPath(r *http.Request) string {
	res := r.URL.Query().Get("path-name")
	if res == "" {
		return res
	}
	if !strings.HasPrefix(res, "/") {
		res = "/" + res
	}
	if !strings.HasSuffix(res, "/") {
		res += "/"
	}
	return res
}
