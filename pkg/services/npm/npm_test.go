package npm

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

func Test_CreatePackage(t *testing.T) {
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
		"language": "npm",
	})

	fs := &filesys.FS{}
	config := Config{
		PackageName:     "@demo/protos",
		Email:           "devs@dev.com",
		Publish:         false,
		RegistryURL:     "https://npmregistry.demo.com",
		ProtobufVersion: "3.10.0",
		ProjectURL:      "https://github.com/gospotcheck/protofact",
		Token:           "anauthtoken",
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
	err = createPackage(ctx, fs, config, logger, "./test-resources", version, payload, procProps)
	if err != nil {
		t.Errorf("%+v", err)
	}

	files, err := fs.GetFileNames(procProps.BuildDir)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, files, 2)
	assert.Contains(t, files, "package.json")
	assert.Contains(t, files, ".npmrc")

	subDirs, err := fs.GetSubDirectories(procProps.BuildDir)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, subDirs, 1)

	pkgFiles, err := fs.GetFileNames(fmt.Sprintf("%s/dist/idl/demo/token/v1", procProps.BuildDir))
	if err != nil {
		t.Errorf("error getting file names in project dir: %s\n", err)
	}
	assert.Len(t, pkgFiles, 8)
	assert.Contains(t, pkgFiles, "token_api_pb.js")
	assert.Contains(t, pkgFiles, "token_api_pb_service.js")
	assert.Contains(t, pkgFiles, "token_api_pb_service.d.ts")
	assert.Contains(t, pkgFiles, "token_api_pb.d.ts")
	assert.Contains(t, pkgFiles, "token_pb.js")
	assert.Contains(t, pkgFiles, "token_pb_service.js")
	assert.Contains(t, pkgFiles, "token_pb_service.d.ts")
	assert.Contains(t, pkgFiles, "token_pb.d.ts")

	err = publishPackage(ctx, config, logger, procProps.BuildDir, version)
	if err != nil {
		t.Errorf("npm publish failed at path %s: %s", procProps.BuildDir, err)
	}
}
