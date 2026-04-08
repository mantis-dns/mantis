package domain

import "fmt"

// ErrorCode represents a machine-readable error code for API mapping.
type ErrorCode string

const (
	CodeBlocked         ErrorCode = "BLOCKED"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeUpstreamTimeout ErrorCode = "UPSTREAM_TIMEOUT"
	CodePoolExhausted   ErrorCode = "POOL_EXHAUSTED"
	CodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	CodeRateLimited     ErrorCode = "RATE_LIMITED"
	CodeValidation      ErrorCode = "VALIDATION_ERROR"
	CodeInternal        ErrorCode = "INTERNAL_ERROR"
)

// DomainError is the base error type for all domain errors.
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Sentinel errors for errors.Is matching.
var (
	ErrBlocked         = &DomainError{Code: CodeBlocked, Message: "domain is blocked"}
	ErrNotFound        = &DomainError{Code: CodeNotFound, Message: "resource not found"}
	ErrUpstreamTimeout = &DomainError{Code: CodeUpstreamTimeout, Message: "upstream DNS timeout"}
	ErrPoolExhausted   = &DomainError{Code: CodePoolExhausted, Message: "DHCP address pool exhausted"}
	ErrUnauthorized    = &DomainError{Code: CodeUnauthorized, Message: "authentication required"}
	ErrRateLimited     = &DomainError{Code: CodeRateLimited, Message: "rate limit exceeded"}
	ErrValidation      = &DomainError{Code: CodeValidation, Message: "validation error"}
	ErrInternal        = &DomainError{Code: CodeInternal, Message: "internal error"}
)

// NewError creates a new DomainError wrapping an existing error.
func NewError(code ErrorCode, message string, err error) *DomainError {
	return &DomainError{Code: code, Message: message, Err: err}
}

// IsErrorCode checks whether an error has a specific error code.
func IsErrorCode(err error, code ErrorCode) bool {
	var de *DomainError
	if As(err, &de) {
		return de.Code == code
	}
	return false
}

// As is a convenience re-export — callers use domain.As instead of importing errors.
func As(err error, target interface{}) bool {
	type asInterface interface {
		As(interface{}) bool
	}
	// Walk the error chain.
	for err != nil {
		if x, ok := err.(asInterface); ok {
			if x.As(target) {
				return true
			}
		}
		// Direct type assertion.
		if de, ok := err.(*DomainError); ok {
			if t, ok2 := target.(**DomainError); ok2 {
				*t = de
				return true
			}
		}
		// Unwrap.
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
