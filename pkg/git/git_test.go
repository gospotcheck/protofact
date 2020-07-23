package git

import (
	"context"
	"encoding/json"
	"net/http"
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

	repo := New(Config{"auser", "apassword", "http://localhost:9000"}, logger)
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

	repo := New(Config{"user", "password", "http://localhost:9000"}, logger)
	url, err := repo.CreateAuthenticatedURL("https://github.com/org/repo")
	if err != nil {
		t.Error(err)
	}
	if strings.Compare(url, "https://user:password@github.com/org/repo") != 0 {
		t.Error("the function did not create a correcly formatted url")
	}
}

func TestCreateRelease(t *testing.T) {
	ctx := context.Background()
	ctx, done := context.WithCancel(ctx)
	go mockGithubServer(ctx)

	t.Run("Success", func(t *testing.T) {
		log.SetLevel(log.DebugLevel)
		logger := log.WithFields(log.Fields{
			"language": "release",
		})

		repo := New(Config{"user", "token", "http://localhost:9000"}, logger)

		var payload github.PushPayload
		fileContent, err := os.Open("./testdata/release-success.json")
		if err != nil {
			t.Errorf("could not open test data file: %s\n", err)
		}
		jsonParser := json.NewDecoder(fileContent)
		if err = jsonParser.Decode(&payload); err != nil {
			t.Error(err)
		}

		req := CreateReleaseRequest{
			TagName:         "v1.0.0",
			TargetCommitish: payload.HeadCommit.ID,
			Name:            "v1.0.0",
			Body:            "Automated release by Protofact.",
			Prerelease:      false,
			Draft:           false,
		}

		err = repo.CreateRelease(ctx, req, payload)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("BadRequest", func(t *testing.T) {
		log.SetLevel(log.DebugLevel)
		logger := log.WithFields(log.Fields{
			"language": "release",
		})

		repo := New(Config{"user", "token", "http://localhost:9000"}, logger)

		var payload github.PushPayload
		fileContent, err := os.Open("./testdata/release-badreq.json")
		if err != nil {
			t.Errorf("could not open test data file: %s\n", err)
		}
		jsonParser := json.NewDecoder(fileContent)
		if err = jsonParser.Decode(&payload); err != nil {
			t.Error(err)
		}

		req := CreateReleaseRequest{
			TagName:         "v1.0.0",
			TargetCommitish: payload.HeadCommit.ID,
			Name:            "v1.0.0",
			Body:            "Automated release by Protofact.",
			Prerelease:      false,
			Draft:           false,
		}

		err = repo.CreateRelease(ctx, req, payload)
		if err == nil {
			t.Error("expected an error from non 201 status code but got none")
		}
	})

	done()
}

func mockGithubServer(ctx context.Context) {
	// if we receive a signal that this goroutine should stop, do that
	// since cleanup is deferred it will still execute after the return statement
	select {
	case <-ctx.Done():
		return
	// otherwise, do our work
	default:
		http.HandleFunc("/repos/test/success/releases", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})
		http.HandleFunc("/repos/test/badreq/releases", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		})
		http.ListenAndServe(":9000", nil)
	}
}
