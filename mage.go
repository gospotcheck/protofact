//+build mage

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/gospotcheck/protofact/pkg/git"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/google/go-github/v32/github"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"golang.org/x/crypto/ssh/terminal"
)

type key int

const (
	releaseID  key = iota
	tokenVal   key = iota
	versionVal key = iota
)

type releaseResponse struct {
	UploadURL string `json:"upload_url"`
}

// Build is a namespace for holding build commands
type Build mg.Namespace

func (Build) Compiler() error {
	if err := sh.Run(
		"docker", "build",
		"-t", "gospotcheck/protofact:compiler", ".",
		"-f", "./docker/compiler/Dockerfile",
	); err != nil {
		return err
	}
	return nil
}

// Linux builds a linux binary
func (Build) Linux() error {
	if err := sh.Run("docker-compose", "run", "compiler", "./docker/compiler/compile.sh"); err != nil {
		return err
	}

	return nil
}

// Release is a namespace for holding release commands
type Release mg.Namespace

// Create creates a new release on Github
func (Release) Create() error {
	ctx := context.Background()
	logger := logrus.WithFields(logrus.Fields{
		"executor": "mage",
	})

	fmt.Println("What is the release tag version?")
	var version string
	fmt.Scanln(&version)

	fmt.Println("Is this a prerelease?")
	var prereleaseStr string
	fmt.Scanln(&prereleaseStr)

	prerelease, err := strconv.ParseBool(prereleaseStr)
	if err != nil {
		err = errors.Wrap(err, "Answer must be a valid boolean.")
		fmt.Printf("%+v\n", err)
		return err
	}

	fmt.Println("Please enter a git user token:")
	byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}
	token := string(byteToken)

	fmt.Println("Creating release.")

	gitConfig := git.Config{Token: token}
	repo := git.New(ctx, gitConfig, logger)

	tagMsg := fmt.Sprintf("'Version %s'", version)
	if err = repo.CreateTag(".", version, tagMsg); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	if err = repo.PushTags("."); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	bodyMsg := "Release pushed by mage script"
	rel := &github.RepositoryRelease{
		TagName:    &version,
		Name:       &version,
		Body:       &bodyMsg,
		Prerelease: &prerelease,
	}

	rel, err = repo.CreateRelease(ctx, "gospotcheck", "protofact", rel)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	ctx = context.WithValue(ctx, releaseID, *rel.ID)
	ctx = context.WithValue(ctx, tokenVal, token)
	ctx = context.WithValue(ctx, versionVal, version)
	mg.CtxDeps(ctx, Release.UploadLinux)

	return nil
}

// UploadLinux tars and uploads the linux binary to the release
func (Release) UploadLinux(ctx context.Context) error {
	fmt.Println("Uploading linux tarfile.")
	if err := sh.Run("tar", "-czvf", "protofact_linux-amd64.tar.gz", "./protofact_linux-amd64"); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	id := ctx.Value(releaseID).(int64)
	token := ctx.Value(tokenVal).(string)

	file, err := os.Open("./protofact_linux-amd64.tar.gz")
	if err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	opts := &github.UploadOptions{
		Name:      "protofact_linux-amd64.tar.gz",
		MediaType: "octet-stream",
	}

	logger := logrus.WithFields(logrus.Fields{
		"executor": "mage",
	})
	gitConfig := git.Config{Token: token}
	repo := git.New(ctx, gitConfig, logger)

	if err = repo.UploadReleaseAsset(ctx, "gospotcheck", "protofact", id, opts, file); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	mg.CtxDeps(ctx, Release.BuildNPMContainer, Release.BuildReleaseContainer, Release.BuildRubyContainer, Release.BuildScalaContainer)

	return nil
}

func (Release) BuildReleaseContainer(ctx context.Context) error {
	fmt.Println("Building release container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:release-%s", version)
	buildArg := fmt.Sprintf("PROTOFACT_VERSION=%s", version)
	if err := sh.Run("docker", "build", "-t", tag, ".", "-f", "./docker/release/Dockerfile", "--build-arg", buildArg); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	mg.CtxDeps(ctx, Release.PublishReleaseContainer)

	return nil
}

func (Release) PublishReleaseContainer(ctx context.Context) error {
	fmt.Println("Publishing release container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:release-%s", version)
	if err := sh.Run("docker", "push", tag); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}
	return nil
}

func (Release) BuildNPMContainer(ctx context.Context) error {
	fmt.Println("Building npm container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:npm-%s", version)
	buildArg := fmt.Sprintf("PROTOFACT_VERSION=%s", version)
	if err := sh.Run("docker", "build", "-t", tag, ".", "-f", "./docker/npm/Dockerfile", "--build-arg", buildArg); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	mg.CtxDeps(ctx, Release.PublishNPMContainer)

	return nil
}

func (Release) PublishNPMContainer(ctx context.Context) error {
	fmt.Println("Publishing npm container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:npm-%s", version)
	if err := sh.Run("docker", "push", tag); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}
	return nil
}

func (Release) BuildRubyContainer(ctx context.Context) error {
	fmt.Println("Building ruby container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:ruby-%s", version)
	buildArg := fmt.Sprintf("PROTOFACT_VERSION=%s", version)
	if err := sh.Run("docker", "build", "-t", tag, ".", "-f", "./docker/ruby/Dockerfile", "--build-arg", buildArg); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	mg.CtxDeps(ctx, Release.PublishRubyContainer)

	return nil
}

func (Release) PublishRubyContainer(ctx context.Context) error {
	fmt.Println("Publishing ruby container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:ruby-%s", version)
	if err := sh.Run("docker", "push", tag); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}
	return nil
}

func (Release) BuildScalaContainer(ctx context.Context) error {
	fmt.Println("Building scala container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:scala-%s", version)
	buildArg := fmt.Sprintf("PROTOFACT_VERSION=%s", version)
	if err := sh.Run("docker", "build", "-t", tag, ".", "-f", "./docker/scala/Dockerfile", "--build-arg", buildArg); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}

	mg.CtxDeps(ctx, Release.PublishScalaContainer)

	return nil
}

func (Release) PublishScalaContainer(ctx context.Context) error {
	fmt.Println("Publishing scala container.")
	version := ctx.Value(versionVal).(string)
	tag := fmt.Sprintf("gospotcheck/protofact:scala-%s", version)
	if err := sh.Run("docker", "push", tag); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("%+v\n", err)
		return err
	}
	return nil
}
