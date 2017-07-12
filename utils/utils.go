package utils

import "errors"

func NewErr(input string, err error) error {
	return errors.New(input + " " + err.Error())
}
