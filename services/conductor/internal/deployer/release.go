package deployer

import "k8s.io/apimachinery/pkg/util/rand"

type TriggerKind string

const (
	TriggerManual    TriggerKind = "manual"
	TriggerRollback  TriggerKind = "rollback"
	TriggerPromotion TriggerKind = "promotion"
	TriggerPush      TriggerKind = "push"
)

type ReleaseMeta struct {
	ID      string
	Trigger TriggerKind
	Actor   string
}

func NewRelease(trigger TriggerKind, actor string) ReleaseMeta {
	return ReleaseMeta{
		ID:      "rel-" + rand.String(6),
		Trigger: trigger,
		Actor:   actor,
	}
}
