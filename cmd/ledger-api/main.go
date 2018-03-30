package main

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"ledger.api/pkg/app"
	"ledger.api/pkg/server"
)

func main() {
	cfg := app.GetConfig()
	db := app.OpenGormConnection(cfg.GetString("DB_URL"))
	defer db.Close()

	// ledgersService := ledgers.CreateService(db)

	router := server.
		CreateDevRouter().
		RegisterRoutes(app.Routes)
	router.Run(cfg.GetInt("PORT"))
}
