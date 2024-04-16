package storage

import "errors"

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountExists   = errors.New("account exists")
	ErrRecordNotFound  = errors.New("record not found")
)
