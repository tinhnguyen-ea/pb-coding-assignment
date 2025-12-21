package entities

import "errors"

var (
	ErrFxService       = errors.New("fx service error")
	ErrDBService       = errors.New("db service error")
	ErrBillingNotFound = errors.New("billing not found")
)
