package python

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/gospotcheck/protofact/pkg/filesys"
	"github.com/gospotcheck/protofact/pkg/webhook"
)

func Test_CreateAndPublishPackage(t *testing.T) {
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
		"language": "python",
	})

	fs := &filesys.FS{}
	config := Config{
		Publish:      false,
		SourceDir:    "python",
		RegistryURL:  "https://test.pypi.example.com",
		RegistryUser: "user",
		RegistryPass: "pass",
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
	path, err := createPackage(ctx, fs, config, logger, "./test-resources", version, procProps)
	if err != nil {
		t.Errorf("%+v", err)
	}

	files, err := fs.GetFileNames(path)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, files, 1)
	assert.Contains(t, files, "pyproject.toml")

	pyprojectContent, err := ioutil.ReadFile(fmt.Sprintf("%s/pyproject.toml", path))
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, string(pyprojectContent), fmt.Sprintf(`version = "%s"`, version))

	subDirs, err := fs.GetSubDirectories(path)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, subDirs, 1)
	assert.Contains(t, subDirs, "idl")

	pyFiles, err := fs.GetFileNames(fmt.Sprintf("%s/idl/demo/health", path))
	if err != nil {
		t.Errorf("error getting file names in project dir: %s\n", err)
	}
	assert.Len(t, pyFiles, 3)
	assert.Contains(t, pyFiles, "events_pb2.py")
	assert.Contains(t, pyFiles, "health_pb2.py")
	assert.Contains(t, pyFiles, "pagination_pb2.py")

	err = publishPackage(ctx, fs, config, logger, path)
	if err != nil {
		t.Errorf("python build failed at path %s: %s", path, err)
	}

	distFiles, err := fs.GetFileNames(fmt.Sprintf("%s/dist", path))
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, distFiles, 2)
	var sawSdist, sawWheel bool
	for _, f := range distFiles {
		if strings.HasSuffix(f, ".tar.gz") {
			sawSdist = true
		}
		if strings.HasSuffix(f, ".whl") {
			sawWheel = true
		}
	}
	assert.True(t, sawSdist, "expected an sdist (.tar.gz) in dist/")
	assert.True(t, sawWheel, "expected a wheel (.whl) in dist/")
}
