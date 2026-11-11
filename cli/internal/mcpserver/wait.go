package mcpserver

import (
	"context"
	"fmt"
	"time"
)

const (
	maxWait      = 300
	waitInterval = 4 * time.Second
)

func validateWait(seconds int) error {
	if seconds < 0 || seconds > maxWait {
		return fmt.Errorf("wait must be between 0 and %d seconds", maxWait)
	}
	return nil
}

func pollUntil[T any](ctx context.Context, seconds int, check func() (T, bool, error)) (T, bool, error) {
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)

	for {
		result, settled, err := check()
		if err != nil || settled || seconds == 0 {
			return result, settled, err
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return result, false, nil
		}

		select {
		case <-ctx.Done():
			return result, false, nil
		case <-time.After(min(waitInterval, remaining)):
		}
	}
}

func waitNote(seconds int, settled bool, status any) string {
	if seconds == 0 || settled {
		return ""
	}
	return fmt.Sprintf("still %v after waiting %ds; call again with wait to keep waiting", status, seconds)
}
