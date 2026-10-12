// Package usererror marks errors whose message is meant for the API caller.
// The GraphQL layer shows these messages verbatim and replaces every other
// error with a generic one, so internal details never reach a client.
package usererror

import "fmt"

// Error is an error whose message is safe to show to the API caller.
type Error struct {
	message string
}

// New returns an error whose message is safe to show to the API caller.
func New(message string) error {
	return &Error{message: message}
}

// Errorf formats an error whose message is safe to show to the API caller.
// It never wraps: pass causes with %v, and only when their text is safe too.
func Errorf(format string, args ...any) error {
	return &Error{message: fmt.Sprintf(format, args...)}
}

func (e *Error) Error() string {
	return e.message
}
