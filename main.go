package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	uuid "github.com/satori/go.uuid"
	logrus "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	jaeger "github.com/uber/jaeger-client-go/config"
	"gopkg.in/go-playground/webhooks.v5/github"

	"github.com/gospotcheck/protofact/pkg/config"
	"github.com/gospotcheck/protofact/pkg/filesys"
	"github.com/gospotcheck/protofact/pkg/git"
	"github.com/gospotcheck/protofact/pkg/metrics"
	"github.com/gospotcheck/protofact/pkg/services/release"
	"github.com/gospotcheck/protofact/pkg/services/ruby"
	"github.com/gospotcheck/protofact/pkg/services/scala"
	"github.com/gospotcheck/protofact/pkg/webhook"
)

var configFilePath string

func init() {
	flag.StringVarP(&configFilePath, "config", "c", "", "path to config file, default is none")
}

type languageProcessor interface {
	Process(ctx context.Context, payload github.PushPayload)
}

type parser interface {
	ValidateAndParsePushEvent(r *http.Request) (github.PushPayload, error)
}

func main() {
	conf, err := config.Read(configFilePath)
	if err != nil {
		errors.Wrap(err, "could not read in config:\n")
		log.Fatalf("%+v", err)
	}
	conf.Git.BaseAPIURL = "https://api.github.com"

	// setup logger

	// can add to this as other log level statements are added
	switch conf.LogLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.ErrorLevel)
	}

	// at the top level make sure the language is added to every log
	logger := logrus.WithFields(logrus.Fields{
		"language": conf.Language,
	})

	// set up a context that can be passed to all goroutines
	// with cancel so they can be cleaned up if a sig is received
	var ctx context.Context
	{
		var cancel context.CancelFunc
		ctx = context.Background()
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()

		// if we receive a sigint or sigterm, we call cancel,
		// which should cause all goroutines to return
		// then we kill the app
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigc
			cancel()
			fmt.Println("received a SIGINT or SIGTERM")
			logger.Fatal(sig)
		}()
	}

	// setup prometheus
	var counters *metrics.Counters
	{
		// only custom metric for now, but setup provides for more to be added as needed
		counters = &metrics.Counters{
			PackagingErrorCounter: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "error_total",
			}, []string{"language", "type", "app"}).MustCurryWith(prometheus.Labels{"language": conf.Language, "app": "proto-pkg"}),
			PackagingProcessDuration: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "bulk_process_duration_secs",
			}, []string{"language", "app"}).MustCurryWith(prometheus.Labels{"language": conf.Language, "app": "proto-pkg"}),
		}
	}
	http.Handle("/metrics", promhttp.Handler())

	// setup tracing
	{
		var tags []opentracing.Tag
		tags = append(tags, opentracing.Tag{
			Key:   "language",
			Value: conf.Language,
		})

		cfg, err := jaeger.FromEnv()
		if err != nil {
			err = errors.Wrap(err, "error reading in jaeger config")
			logger.Fatalf("%+v\n", err)
		}
		cfg.ServiceName = conf.Name
		cfg.Tags = tags

		tracer, closer, err := cfg.NewTracer()
		if err != nil {
			err = errors.Wrap(err, "error creating new tracer")
			logger.Fatalf("%+v\n", err)
		}
		defer closer.Close()

		opentracing.SetGlobalTracer(tracer)
	}

	// setup webhook parser
	var prsr parser
	{
		var err error
		prsr, err = webhook.NewParser(false, conf.Webhook)
		if err != nil {
			err = errors.Wrap(err, "error creating new parser")
			logger.Fatalf("%+v\n", err)
		}
	}

	// set up other service dependencies
	fs := &filesys.FS{}
	repo := git.New(conf.Git, logger)

	// based on language of container, setup the processor to use the correct service
	var svc languageProcessor
	{
		var err error
		switch conf.Language {
		case "scala":
			svc = scala.New(conf.Scala, fs, repo, logger, counters, opentracing.GlobalTracer())
		case "ruby":
			svc = ruby.New(conf.Ruby, fs, repo, logger, counters, opentracing.GlobalTracer())
		case "release":
			svc = release.New(repo, logger, counters, opentracing.GlobalTracer())
		default:
			err = errors.New("LANGUAGE configuration did not match any supported language")
			logger.Fatalf("%+v\n", err)
		}
	}

	// one route that receives all webhook requests
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		// start a span that can be added to the context for reference in the goroutine
		requestID := uuid.NewV4()
		span := opentracing.StartSpan("handle_webhook")
		span.SetTag("request_id", requestID)
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)

		// check push event
		payload, err := prsr.ValidateAndParsePushEvent(r)
		if err != nil {
			// if the request is bad log it and send it back
			// so Github can register the error
			w.WriteHeader(400)
			err = errors.Wrap(err, "error validating and parsing push event")
			logger.Errorf("%+v\n", err)
			return
		}
		// if the request is good set 200 header and send it back
		// Github may not wait as long as it takes to do this processing
		// so we want to handle failures in the app separately from
		// failures in receiving the event
		w.WriteHeader(200)

		// this is spun off as a cancelable goroutine
		// so it is not blocking on the response to Github
		go svc.Process(ctx, payload)

		return
	})

	// basic health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	port := fmt.Sprintf(":%s", conf.Port)
	log.Fatal(http.ListenAndServe(port, nil))
}
