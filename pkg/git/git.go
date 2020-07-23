package git

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	g "github.com/gogits/git-module"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

// Config represents the inputs needed to set up a Repo.
type Config struct {
	Username   string
	Token      string
	BaseAPIURL string
}

// CreateReleaseRequest is the fields necessary for the body of a
// post request to create a release in Github.
type CreateReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

// Repo represents a git repository, which receives convenience methods
// for retrieving code.
type Repo struct {
	username   string
	token      string
	logger     *log.Entry
	baseAPIURL string
}

// New creates a Repo struct using a Config struct.
func New(c Config, logger *log.Entry) *Repo {
	return &Repo{
		username:   c.Username,
		token:      c.Token,
		logger:     logger,
		baseAPIURL: c.BaseAPIURL,
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
	p := url.QueryEscape(r.token)
	authURL := fmt.Sprintf("%s://%s:%s@%s", splitURL[0], r.username, p, splitURL[1])
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		return "", errors.Wrap(err, "clone url was not valid")
	}
	return parsedURL.String(), nil
}

// CreateRelease sends a request to github using the passed body and payload structs
// to create a new release on a repo
func (r *Repo) CreateRelease(ctx context.Context, req CreateReleaseRequest, payload github.PushPayload) error {
	endpoint := fmt.Sprintf(
		"%s/repos/%s/%s/releases",
		r.baseAPIURL,
		payload.Repository.Owner.Login,
		payload.Repository.Name,
	)

	// make a release
	// https://docs.github.com/en/rest/reference/repos#create-a-release
	client := resty.New()
	resp, err := client.R().
		SetBody(req).
		SetContext(ctx).
		SetHeader("accept", "application/vnd.github.v3+json").
		SetBasicAuth(r.username, r.token).
		Post(endpoint)

	if err != nil {
		return errors.Wrap(err, "failed to create release")
	}

	// default status code is 201, if it's not, and there wasn't an error
	// something still likely went wrong
	if resp.StatusCode() != int(201) {
		msg := fmt.Sprintf(
			"tag for repo %s on commit %s failed with code %d\n",
			payload.Repository.FullName,
			payload.HeadCommit.ID,
			resp.StatusCode(),
		)
		return errors.New(msg)
	}

	return nil
}
