package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gospotcheck/protofact/pkg/git"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

type repo interface {
	CreateRelease(context.Context, git.CreateReleaseRequest, github.PushPayload) error
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
	repo    repo
	logger  log.FieldLogger
	tracer  opentracing.Tracer
	metrics counters
}

// New returns a pointer to a release Service configured with the parameters passed in.
func New(repo repo, logger log.FieldLogger, metrics counters, tracer opentracing.Tracer) *Service {
	return &Service{
		repo,
		logger,
		tracer,
		metrics,
	}
}

// Process is the main method for use by the main function of the application, and the only one required
// by the interface in main.go. It takes a context, used for cancelling itself in the case of a sigterm or sigint,
// and a Push Event payload. It executes all steps necessary for creating a release on the repo passed to the service.
func (s *Service) Process(ctx context.Context, payload github.PushPayload) {
	start := time.Now()
	// this span is a child of the parent span in the http handler, but since this will finish after
	// the http handler returns, it follows from that span so it will display correctly.
	parentContext := opentracing.SpanFromContext(ctx).Context()
	spanOption := opentracing.FollowsFrom(parentContext)
	span := opentracing.StartSpan("process_scala", spanOption)
	defer span.Finish()

	// creates a new copy of the context with the following span
	ctx = opentracing.ContextWithSpan(ctx, span)

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

		// on master branch we want to cut a full release
		// but on any other branch or commit we should be be making
		// a prerelease
		var version string
		var prerelease bool
		if strings.Contains(payload.Ref, "master") {
			version = fmt.Sprintf("v1.0.%d", payload.Repository.PushedAt)
			prerelease = false
		} else {
			version = fmt.Sprintf("v1.0.%d-beta", payload.Repository.PushedAt)
			prerelease = true
		}

		reqBody := git.CreateReleaseRequest{
			TagName:         version,
			TargetCommitish: payload.HeadCommit.ID,
			Name:            version,
			Body:            "Automated release by Protofact.",
			Prerelease:      prerelease,
			Draft:           false,
		}

		err := s.repo.CreateRelease(ctx, reqBody, payload)

		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "tag"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
			return
		}

		duration := time.Since(start)
		s.metrics.AddPackagingProcessDuration(prometheus.Labels{}, duration.Seconds())

		return
	}
}
