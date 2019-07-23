package git

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"

	"github.com/gospotcheck/protofact/pkg/filesys"
)

func TestCloneWithCheckout(t *testing.T) {
	fs := &filesys.FS{}
	path, err := fs.CreateUniqueTmpDir("/tmp")
	if err != nil {
		t.Error(err)
	}

	var payload github.PushPayload
	fileContent, err := os.Open("./testdata/push.json")
	if err != nil {
		t.Errorf("could not open test data file: %s\n", err)
	}
	jsonParser := json.NewDecoder(fileContent)
	if err = jsonParser.Decode(&payload); err != nil {
		t.Error(err)
	}

	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"language": "scala",
	})

	repo := New(Config{"auser", "apassword"}, logger)
	err = repo.CloneWithCheckout(path, payload)
	if err != nil {
		t.Error(err)
	}

	err = fs.DeleteDir(path)
}

func TestCreateAuthenticatedURL(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"language": "scala",
	})

	repo := New(Config{"user", "password"}, logger)
	url, err := repo.CreateAuthenticatedURL("https://github.com/org/repo")
	if err != nil {
		t.Error(err)
	}
	if strings.Compare(url, "https://user:password@github.com/org/repo") != 0 {
		t.Error("the function did not create a correcly formatted url")
	}
}
