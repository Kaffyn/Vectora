package test

import "errors"

var (
	// ErrTimeout is returned when a wait operation times out
	ErrTimeout = errors.New("operation timeout")

	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput is returned when input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrConnectionFailed is returned when connection fails
	ErrConnectionFailed = errors.New("connection failed")

	// ErrOperationFailed is returned when an operation fails
	ErrOperationFailed = errors.New("operation failed")
)
