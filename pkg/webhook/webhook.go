package webhook

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/webhooks.v5/github"
)

// Config represents config values for the webhook parser.
type Config struct {
	Secret string
}

// Parser represents a service wrapping the github package
// providing convenience methods for interacting with Push Events.
type Parser struct {
	webhook *github.Webhook
}

// NewParser creates a Parser struct with injection of options.
// In production, secure should be true, but for testing and local development
// it can be false.
func NewParser(secure bool, config Config) (*Parser, error) {
	var webhook *github.Webhook
	var err error
	if secure {
		if strings.Compare(config.Secret, "") == 0 {
			return &Parser{}, errors.New("if parser is secure, you must pass through a non-empty string")
		}
		webhook, err = github.New(github.Options.Secret(config.Secret))
	} else {
		webhook, err = github.New()
	}
	if err != nil {
		return &Parser{}, errors.Wrap(err, "could not create new github webhook")
	}

	return &Parser{
		webhook: webhook,
	}, nil
}

// IsPingEvent determines if the event is a Ping event by attempting
// to parse it.
func (p *Parser) IsPingEvent(r *http.Request) bool {
	_, err := p.webhook.Parse(r, github.PingEvent)
	if err != nil {
		return false
	}
	return true
}

// ValidateAndParsePushEvent is a convenience method for receiving an http request,
// ensuring it is specifically a Push Event payload and returning it as that struct.
func (p *Parser) ValidateAndParsePushEvent(r *http.Request) (github.PushPayload, error) {
	payload, err := p.webhook.Parse(r, github.PushEvent)
	if err != nil {
		if err == github.ErrEventNotFound {
			return github.PushPayload{}, errors.Wrap(err, "event was not a push event")
		}

		return github.PushPayload{}, errors.Wrap(err, "error parsing webhook payload")
	}
	event := payload.(github.PushPayload)
	return event, nil
}
