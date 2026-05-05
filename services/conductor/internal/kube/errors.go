package kube

import "errors"

// ErrNoLabels indicates that a namespace is missing one or more of the
// required Lucity ownership labels (workspace, project, environment).
// Callers typically treat this as "not a managed namespace" rather
// than as an operational failure.
var ErrNoLabels = errors.New("namespace is missing required lucity.dev labels")
