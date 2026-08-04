package python

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	g "github.com/gogits/git-module"
	"github.com/opentracing/opentracing-go"
	cp "github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

type fs interface {
	CreateUniqueTmpDir(parentPath string) (string, error)
	DeleteDir(string) error
	GetFileNames(baseDir string) ([]string, error)
}

type repo interface {
	CloneWithCheckout(tmpDir string, payload github.PushPayload) error
}

type counters interface {
	AddPackagingErrors(labels prometheus.Labels, count float64)
	AddPackagingProcessDuration(labels prometheus.Labels, count float64)
}

// Service represents a python packaging service. Unlike ruby/npm, it does not
// template a package manifest from scratch - the source repo already ships a
// complete pyproject.toml per SourceDir (name, dependencies, namespace package
// config), so this service only patches the version in place, builds an sdist
// and wheel, and optionally publishes them via twine. It fulfills the
// LanguageProcessor interface in package main.
type Service struct {
	fs      fs
	repo    repo
	logger  log.FieldLogger
	tracer  opentracing.Tracer
	metrics counters
	config  Config
}

type processorProps struct {
	ID       string
	BuildDir string
}

// New returns a pointer to a python Service configured with the parameters passed in.
func New(config Config, fs fs, repo repo, logger log.FieldLogger, metrics counters, tracer opentracing.Tracer) *Service {
	return &Service{
		fs,
		repo,
		logger,
		tracer,
		metrics,
		config,
	}
}

// Process is the main method for use by the main function of the application, and the only one required
// by the interface in main.go. It takes a context, used for cancelling itself in the case of a sigterm or sigint,
// and a Push Event payload. It executes all steps necessary for building a python distribution and publishing it via twine.
func (s *Service) Process(ctx context.Context, payload github.PushPayload) {
	start := time.Now()
	// this span is a child of the parent span in the http handler, but since this will finish after
	// the http handler returns, it follows from that span so it will display correctly.
	parentContext := opentracing.SpanFromContext(ctx).Context()
	spanOption := opentracing.FollowsFrom(parentContext)
	span := opentracing.StartSpan("process_python", spanOption)
	defer span.Finish()

	// creates a new copy of the context with the following span
	ctx = opentracing.ContextWithSpan(ctx, span)

	// create a new directory to do all the work of this Process call which can be cleaned up at the end.
	id := uuid.NewV4().String()
	buildDir := fmt.Sprintf("/tmp/%s", id)
	err := os.Mkdir(buildDir, 0750)
	if err != nil {
		s.metrics.AddPackagingErrors(prometheus.Labels{"type": "mkdir"}, 1)
		err = errors.WithStack(err)
		s.logger.Errorf("%+v", err)
		return
	}

	// create a struct for passing to other functions referencing the location of the work
	// this specific Process call is executing.
	procProps := processorProps{
		ID:       id,
		BuildDir: buildDir,
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
		// clone down the repository
		path, err := s.cloneCode(ctx, payload, procProps)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "clone"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
			return
		}

		var version string
		isMaster := !strings.Contains(payload.Ref, "master")
		a := strconv.Itoa(int(payload.Repository.PushedAt))
		if !isMaster {
			version = fmt.Sprintf("1.0.%s", a)
		} else {
			branchName := g.RefEndName(payload.Ref)
			// PEP 440 local version labels only allow alphanumerics and periods as delimiters
			branchName = strings.Replace(branchName, "/", ".", -1)
			branchName = strings.Replace(branchName, "-", ".", -1)
			branchName = strings.Replace(branchName, "_", ".", -1)
			version = fmt.Sprintf("1.0.%s+%s", a, branchName)
		}

		// copy the configured source directory (which already contains a pyproject.toml) into
		// its own directory to package, patching in the real version.
		dir, err := createPackage(ctx, s.fs, s.config, s.logger, path, version, procProps)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "create"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
			return
		}

		// build the distribution, and publish it if configured to do so
		err = publishPackage(ctx, s.fs, s.config, s.logger, dir)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "publish"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
		}

		duration := time.Since(start)
		s.metrics.AddPackagingProcessDuration(prometheus.Labels{}, duration.Seconds())

		return
	}
}

// cleanup runs a fs.DeleteDir on the build directory created when running Process.
func cleanup(ctx context.Context, fs fs, logger log.FieldLogger, props processorProps) {
	span, _ := opentracing.StartSpanFromContext(ctx, "cleanup")
	defer span.Finish()

	err := fs.DeleteDir(props.BuildDir)
	if err != nil {
		logger.Errorf("%+v\n", err)
	}
}

func (s *Service) cloneCode(ctx context.Context, payload github.PushPayload, props processorProps) (string, error) {
	// create tmp dir inside of the parent build directory so it gets cleaned up at the end
	cloneDir, err := s.fs.CreateUniqueTmpDir(props.BuildDir)
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

var versionLineRegexp = regexp.MustCompile(`(?m)^version\s*=\s*".*"$`)

// createPackage copies config.SourceDir (e.g. "python" or "python-protobuf4") out of the cloned
// repo into its own build directory, then patches the version already declared in its
// pyproject.toml so the built distribution carries the version corresponding to this push event.
func createPackage(ctx context.Context, fs fs, config Config, logger log.FieldLogger, codePath, version string, props processorProps) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "create_python_package")
	span.SetTag("directory", codePath)
	defer span.Finish()

	sourceDir := fmt.Sprintf("%s/%s", codePath, config.SourceDir)

	packageDirParent, err := fs.CreateUniqueTmpDir(props.BuildDir)
	if err != nil {
		return "", errors.Wrap(err, "could not create tmp dir for packaging")
	}
	packageDir := fmt.Sprintf("%s/%s", packageDirParent, config.SourceDir)

	err = cp.Copy(sourceDir, packageDir)
	if err != nil {
		return "", errors.Wrap(err, "could not copy over code files")
	}

	logger.Debug(fmt.Sprintf("setting version %s in %s", version, packageDir))

	if err = setPackageVersion(packageDir, version); err != nil {
		return "", errors.Wrap(err, "could not set package version in pyproject.toml")
	}

	return packageDir, nil
}

// setPackageVersion rewrites the `version = "..."` line of the pyproject.toml at the root of
// packageDir. It leaves everything else (name, dependencies, namespace package config) untouched
// since proto-gen-v3 already declares the correct values for the variant being packaged.
func setPackageVersion(packageDir, version string) error {
	pyprojectPath := fmt.Sprintf("%s/pyproject.toml", packageDir)

	content, err := ioutil.ReadFile(pyprojectPath)
	if err != nil {
		return errors.Wrap(err, "could not read pyproject.toml")
	}

	replacement := fmt.Sprintf(`version = "%s"`, version)
	updated := versionLineRegexp.ReplaceAll(content, []byte(replacement))

	if err = ioutil.WriteFile(pyprojectPath, updated, 0750); err != nil {
		return errors.Wrap(err, "could not write updated pyproject.toml")
	}

	return nil
}

// publishPackage builds an sdist and wheel from packageDir via `python -m build`.
// If config.Publish is true, it also uploads the built artifacts to config.RegistryURL via twine.
func publishPackage(ctx context.Context, fs fs, config Config, logger log.FieldLogger, packageDir string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "build_python_package")
	span.SetTag("directory", packageDir)

	distDir := fmt.Sprintf("%s/dist", packageDir)

	buildCmd := exec.Command("python3", "-m", "build", "--sdist", "--wheel", "--outdir", distDir, packageDir)
	out, err := buildCmd.CombinedOutput()
	logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		errMessage := fmt.Sprintf("error running python -m build: %s\n", out)
		return errors.Wrap(err, errMessage)
	}

	span.Finish()

	if config.Publish {
		span, _ := opentracing.StartSpanFromContext(ctx, "publish_python_package")
		span.SetTag("directory", distDir)

		artifacts, err := fs.GetFileNames(distDir)
		if err != nil {
			return errors.Wrap(err, "could not list built artifacts")
		}

		args := []string{"upload", "--repository-url", config.RegistryURL}
		for _, artifact := range artifacts {
			args = append(args, fmt.Sprintf("%s/%s", distDir, artifact))
		}

		publishCmd := exec.Command("twine", args...)
		publishCmd.Env = append(
			os.Environ(),
			fmt.Sprintf("TWINE_USERNAME=%s", config.RegistryUser),
			fmt.Sprintf("TWINE_PASSWORD=%s", config.RegistryPass),
		)
		out, err := publishCmd.CombinedOutput()
		logger.Debug(fmt.Sprintf("%s", out))
		if err != nil {
			errMessage := fmt.Sprintf("error running twine upload: %s\n", out)
			return errors.Wrap(err, errMessage)
		}

		span.Finish()
	}

	return nil
}
