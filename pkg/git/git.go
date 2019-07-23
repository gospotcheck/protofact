package git

import (
	"fmt"
	"net/url"
	"strings"

	g "github.com/gogits/git-module"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

// Config represents the inputs needed to set up a Repo.
type Config struct {
	Username string
	Password string
}

// Repo represents a git repository, which receives convenience methods
// for retrieving code.
type Repo struct {
	username string
	password string
	logger   *log.Entry
}

// New creates a Repo struct using a Config struct.
func New(c Config, logger *log.Entry) *Repo {
	return &Repo{
		username: c.Username,
		password: c.Password,
		logger:   logger,
	}
}

// CloneWithCheckout uses a github Push Event payload to clone down a repository
// and immediately checkout HEAD on the branch that caused that Push Event.
func (r *Repo) CloneWithCheckout(tmpDir string, payload github.PushPayload) error {
	branch := g.RefEndName(payload.Ref)
	url, err := r.CreateAuthenticatedURL(payload.Repository.CloneURL)
	if err != nil {
		return errors.Wrap(err, "could not create authenticated url")
	}
	err = g.Clone(url, tmpDir, g.CloneRepoOptions{Branch: branch})
	if err != nil {
		errMsg := fmt.Sprintf("could not clone repo %s on branch %s to temp dir %s", payload.Repository.CloneURL, payload.Ref, tmpDir)
		return errors.Wrap(err, errMsg)
	}
	r.logger.Debug(fmt.Sprintf("successfully cloned %s to temp dir %s", url, tmpDir))
	return nil
}

// CreateAuthenticatedURL adds the username and password from the instantiation of
// the Repo to use https authentication on all calls to the origin.
func (r *Repo) CreateAuthenticatedURL(cloneURL string) (string, error) {
	splitURL := strings.Split(cloneURL, "://")
	// passwords often have characters that need escaping in them
	p := url.QueryEscape(r.password)
	authURL := fmt.Sprintf("%s://%s:%s@%s", splitURL[0], r.username, p, splitURL[1])
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		return "", errors.Wrap(err, "clone url was not valid")
	}
	return parsedURL.String(), nil
}
