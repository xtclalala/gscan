package util

import (
	"github.com/pkg/errors"
	"strings"
)

var (
	ErrEmptyInput = errors.New("empty data")
)

func Sanitize(data string) (string, error) {
	data = strings.Trim(data, "\n\t\"'")
	if data == "" {
		return "", ErrEmptyInput
	}
	return data, nil
}
