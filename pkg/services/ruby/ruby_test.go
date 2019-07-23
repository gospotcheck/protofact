package ruby

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/gospotcheck/protofact/pkg/filesys"
	"github.com/gospotcheck/protofact/pkg/webhook"
)

func Test_CreateGem(t *testing.T) {
	p, err := webhook.NewParser(false, webhook.Config{Secret: ""})
	if err != nil {
		t.Errorf("could not create new parser: %s\n", err)
	}

	fileContent, err := ioutil.ReadFile("../../webhook/testdata/push.json")
	if err != nil {
		t.Errorf("could not read test data file: %s\n", err)
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
		t.Error(err)
	}

	log.SetLevel(log.DebugLevel)

	logger := log.WithFields(log.Fields{
		"language": "ruby",
	})

	fs := &filesys.FS{}
	config := Config{
		Authors:     "Devs",
		Email:       "devs@dev.com",
		Publish:     false,
		GemName:     "protos-demo",
		GemRepoHost: "https://somerepo.gems.com",
		GRPCVersion: "1.19.0",
		Homepage:    "https://github.com/gospotcheck/protofact",
	}
	if err != nil {
		t.Error(err)
	}

	id := uuid.NewV4().String()
	buildDir := fmt.Sprintf("/tmp/%s", id)
	err = os.Mkdir(buildDir, 0750)
	if err != nil {
		err = errors.WithStack(err)
		t.Errorf("%+v", err)
		return
	}

	procProps := processorProps{
		ID:       id,
		BuildDir: buildDir,
	}

	ctx := context.Background()

	a := strconv.Itoa(int(payload.Repository.PushedAt))
	version := fmt.Sprintf("1.0.%s", a)

	defer cleanup(ctx, fs, logger, procProps)
	path, err := createGem(ctx, fs, config, logger, "./test-resources", version, payload, procProps)
	if err != nil {
		t.Errorf("%+v", err)
	}

	files, err := fs.GetFileNames(path)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, files, 2)
	assert.Contains(t, files, "Gemfile")
	assert.Contains(t, files, "protos-demo.gemspec")

	subDirs, err := fs.GetSubDirectories(path)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, subDirs, 1)

	libFiles, err := fs.GetFileNames(fmt.Sprintf("%s/lib", path))
	if err != nil {
		t.Errorf("error getting file names in lib dir")
	}
	assert.Len(t, libFiles, 1)
	assert.Contains(t, libFiles, "protos-demo.rb")

	rubyFiles, err := fs.GetFileNames(fmt.Sprintf("%s/lib/idl/demo/health", path))
	if err != nil {
		t.Errorf("error getting file names in project dir: %s\n", err)
	}
	assert.Len(t, rubyFiles, 3)
	assert.Contains(t, rubyFiles, "events_pb.rb")
	assert.Contains(t, rubyFiles, "health_pb.rb")
	assert.Contains(t, rubyFiles, "pagination_pb.rb")

	err = publishGem(ctx, config, logger, path, version)
	if err != nil {
		t.Errorf("gem publish failed at path %s: %s", path, err)
	}

	files, err = fs.GetFileNames(path)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, files, 3)
	gemName := fmt.Sprintf("protos-demo-%s.gem", version)
	assert.Contains(t, files, gemName)
}
