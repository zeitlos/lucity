package github

import (
	"fmt"
	"net/http"

	gh "github.com/google/go-github/v68/github"
)

// Event is a normalized representation of a GitHub webhook event.
type Event struct {
	Type         string // e.g. "push", "pull_request", "installation"
	Action       string // e.g. "opened", "closed", "synchronize"
	RepoFullName string // e.g. "zeitlos/myapp"
	RepoCloneURL string
	Ref          string // e.g. "refs/heads/main"
	CommitSHA    string
	Sender       string // GitHub login of the actor
	PRNumber     int
}

// ValidateAndParse validates the webhook HMAC signature and parses
// the payload into a normalized Event.
func ValidateAndParse(secret []byte, r *http.Request) (*Event, error) {
	payload, err := gh.ValidatePayload(r, secret)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	eventType := gh.WebHookType(r)
	raw, err := gh.ParseWebHook(eventType, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	event := &Event{Type: eventType}

	switch e := raw.(type) {
	case *gh.PushEvent:
		event.RepoFullName = e.GetRepo().GetFullName()
		event.RepoCloneURL = e.GetRepo().GetCloneURL()
		event.Ref = e.GetRef()
		event.CommitSHA = e.GetAfter()
		event.Sender = e.GetSender().GetLogin()

	case *gh.PullRequestEvent:
		event.Action = e.GetAction()
		event.RepoFullName = e.GetRepo().GetFullName()
		event.RepoCloneURL = e.GetRepo().GetCloneURL()
		event.Ref = e.GetPullRequest().GetHead().GetRef()
		event.CommitSHA = e.GetPullRequest().GetHead().GetSHA()
		event.Sender = e.GetSender().GetLogin()
		event.PRNumber = e.GetNumber()

	case *gh.InstallationEvent:
		event.Action = e.GetAction()
		event.Sender = e.GetSender().GetLogin()

	case *gh.InstallationRepositoriesEvent:
		event.Action = e.GetAction()
		event.Sender = e.GetSender().GetLogin()

	default:
		// Unknown event type — return what we have (Type is set)
	}

	return event, nil
}
