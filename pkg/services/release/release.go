package release

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	g "github.com/gogits/git-module"
	"github.com/google/go-github/v32/github"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	hooks "gopkg.in/go-playground/webhooks.v5/github"
)

type fs interface {
	CreateUniqueTmpDir(parentPath string) (string, error)
	DeleteDir(string) error
}

type repo interface {
	CreateRelease(ctx context.Context, owner, repo string, rel *github.RepositoryRelease) (*github.RepositoryRelease, error)
	CloneWithCheckout(tmpDir string, payload hooks.PushPayload) error
	CreateTag(dir, version, msg string) error
	PushTags(dir string) error
}

type counters interface {
	AddPackagingErrors(labels prometheus.Labels, count float64)
	AddPackagingProcessDuration(labels prometheus.Labels, count float64)
}

// Service represents a release service. The release service
// does a release with a version
// corresponding to the timestamp from the commit push event
// so that it represents the same release/version as
// the other packaged languages. This is most useful for go
// as it uses git as its repository/package format.
type Service struct {
	fs      fs
	repo    repo
	logger  log.FieldLogger
	tracer  opentracing.Tracer
	metrics counters
}

type processorProps struct {
	ID      string
	WorkDir string
}

// New returns a pointer to a release Service configured with the parameters passed in.
func New(fs fs, repo repo, logger log.FieldLogger, metrics counters, tracer opentracing.Tracer) *Service {
	return &Service{
		fs,
		repo,
		logger,
		tracer,
		metrics,
	}
}

// Process is the main method for use by the main function of the application, and the only one required
// by the interface in main.go. It takes a context, used for cancelling itself in the case of a sigterm or sigint,
// and a Push Event payload. It executes all steps necessary for creating a release on the repo passed to the service.
func (s *Service) Process(ctx context.Context, payload hooks.PushPayload) {
	start := time.Now()
	// this span is a child of the parent span in the http handler, but since this will finish after
	// the http handler returns, it follows from that span so it will display correctly.
	parentContext := opentracing.SpanFromContext(ctx).Context()
	spanOption := opentracing.FollowsFrom(parentContext)
	span := opentracing.StartSpan("process_scala", spanOption)
	defer span.Finish()

	// creates a new copy of the context with the following span
	ctx = opentracing.ContextWithSpan(ctx, span)

	// create a new directory to do all the work of this Process call which can be cleaned up at the end.
	id := uuid.NewV4().String()
	workDir := fmt.Sprintf("/tmp/%s", id)
	err := os.Mkdir(workDir, 0750)
	if err != nil {
		s.metrics.AddPackagingErrors(prometheus.Labels{"type": "mkdir"}, 1)
		err = errors.WithStack(err)
		s.logger.Errorf("%+v", err)
		return
	}

	// create a struct for passing to other functions referencing the location of the work
	// this specific Process call is executing.
	procProps := processorProps{
		ID:      id,
		WorkDir: workDir,
	}

	// defer cleanup of this Process execution
	defer cleanup(ctx, s.fs, s.logger, procProps)

	// if we receive a signal that this goroutine should stop, do that
	// since cleanup is deferred it will still execute after the return statement
	select {
	case <-ctx.Done():
		return
	// otherwise, do our work
	default:
		// ignore tags, as we're trying to push them, so otherwise
		// we get into a loop
		if strings.Contains(payload.Ref, "tags") {
			return
		}

		// clone down the repository
		path, err := s.cloneCode(ctx, payload, procProps)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "clone"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
			return
		}

		// on master branch we want to cut a full release
		// but on any other branch or commit we should be be making
		// a prerelease
		var version string
		var prerelease bool
		if strings.Contains(payload.Ref, "master") {
			version = fmt.Sprintf("v1.0.%d", payload.Repository.PushedAt)
			prerelease = false
		} else {
			branch := g.RefEndName(payload.Ref)
			version = fmt.Sprintf("v1.0.%d-beta.%s", payload.Repository.PushedAt, branch)
			prerelease = true
		}

		if err = s.repo.CreateTag(path, version, "Automated tag by Protofact."); err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "release"}, 1)
			s.logger.Errorf("%+v\n", err)
			return
		}

		if err = s.repo.PushTags(path); err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "release"}, 1)
			s.logger.Errorf("%+v\n", err)
			return
		}

		bodyMsg := "Automated release by Protofact."
		rel := github.RepositoryRelease{
			TagName:    &version,
			Name:       &version,
			Body:       &bodyMsg,
			Prerelease: &prerelease,
		}

		_, err = s.repo.CreateRelease(ctx, payload.Repository.Owner.Login, payload.Repository.Name, &rel)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "release"}, 1)
			s.logger.Errorf("%+v\n", err)
			return
		}

		duration := time.Since(start)
		s.metrics.AddPackagingProcessDuration(prometheus.Labels{}, duration.Seconds())

		return
	}
}

// Cleanup runs a fs.DeleteDir on the build directory created when running Process.
func cleanup(ctx context.Context, fs fs, logger log.FieldLogger, props processorProps) {
	span, _ := opentracing.StartSpanFromContext(ctx, "cleanup")
	defer span.Finish()

	err := fs.DeleteDir(props.WorkDir)
	if err != nil {
		logger.Errorf("%+v\n", err)
	}
}

func (s *Service) cloneCode(ctx context.Context, payload hooks.PushPayload, props processorProps) (string, error) {
	// create tmp dir inside of the parent build directory so it gets cleaned up at the end
	cloneDir, err := s.fs.CreateUniqueTmpDir(props.WorkDir)
	if err != nil {
		return "", errors.Wrap(err, "could not create tmp directory for cloning")
	}
	// git clone the project
	err = s.repo.CloneWithCheckout(cloneDir, payload)
	if err != nil {
		return "", errors.Wrap(err, "could not clone directory and checkout branch")
	}

	return cloneDir, nil
}
