package main

import (
	"fmt"
	"net/http"

	"ledger.api/pkg/ledgers"
	"ledger.api/pkg/transactions"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"ledger.api/pkg/app"
	"ledger.api/pkg/auth"
	"ledger.api/pkg/logging"
	"ledger.api/pkg/server"
)

func createAuthMiddleware(cfg app.Config) server.RouterMiddlewareFunc {
	validator := auth.CreateAuth0Validator(
		cfg.GetString("AUTH0_ISS"),
		cfg.GetString("AUTH0_AUD"),
	)
	return server.CreateAuthMiddlewareFunc(server.AuthMiddlewareParams{
		Validator: validator,
		WhitelistedRoutes: map[string]bool{
			"/v2/healthcheck/ping": true,
		},
	})
}

func main() {
	cfg := app.GetConfig()
	env := cfg.GetString("APP_ENV")
	logger := logging.NewLogger(env)
	db := app.OpenGormConnection(cfg.GetString("DB_URL"), logger)
	defer db.Close()

	ledgersSvc := ledgers.CreateQueryService(db)
	transactonsQuerySvc := transactions.CreateQueryService(db)

	handler := server.
		CreateHTTPApp(server.HTTPAppConfig{Env: env, Logger: logger}).
		Use(server.CreateCorsMiddlewareFunc()).
		Use(createAuthMiddleware(cfg)).
		RegisterRoutes(app.Routes).
		RegisterRoutes(ledgers.CreateRoutes(ledgersSvc)).
		RegisterRoutes(transactions.CreateRoutes(transactonsQuerySvc)).
		CreateHandler()

	port := cfg.GetInt("PORT")
	logger.Infof("Starting server on port: %v", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), handler); err != nil {
		logger.Error(err, "Failed to start server")
	}
}
