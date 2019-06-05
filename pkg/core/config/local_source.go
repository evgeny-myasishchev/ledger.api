package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type localSource struct {
	dir                  string
	configFiles          []string
	envOverrides         map[string]interface{}
	defaultService       string
	ignoreDefaultService bool
}

func (s *localSource) GetParameters(params []param) (map[param]interface{}, error) {
	values := map[param]interface{}{}

	pick := func(obj interface{}, path string) interface{} {
		parts := strings.Split(path, "/")
		paramVal := obj
		for _, part := range parts {
			var ok bool
			if paramVal, ok = paramVal.(map[string]interface{})[part]; !ok {
				paramVal = nil
				break
			}
		}
		return paramVal
	}

	paramPath := func(p param) string {
		if p.service() == "" {
			return p.key()
		}
		if s.ignoreDefaultService && p.service() == s.defaultService {
			return p.key()
		}
		return p.service() + "/" + p.key()
	}

	for _, configFile := range s.configFiles {
		buffer, err := ioutil.ReadFile(path.Join(s.dir, configFile))
		if err != nil {
			if configFile != "default.json" {
				continue
			}
			return nil, err
		}
		var configData map[string]interface{}
		if err := json.Unmarshal(buffer, &configData); err != nil {
			return nil, err
		}

		for _, param := range params {
			paramVal := pick(configData, paramPath(param))
			if paramVal != nil {
				values[param] = paramVal
			}
		}
	}

	if s.envOverrides != nil {
		for _, param := range params {
			envName := pick(s.envOverrides, paramPath(param))
			if envName == nil {
				continue
			}
			if envVal := os.Getenv(envName.(string)); envVal != "" {
				values[param] = envVal
			}
		}
	}

	return values, nil
}

// LocalOpt is an option of a local config source
type LocalOpt func(s *localSource)

// WithDir option to set local dir to load config from
func WithDir(dir string) LocalOpt {
	return func(s *localSource) {
		s.dir = dir
	}
}

// WithIgnoreDefaultService option to skip default service when building param path
// so params for the default service will be resolved from a root of a config
func WithIgnoreDefaultService() LocalOpt {
	return func(s *localSource) {
		s.ignoreDefaultService = true
	}
}

// WithAppEnv option will sent the app env
func WithAppEnv(appEnv AppEnv) LocalOpt {
	return func(s *localSource) {
		s.configFiles = append(s.configFiles, appEnv.Name+".json")
		s.defaultService = appEnv.ServiceName
		if appEnv.Facet != "" {
			s.configFiles = append(s.configFiles, appEnv.Name+"-"+appEnv.Facet+".json")
		}
	}
}

// NewLocalSource creates a source that reads params from a local fs.
// It is similar to node-config, suports json and custom-environment-variables.json
func NewLocalSource(opts ...LocalOpt) (Source, error) {
	// TODO: Default dir
	source := &localSource{
		configFiles: []string{"default.json"},
	}
	for _, opt := range opts {
		opt(source)
	}

	overridesFilePath := path.Join(source.dir, "custom-environment-variables.json")
	if overridesBuffer, err := ioutil.ReadFile(overridesFilePath); err == nil {
		envOverrides := map[string]interface{}{}
		if err := json.Unmarshal(overridesBuffer, &envOverrides); err != nil {
			return nil, err
		}
		source.envOverrides = envOverrides
	}

	return source, nil
}
