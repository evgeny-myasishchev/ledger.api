package main

import (
	"fmt"
	"net/http"

	"ledger.api/config"
	"ledger.api/pkg/core/diag"

	"ledger.api/pkg/ledgers"
	"ledger.api/pkg/transactions"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"ledger.api/pkg/app"
	"ledger.api/pkg/auth"
	"ledger.api/pkg/server"
)

var logger = diag.CreateLogger()

type auth0Cfg struct {
	iss string
	aud string
}

func createAuthMiddleware(cfg auth0Cfg) server.RouterMiddlewareFunc {
	validator := auth.CreateAuth0Validator(
		cfg.iss,
		cfg.aud,
	)
	return server.CreateAuthMiddlewareFunc(server.AuthMiddlewareParams{
		Validator: validator,
		WhitelistedRoutes: map[string]bool{
			"/v2/healthcheck/ping": true,
		},
	})
}

func main() {
	cfg := config.Load()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogMode(cfg.StringParam(config.LogMode).Value())
		setup.SetLogLevel(cfg.StringParam(config.LogLevel).Value())
	})

	db := app.OpenGormConnection(cfg.StringParam(config.DbURL).Value())
	defer db.Close()

	ledgersSvc := ledgers.CreateQueryService(db)
	transactonsQuerySvc := transactions.CreateQueryService(db)

	handler := server.
		CreateHTTPApp(server.HTTPAppConfig{Env: "dev"}).
		Use(server.CreateCorsMiddlewareFunc()).
		Use(createAuthMiddleware(auth0Cfg{
			iss: cfg.StringParam(config.Auth0Iss).Value(),
			aud: cfg.StringParam(config.Auth0Aud).Value(),
		})).
		RegisterRoutes(app.Routes).
		RegisterRoutes(ledgers.CreateRoutes(ledgersSvc)).
		RegisterRoutes(transactions.CreateRoutes(transactonsQuerySvc)).
		CreateHandler()

	port := cfg.IntParam(config.ServerPort).Value()
	logger.Info(nil, "Starting server on port: %v", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), handler); err != nil {
		logger.WithError(err).Error(nil, "Failed to start server")
	}
}
