package agentrunner

import "strings"

const triggerActorPrefix = "trigger:"

// TriggerActor is how a trigger appears as the author of a session or a
// message. A member appears as their user id.
func TriggerActor(trigger string) string {
	return triggerActorPrefix + trigger
}

// ActorTrigger returns the trigger an actor names, if it names one.
func ActorTrigger(actor string) (string, bool) {
	return strings.CutPrefix(actor, triggerActorPrefix)
}
