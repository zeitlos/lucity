package agentrunner

import "errors"

// ErrConflict marks a request the runner refused because of the session's
// state, like a message while the session is working on one.
var ErrConflict = errors.New("the session can't take this right now")

type conflictError struct {
	user error
}

func (e conflictError) Error() string {
	return e.user.Error()
}

func (e conflictError) Unwrap() []error {
	return []error{e.user, ErrConflict}
}
