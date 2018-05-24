package errlog

import "errors"

var (
	ErrFormat       = errors.New("Format error")
	ErrNotFound     = errors.New("Not found")
	ErrInvalidParam = errors.New("Invalid param")
)
