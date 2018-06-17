package main

import (
	"fmt"
	"net/http"

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
	return server.CreateAuthMiddlewareFunc(validator)
}

func main() {
	cfg := app.GetConfig()
	env := cfg.GetString("APP_ENV")
	logger := logging.NewLogger(env)
	db := app.OpenGormConnection(cfg.GetString("DB_URL"), logger)
	defer db.Close()

	handler := server.
		CreateHTTPApp(server.HTTPAppConfig{Env: env, Logger: logger}).
		Use(createAuthMiddleware(cfg)).
		RegisterRoutes(app.Routes).
		CreateHandler()

	port := cfg.GetInt("PORT")
	logger.Infof("Starting server on port: %v", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), handler); err != nil {
		logger.Error(err, "Failed to start server")
	}
}
