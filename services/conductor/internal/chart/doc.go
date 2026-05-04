// Package chart turns a typed Spec into the values map consumed by
// the lucity-app Helm chart.
//
// Both the argo and (future) helm Backend implementations call into
// this package; the chart shape is platform-internal, not vendor-
// specific. Keeping rendering here means new fields on Spec only
// require touching one place.
//
// Populated in phase 2 from services/packager/chart/.
package chart
