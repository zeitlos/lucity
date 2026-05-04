package deployer

import "errors"

// ErrNotFound is returned by Backend methods (Get, Status, History,
// Rollback) when the addressed Target has no corresponding state in
// the underlying system. Apply and Remove never return this — Apply
// creates state and Remove is idempotent.
var ErrNotFound = errors.New("target not found")

// ErrInProgress is returned when the Backend cannot service a request
// because a previous mutating operation is still running on the same
// Target. Callers should retry after a short backoff. Implementations
// that serialize internally do not need to surface this — return only
// when the wait would exceed a reasonable bound.
var ErrInProgress = errors.New("operation in progress")
