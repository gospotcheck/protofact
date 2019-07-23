package webhook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNewSecureParser(t *testing.T) {
	c := Config{"IfWishesWereHorsesWedAllBeEatingSteak!"}
	_, err := NewParser(true, c)
	if err != nil {
		t.Errorf("could not create new parser: %s\n", err)
	}
}

func TestNewInsecureParser(t *testing.T) {
	c := Config{""}
	_, err := NewParser(false, c)
	if err != nil {
		t.Errorf("could not create new parser: %s\n", err)
	}
}

func TestNewSecureParserEmptySecret(t *testing.T) {
	c := Config{""}
	_, err := NewParser(true, c)
	if err == nil {
		t.Errorf("should have disallowed empty string secret but didn't")
	}
}

func TestPushEventParse(t *testing.T) {
	c := Config{""}
	p, err := NewParser(false, c)
	if err != nil {
		t.Errorf("could not create new parser: %s\n", err)
	}

	fileContent, err := ioutil.ReadFile("./testdata/push.json")
	if err != nil {
		t.Errorf("could not read test data file: %s\n", err)
	}
	reader := bytes.NewReader(fileContent)

	req, err := http.NewRequest("POST", "/webhook", reader)
	if err != nil {
		t.Errorf("error making new http request: %s\n", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Github-Event", "push")
	req.Header.Set("X-Hub-Signature", "sha1=156404ad5f721c53151147f3d3d302329f95a3ab")

	_, err = p.ValidateAndParsePushEvent(req)
	if err != nil {
		t.Errorf("error validating and parsing payload: %s\n", err)
	}
}

func TestNonPushEventParse(t *testing.T) {
	c := Config{""}
	p, err := NewParser(false, c)
	if err != nil {
		t.Errorf("could not create new parser: %s\n", err)
	}

	fileContent, err := ioutil.ReadFile("./testdata/pull-request.json")
	if err != nil {
		t.Errorf("could not read test data file: %s\n", err)
	}
	reader := bytes.NewReader(fileContent)

	req, err := http.NewRequest("POST", "/webhook", reader)
	if err != nil {
		t.Errorf("error making new http request: %s\n", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Github-Event", "pull_request")
	req.Header.Set("X-Hub-Signature", "sha1=35712c8d2bc197b7d07621dcf20d2fb44620508f")

	_, err = p.ValidateAndParsePushEvent(req)
	if err == nil {
		t.Error("should have created an error because of incorrect event type, but didn't")
	}
}

func TestOtherParseError(t *testing.T) {
	c := Config{""}
	p, err := NewParser(false, c)
	if err != nil {
		t.Errorf("could not create new parser: %s\n", err)
	}

	fileContent, err := ioutil.ReadFile("./testdata/push.json")
	if err != nil {
		t.Errorf("could not read test data file: %s\n", err)
	}
	reader := bytes.NewReader(fileContent)

	req, err := http.NewRequest("POST", "/webhook", reader)
	if err != nil {
		t.Errorf("error making new http request: %s\n", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature", "sha1=35712c8d2bc197b7d07621dcf20d2fb44620508f")

	_, err = p.ValidateAndParsePushEvent(req)
	if err == nil {
		t.Error("should have created an error because of missing headers, but didn't")
	}
}
