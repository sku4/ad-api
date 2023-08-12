package model

import "errors"

var (
	ErrSubAlreadyExists = errors.New("subscription already exists")
	ErrResultNotFound   = errors.New("result not found")
)
