package git

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	g "github.com/gogits/git-module"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	hooks "gopkg.in/go-playground/webhooks.v5/github"
)

// Config represents the inputs needed to set up a Repo.
type Config struct {
	Username string
	Token    string
	Email    string
}

// Repo represents a git repository, which receives convenience methods
// for retrieving code.
type Repo struct {
	username string
	token    string
	email    string
	logger   *log.Entry
	client   *github.Client
}

// New creates a Repo struct using a Config struct.
func New(ctx context.Context, c Config, logger *log.Entry) *Repo {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Repo{
		username: c.Username,
		token:    c.Token,
		logger:   logger,
		client:   client,
		email:    c.Email,
	}
}

// SetGitConfig sets the email and name for usage when sending in git commits
func (r *Repo) SetGitConfig() error {
	emailCmd := exec.Command("git", "config", "--global", r.email)
	out, err := emailCmd.CombinedOutput()
	r.logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		errMessage := fmt.Sprintf("error setting git config email: %s\n", out)
		return errors.Wrap(err, errMessage)
	}

	nameCmd := exec.Command("git", "config", "--global", "protofact")
	out, err = nameCmd.CombinedOutput()
	r.logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		errMessage := fmt.Sprintf("error setting git config name: %s\n", out)
		return errors.Wrap(err, errMessage)
	}

	return nil
}

// CloneWithCheckout uses a github Push Event payload to clone down a repository
// and immediately checkout HEAD on the branch that caused that Push Event.
func (r *Repo) CloneWithCheckout(tmpDir string, payload hooks.PushPayload) error {
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
func (r *Repo) CreateRelease(ctx context.Context, owner, repo string, rel *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	rel, _, err := r.client.Repositories.CreateRelease(ctx, "gospotcheck", "protofact", rel)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rel, nil
}

// CreateTag makes an annotated git tag on a repo.
func (r Repo) CreateTag(dir, version, msg string) error {
	tagCmd := exec.Command("git", "tag", "-a", version, "-m", msg)
	tagCmd.Dir = dir
	out, err := tagCmd.CombinedOutput()
	r.logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		errMessage := fmt.Sprintf("error tagging git repo: %s\n", out)
		return errors.Wrap(err, errMessage)
	}

	return nil
}

// PushTags pushes a repo with no other changes but tags up to the origin.
func (r Repo) PushTags(dir string) error {
	pushCmd := exec.Command("git", "push", "--follow-tags")
	pushCmd.Dir = dir
	out, err := pushCmd.CombinedOutput()
	r.logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		errMessage := fmt.Sprintf("error pushing git tags: %s\n", out)
		return errors.Wrap(err, errMessage)
	}

	return nil
}

// UploadReleaseAsset uploads a file to a Github release
func (r Repo) UploadReleaseAsset(ctx context.Context, owner, repo string, releaseID int64, opts *github.UploadOptions, file *os.File) error {
	_, _, err := r.client.Repositories.UploadReleaseAsset(ctx, "gospotcheck", "protofact", releaseID, opts, file)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
