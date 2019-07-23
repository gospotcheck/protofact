package config

import (
	"fmt"
	"io/ioutil"

	"github.com/imdario/mergo"
	"github.com/jlevesy/envconfig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/gospotcheck/protofact/pkg/git"
	"github.com/gospotcheck/protofact/pkg/services/ruby"
	"github.com/gospotcheck/protofact/pkg/services/scala"
	"github.com/gospotcheck/protofact/pkg/webhook"
)

// Values represents the config values needed by the entire application.
// In addition to top level values like Language and LogLevel,
// it has nested config structs for each of the sub-packages like Git,
// and all the languages supported.
type Values struct {
	Git      git.Config
	Language string
	LogLevel string
	Name     string
	Port     string
	Ruby     ruby.Config
	Scala    scala.Config
	Webhook  webhook.Config
}

// Read will bring in config values from a YAML file at
// the passed filepath. It will then attempt to read config
// values from the environment. Any environment variables that
// are not zero-value will override any values from the YAML file.
func Read(configFilePath string) (*Values, error) {
	var yamlValues Values
	var envValues Values

	// if the configFilePath is populated, read from the file
	if configFilePath != "" {
		content, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			errMsg := fmt.Sprintf("could not find config file in provided path %s", configFilePath)
			return nil, errors.Wrap(err, errMsg)
		}
		err = yaml.Unmarshal(content, &yamlValues)
		if err != nil {
			return nil, errors.Wrap(err, "could not unmarshal config conteng into struct:\n")
		}
	}

	// now read from environment
	err := envconfig.New("PF", "_").Load(&envValues)
	if err != nil {
		return nil, errors.Wrap(err, "could not load config from env variables")
	}

	// and merge the environment onto the yaml, overriding with any
	// non-zero values
	err = mergo.Merge(&yamlValues, envValues, mergo.WithOverride)
	if err != nil {
		return nil, errors.Wrap(err, "could not merge env values onto yaml values:\n")
	}

	return &yamlValues, nil
}
