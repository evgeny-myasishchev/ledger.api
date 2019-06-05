package config

import (
	"ledger.api/pkg/core/config"
	"ledger.api/version"
)

var appEnv = config.NewAppEnv(version.AppName)
var configBuilder = config.NewBuilder(appEnv)

var localParams = configBuilder.NewParamsBuilder(configBuilder.WithLocalSource())

// Do not change vars below at runtime
var (
	// ServerPort - server port of a service
	ServerPort = localParams.NewParam("server/port").Int()

	// LogMode - possible values: json or test
	// In test mode logs will be written to test.log file
	// IN json mode logs will go to stdout
	// TODO: Rename json mode to default
	LogMode = localParams.NewParam("log/mode").String()

	DbURL = localParams.NewParam("db/url").String()

	Auth0Aud = localParams.NewParam("auth0/aud").String()

	Auth0Iss = localParams.NewParam("auth0/iss").String()
)

// Load will load and initialize config
func Load() (config.ServiceConfig, error) {
	return configBuilder.LoadConfig()
}
