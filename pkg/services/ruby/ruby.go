package ruby

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gobuffalo/packr/v2"
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
	GetSubDirectories(string) ([]string, error)
	CopyFile(src, dest string) error
}

type repo interface {
	CloneWithCheckout(tmpDir string, payload github.PushPayload) error
}

type counters interface {
	AddPackagingErrors(labels prometheus.Labels, count float64)
	AddPackagingProcessDuration(labels prometheus.Labels, count float64)
}

// Service represents a ruby packaging service. It has properties
// that are the dependencies necessary for the service to function
// and receives methods allowing it to build the code directory
// necessary for publishing a gem. It fulfills the LanguageProcessor
// interface in package main.
type Service struct {
	fs      fs
	repo    repo
	logger  log.FieldLogger
	tracer  opentracing.Tracer
	metrics counters
	config  Config
}

type templateValues struct {
	Authors     string
	Email       string
	GemName     string
	GemRepoHost string
	GRPCVersion string
	Homepage    string
	Version     string
}

type processorProps struct {
	ID       string
	BuildDir string
}

// New returns a pointer to a ruby Service configured with the parameters passed in.
func New(config Config, fs fs, repo repo, logger log.FieldLogger, metrics counters, tracer opentracing.Tracer) *Service {
	// set up gem config file
	if config.Publish {
		if err := getGemCredentials(config.GemRepoUser, config.GemRepoPass, config.GemRepoHost); err != nil {
			err = errors.Wrap(err, "could not get gem credentials:\n")
			logger.Fatalf("%+v", err)
		}
	}
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
// and a Push Event payload. It executes all steps necessary for creating jars and publishing them via sbt.
func (s *Service) Process(ctx context.Context, payload github.PushPayload) {
	start := time.Now()
	// this span is a child of the parent span in the http handler, but since this will finish after
	// the http handler returns, it follows from that span so it will display correctly.
	parentContext := opentracing.SpanFromContext(ctx).Context()
	spanOption := opentracing.FollowsFrom(parentContext)
	span := opentracing.StartSpan("process_ruby", spanOption)
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
			// ruby gem version only allow periods as delimiters
			branchName = strings.Replace(branchName, "/", ".", -1)
			branchName = strings.Replace(branchName, "-", ".", -1)
			branchName = strings.Replace(branchName, "_", ".", -1)
			version = fmt.Sprintf("1.0.%s.pre.%s", a, branchName)
		}

		// get all relevant subdirectories (ruby/*) and process them into their own directories to publish
		dir, err := createGem(ctx, s.fs, s.config, s.logger, path, version, payload, procProps)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "create"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
			return
		}

		// publish the gem, either locally or to to a repo based on the config
		err = publishGem(ctx, s.config, s.logger, dir, version)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "publish"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
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

// CreateJars takes a path and finds all directories in the subpath of "ruby" in that path. We package at that level.
// For those directories it processes the templates in the ruby package to create a directory mirroring the structure
// of a publishable jar.
func createGem(ctx context.Context, fs fs, config Config, logger log.FieldLogger, codePath, version string, payload github.PushPayload, props processorProps) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "create_gem")
	span.SetTag("directory", codePath)
	// subdirectories of the language
	gemSubDir := fmt.Sprintf("%s/ruby", codePath)

	values := templateValues{
		Authors:     config.Authors,
		Email:       config.Email,
		GemName:     config.GemName,
		GemRepoHost: config.GemRepoHost,
		GRPCVersion: config.GRPCVersion,
		Homepage:    config.Homepage,
		Version:     version,
	}

	logger.Debug(fmt.Sprintf("%+v", values))

	gemDirParent, err := fs.CreateUniqueTmpDir(props.BuildDir)
	if err != nil {
		return "", errors.Wrap(err, "could not create tmp dir for templating")
	}
	gemDir := fmt.Sprintf("%s/%s", gemDirParent, config.GemName)
	libDir := fmt.Sprintf("%s/lib", gemDir)
	err = os.MkdirAll(libDir, 0750)
	if err != nil {
		return "", errors.Wrap(err, "could not make gem dir with 'lib'")
	}

	err = processTemplates(ctx, config, logger, gemDir, values)
	if err != nil {
		return "", errors.Wrap(err, "could not process templates")
	}

	// move files from git repo over
	err = cp.Copy(gemSubDir, libDir)
	if err != nil {
		return "", errors.Wrap(err, "could not copy over code files")
	}

	return gemDir, nil
}

// PublishGem publishes the gem to the repository defined on the machine.
// If service.config.Publish is true, it will publish to a live online external repository.
// If Publish is false it will just build the gem and not push it.
func publishGem(ctx context.Context, config Config, logger log.FieldLogger, path, version string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "buildGem")
	span.SetTag("directory", path)

	gemspecName := fmt.Sprintf("%s.gemspec", config.GemName)
	buildCmd := exec.Command("gem", "build", gemspecName)
	buildCmd.Dir = path
	out, err := buildCmd.CombinedOutput()
	logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		errMessage := fmt.Sprintf("error running gem build: %s\n", out)
		return errors.Wrap(err, errMessage)
	}

	span.Finish()

	if config.Publish {
		gemName := fmt.Sprintf("%s-%s.gem", config.GemName, version)

		// Push to the JFrog gem repo. Auth comes from the credentials file
		// written at startup by getGemCredentials.
		if err := pushGem(ctx, logger, path, gemName, config.GemRepoHost, nil); err != nil {
			return errors.Wrap(err, "could not push gem to gem repo host")
		}

		// During the migration we also push to Google Artifact Registry when a
		// GARHost is configured, so both JFrog and GAR stay in sync until all
		// consumers have moved over. GAR uses a short-lived OAuth token instead
		// of the credentials file, passed via GEM_HOST_API_KEY.
		if config.GARHost != "" {
			token, err := getGARToken(ctx)
			if err != nil {
				return errors.Wrap(err, "could not get GAR access token")
			}
			env := append(os.Environ(), fmt.Sprintf("GEM_HOST_API_KEY=Bearer %s", token))
			if err := pushGem(ctx, logger, path, gemName, config.GARHost, env); err != nil {
				return errors.Wrap(err, "could not push gem to GAR")
			}
		}
	}

	return nil
}

// pushGem runs `gem push` for the built gem against the given host. When env is
// non-nil it is used as the command environment (e.g. to supply GEM_HOST_API_KEY
// for token-based hosts like GAR); otherwise the process environment and the
// ~/.gem/credentials file are used.
func pushGem(ctx context.Context, logger log.FieldLogger, path, gemName, host string, env []string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "push_gem")
	span.SetTag("directory", path)
	span.SetTag("host", host)
	defer span.Finish()

	publishCmd := exec.Command("gem", "push", gemName, "--host", host)
	publishCmd.Dir = path
	if env != nil {
		publishCmd.Env = env
	}
	out, err := publishCmd.CombinedOutput()
	logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		errMessage := fmt.Sprintf("error running gem push to %s: %s\n", host, out)
		return errors.Wrap(err, errMessage)
	}

	return nil
}

// getGARToken fetches a short-lived OAuth access token for the workload's service
// account from the GCE/GKE metadata server. This relies on Workload Identity (or
// a service account attached to the node) rather than a static credential.
func getGARToken(ctx context.Context) (string, error) {
	const metadataURL = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"

	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "could not create metadata token request")
	}
	req = req.WithContext(ctx)
	req.Header.Set("Metadata-Flavor", "Google")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "could not fetch access token from metadata server")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read metadata token response")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("metadata server returned %d fetching access token: %s", resp.StatusCode, body)
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return "", errors.Wrap(err, "could not unmarshal metadata token response")
	}
	if token.AccessToken == "" {
		return "", errors.New("metadata server returned an empty access token")
	}

	return token.AccessToken, nil
}

// processTemplates processes all the templates and files in the template directory to the build directory.
func processTemplates(ctx context.Context, config Config, logger log.FieldLogger, gemDir string, values templateValues) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "process_templates")
	span.SetTag("directory", gemDir)
	defer span.Finish()

	templates := packr.New("ruby", "./template")

	buildTemplate, err := templates.FindString("Gemfile")
	if err != nil {
		return errors.Wrap(err, "could not find Gemfile template")
	}
	buildOutPath := fmt.Sprintf("%s/Gemfile", gemDir)
	err = processFileTemplate(logger, "Gemfile", buildTemplate, buildOutPath, values)
	if err != nil {
		return errors.Wrap(err, "could not process Gemfile template")
	}

	gemspecTemplate, err := templates.FindString("gem.gemspec")
	if err != nil {
		return errors.Wrap(err, "could not find gem.gemspec template")
	}
	gemspecOutPath := fmt.Sprintf("%s/%s.gemspec", gemDir, config.GemName)
	err = processFileTemplate(logger, "gem.gemspec", gemspecTemplate, gemspecOutPath, values)
	if err != nil {
		return errors.Wrap(err, "could not process gem.gemspec template")
	}

	gemTemplate, err := templates.FindString("gem.rb")
	if err != nil {
		return errors.Wrap(err, "could not find gem.rb template")
	}
	gemOutPath := fmt.Sprintf("%s/lib/%s.rb", gemDir, config.GemName)
	err = processFileTemplate(logger, "gem.rb", gemTemplate, gemOutPath, values)
	if err != nil {
		return errors.Wrap(err, "could not process gem.rb template")
	}

	return nil
}

func processFileTemplate(logger log.FieldLogger, filename, templateContent, outPath string, values templateValues) error {
	fileTemplate, err := template.New(filename).Parse(templateContent)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not parse %s at path %s into template", filename, templateContent))
	}

	var fileBuffer bytes.Buffer

	err = fileTemplate.Execute(&fileBuffer, values)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not execute %s template against values struct %+v", filename, values))
	}

	logger.Debug(string(fileBuffer.Bytes()))

	err = ioutil.WriteFile(outPath, fileBuffer.Bytes(), 0750)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not write %s to path %s", filename, outPath))
	}

	return nil
}

func getGemCredentials(user, pass, host string) error {
	client := &http.Client{}
	keyURL := fmt.Sprintf("%s/api/v1/api_key.yaml", host)
	req, err := http.NewRequest("GET", keyURL, nil)
	if err != nil {
		return errors.Wrap(err, "could not create new http request for api key url")
	}
	req.SetBasicAuth(user, pass)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not successfully get api key")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read response body")
	}
	err = ioutil.WriteFile("/root/.gem/credentials", body, 0600)
	if err != nil {
		return errors.Wrap(err, "could not write credentials to file")
	}
	return nil
}
