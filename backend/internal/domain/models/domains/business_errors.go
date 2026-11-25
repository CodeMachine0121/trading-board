package domains

import "errors"

// The kinds every business refusal falls into; the controller maps each to one status code.
var (
	ErrValidation             = errors.New("validation failed")
	ErrNotFound               = errors.New("not found")
	ErrConflict               = errors.New("conflict")
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrForbidden              = errors.New("forbidden")
	ErrTooManyAttempts        = errors.New("too many attempts")
)

// BusinessError carries the contract's error code and a message for the owner, and unwraps to its kind.
type BusinessError struct {
	kind    error
	Code    string
	Message string
}

func (businessError BusinessError) Error() string {
	return businessError.Message
}

func (businessError BusinessError) Unwrap() error {
	return businessError.kind
}

func ValidationFailed(code string, message string) error {
	return BusinessError{kind: ErrValidation, Code: code, Message: message}
}

func NotFound(code string, message string) error {
	return BusinessError{kind: ErrNotFound, Code: code, Message: message}
}

func Conflict(code string, message string) error {
	return BusinessError{kind: ErrConflict, Code: code, Message: message}
}

func AuthenticationRequired(message string) error {
	return BusinessError{kind: ErrAuthenticationRequired, Code: "unauthenticated", Message: message}
}

func Forbidden(message string) error {
	return BusinessError{kind: ErrForbidden, Code: "forbidden", Message: message}
}
