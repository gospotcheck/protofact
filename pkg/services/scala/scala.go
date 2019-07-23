// Package scala builds jars from Java code and uses sbt to publish them.
package scala

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/gobuffalo/packr/v2"
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

// Service represents a scala packaging service. It has properties
// that are the dependencies necessary for the service to function
// and receives methods allowing it to build the code directory
// necessary for publishing a jar. It fulfills the LanguageProcessor
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
	BuildNumber            int64
	Description            string
	GRPCPackagesVersion    string
	GogoProtoJavaVersion   string
	JarDir                 string
	MavenRepoPublishTarget string
	MavenRepoHost          string
	MavenRepoUser          string
	MavenRepoPassword      string
	Name                   string
	Organization           string
	Realm                  string
	SBTVersion             string
	SBTAssemblyVersion     string
	ScalaVersion           string
	ScalaTestVersion       string
	Snapshot               bool
}

type processorProps struct {
	ID       string
	BuildDir string
}

// New returns a pointer to a scala Service configured with the parameters passed in.
func New(config Config, fs fs, repo repo, logger log.FieldLogger, metrics counters, tracer opentracing.Tracer) *Service {
	ptURL, err := url.Parse(config.MavenRepoPublishTarget)
	if err != nil {
		err = errors.WithStack(err)
		logger.Fatal(err)
	}
	config.MavenRepoHost = ptURL.Host
	logger.Debug(config.MavenRepoHost)
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
	span := opentracing.StartSpan("process_scala", spanOption)
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

		// get all relevant subdirectories (java/com/*) and process them into their own directories to publish
		jarDir, err := createJar(ctx, s.fs, s.config, s.logger, path, payload, procProps)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "create"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
			return
		}

		// for each of those directories, publish the jar, either locally or to to a repo based on the config
		err = publishJar(ctx, s.config, s.logger, jarDir)
		if err != nil {
			s.metrics.AddPackagingErrors(prometheus.Labels{"type": "publish"}, 1)
			s.logger.Errorf("%+v\n", errors.WithStack(err))
		}

		duration := time.Since(start)
		s.metrics.AddPackagingProcessDuration(prometheus.Labels{}, duration.Seconds())

		return
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

// CreateJar takes a path and finds all directories in the subpath of java/com in that path. We package at that level.
// For those directories it processes the templates in the scala package to create a directory mirroring the structure
// of a publishable jar.
func createJar(ctx context.Context, fs fs, config Config, logger log.FieldLogger, codePath string, payload github.PushPayload, props processorProps) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "create_jars")
	span.SetTag("directory", codePath)
	// subdirectories of the language
	javaSubDir := fmt.Sprintf("%s/java/com", codePath)
	subdirs, err := fs.GetSubDirectories(javaSubDir)
	if err != nil {
		return "", errors.Wrap(err, "could not get subdirectories in the clone dir")
	}

	snapshot := !strings.Contains(payload.Ref, "master")
	values := templateValues{
		Name:                   config.JarName,
		JarDir:                 ".",
		BuildNumber:            payload.Repository.PushedAt,
		Description:            config.Description,
		GRPCPackagesVersion:    config.GRPCPackagesVersion,
		MavenRepoPublishTarget: config.MavenRepoPublishTarget,
		MavenRepoHost:          config.MavenRepoHost,
		MavenRepoUser:          config.MavenRepoUser,
		MavenRepoPassword:      config.MavenRepoPassword,
		Organization:           config.Organization,
		Realm:                  config.Realm,
		SBTVersion:             config.SBTVersion,
		SBTAssemblyVersion:     config.SBTAssemblyVersion,
		ScalaVersion:           config.ScalaVersion,
		Snapshot:               snapshot,
	}

	logger.Debug(fmt.Sprintf("%+v", values))

	jarDir, err := fs.CreateUniqueTmpDir(props.BuildDir)
	if err != nil {
		return "", errors.Wrap(err, "could not create tmp dir for templating")
	}

	err = processTemplates(ctx, logger, jarDir, values)
	if err != nil {
		return "", errors.Wrap(err, "could not process templates")
	}

	for _, dir := range subdirs {
		err = os.MkdirAll(fmt.Sprintf("%s/com/%s", jarDir, dir), 0750)
		if err != nil {
			return "", errors.Wrap(err, "could not create com subdirectory in template directory")
		}

		// move files from git repo over
		err = cp.Copy(fmt.Sprintf("%s/java/com/%s", codePath, dir), fmt.Sprintf("%s/com/%s", jarDir, dir))
		if err != nil {
			return "", errors.Wrap(err, "could not copy over code files")
		}
	}

	return jarDir, nil
}

// PublishJar publishes the jar to the repository defined by the target project's files.
// If service.config.Publish is true, it will publish to a live online external repository.
// If Publish is false it will publish locally for development and testing purposes.
func publishJar(ctx context.Context, config Config, logger log.FieldLogger, path string) error {
	var action string
	if config.Publish {
		action = "publish"
	} else {
		action = "compile"
	}

	span, _ := opentracing.StartSpanFromContext(ctx, action)
	span.SetTag("directory", path)
	defer span.Finish()

	// run sbt command in the tmp dirs
	// #nosec
	cmd := exec.Command("sbt", action)
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	logger.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error running sbt %s", action))
	}
	return nil
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

// processTemplates processes all the templates and files in the template directory to the build directory.
func processTemplates(ctx context.Context, logger log.FieldLogger, jarDir string, values templateValues) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "process_templates")
	span.SetTag("directory", jarDir)
	defer span.Finish()

	templates := packr.New("scala", "./template")

	projectDirPath := fmt.Sprintf("%s/project", jarDir)
	err := os.Mkdir(projectDirPath, 0750)
	if err != nil {
		return errors.Wrap(err, "could not create sub dir for project files")
	}

	buildTemplate, err := templates.FindString("build.sbt")
	if err != nil {
		return errors.Wrap(err, "could not find build.sbt template")
	}
	buildOutPath := fmt.Sprintf("%s/build.sbt", jarDir)
	err = processFileTemplate(logger, "build.sbt", buildTemplate, buildOutPath, values)
	if err != nil {
		return errors.Wrap(err, "could not process build.sbt template")
	}

	versionTemplate, err := templates.FindString("version.sbt")
	if err != nil {
		return errors.Wrap(err, "could not find version.sbt template")
	}
	versionOutPath := fmt.Sprintf("%s/version.sbt", jarDir)
	err = processFileTemplate(logger, "version.sbt", versionTemplate, versionOutPath, values)
	if err != nil {
		return errors.Wrap(err, "could not process version.sbt template")
	}

	buildPropsTemplate, err := templates.FindString("project/build.properties")
	if err != nil {
		return errors.Wrap(err, "could not find build.properties template")
	}
	buildPropsOutPath := fmt.Sprintf("%s/project/build.properties", jarDir)
	err = processFileTemplate(logger, "build.properties", buildPropsTemplate, buildPropsOutPath, values)
	if err != nil {
		return errors.Wrap(err, "could not copy build.properties template")
	}

	pluginsTemplate, err := templates.FindString("project/plugins.sbt")
	if err != nil {
		return errors.Wrap(err, "could not find plugins.sbt template")
	}
	pluginsOutPath := fmt.Sprintf("%s/project/plugins.sbt", jarDir)
	err = processFileTemplate(logger, "plugins.sbt", pluginsTemplate, pluginsOutPath, values)
	if err != nil {
		return errors.Wrap(err, "could not copy plugins.sbt template")
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
