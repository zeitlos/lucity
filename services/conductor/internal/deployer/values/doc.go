// Package values is the in-memory model of a lucity-app release's Helm
// values. The mutators (Create*, Set*, Add*, Remove*) shape an Env loaded
// from the release and only enforce transition rules, the ones that need
// the state before and after the change, such as a volume that must not
// shrink. Every invariant of the resulting state is checked once, in
// Validate, which the deployer runs on the whole Env before it applies
// the release.
package values
