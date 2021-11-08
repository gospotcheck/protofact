package scala

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/gospotcheck/protofact/pkg/filesys"
	"github.com/gospotcheck/protofact/pkg/webhook"
)

func Test_CreateJars(t *testing.T) {
	p, err := webhook.NewParser(false, webhook.Config{Secret: ""})
	if err != nil {
		t.Errorf("could not create new parser: %+v\n", err)
	}

	fileContent, err := ioutil.ReadFile("../../webhook/testdata/push.json")
	if err != nil {
		t.Errorf("could not read test data file: %+v\n", err)
	}
	reader := bytes.NewReader(fileContent)

	req, err := http.NewRequest("POST", "/webhook", reader)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Github-Event", "push")
	req.Header.Set("X-Hub-Signature", "sha1=156404ad5f721c53151147f3d3d302329f95a3ab")

	payload, err := p.ValidateAndParsePushEvent(req)
	if err != nil {
		t.Errorf("%+v\n", err)
	}

	log.SetLevel(log.DebugLevel)

	logger := log.WithFields(log.Fields{
		"language": "scala",
	})

	fs := &filesys.FS{}
	config := Config{
		Description:                   "Proto-Gen files for Organization X",
		MavenRepoPublishTarget:        "https://repo1.maven.org/maven2",
		MavenRepoUser:                 "user",
		MavenRepoPassword:             "password",
		Organization:                  "Org X",
		Publish:                       false,
		Realm:                         "Artifactory",
		SBTVersion:                    "1.5.5",
		SBTProtocPluginPackageVersion: "0.99.33",
		ScalaVersion                   "2.12.10",
		LegacyScalaVersion             "2.11.12",
		ScalaPBRuntimePackageVersion:  "0.10.0-M4",
	}
	if err != nil {
		t.Error(err)
	}

	id := uuid.NewV4().String()
	buildDir := fmt.Sprintf("/tmp/%s", id)
	err = os.Mkdir(buildDir, 0750)
	if err != nil {
		err = errors.WithStack(err)
		t.Errorf("%+v\n", err)
		return
	}

	procProps := processorProps{
		ID:       id,
		BuildDir: buildDir,
	}

	ctx := context.Background()

	defer cleanup(ctx, fs, logger, procProps)
	path, err := createJar(ctx, fs, config, logger, "./test-resources", payload, procProps)
	if err != nil {
		t.Errorf("%+v\n", err)
	}

	files, err := fs.GetFileNames(path)
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	assert.Len(t, files, 3)
	assert.Contains(t, files, "build.sbt")
	assert.Contains(t, files, "version.sbt")

	subDirs, err := fs.GetSubDirectories(path)
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	assert.Len(t, subDirs, 2)
	projectFiles, err := fs.GetFileNames(fmt.Sprintf("%s/project", path))
	if err != nil {
		t.Errorf("error getting file names in project dir: %s\n", err)
	}
	assert.Len(t, projectFiles, 2)
	assert.Contains(t, projectFiles, "build.properties")
	assert.Contains(t, projectFiles, "plugins.sbt")
	projectDirs, err := fs.GetSubDirectories(path)
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	assert.Len(t, projectDirs, 2)
	if strings.Compare(projectDirs[0], "demo") == 0 {
		contents, err := ioutil.ReadFile(fmt.Sprintf("%s/build.sbt", path))
		if err != nil {
			t.Errorf("%+v\n", err)
		}
		stringContent := string(contents[:])
		if !strings.Contains(stringContent, "proto-gen-demo") {
			t.Error("template should have generated name with proto-gen-demo but did not\n")
		}

		codeDirs, err := fs.GetSubDirectories(fmt.Sprintf("%s/com/demo", path))
		if err != nil {
			t.Errorf("%+v\n", err)
		}
		assert.Len(t, codeDirs, 1)
		assert.Contains(t, codeDirs, "health")
		healthFiles, err := fs.GetFileNames(fmt.Sprintf("%s/com/demo/health", path))
		if err != nil {
			t.Errorf("%+v\n", err)
		}
		assert.Len(t, healthFiles, 4)
		assert.Contains(t, healthFiles, "Health.java")
		assert.Contains(t, healthFiles, "HealthOrBuilder.java")
		assert.Contains(t, healthFiles, "HealthProto.java")
		assert.Contains(t, healthFiles, "HealthStatus.java")
	}

	err = publishJar(ctx, config, logger, path)
	if err != nil {
		t.Errorf("jar publish failed at path %s: %s", path, err)
	}
}
