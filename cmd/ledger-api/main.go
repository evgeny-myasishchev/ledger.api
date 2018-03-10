package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"ledger.api/pkg/app"
	"ledger.api/pkg/ledgers"
)

func main() {
	cfg := app.GetConfig()
	db := app.OpenGormConnection(cfg.GetString("DB_URL"))
	defer db.Close()

	ledgersService := ledgers.CreateService(db)

	r := gin.Default()
	ledgers.RegisterRoutes(r, &ledgersService)
	app.RegisterRoutes(r)
	r.Run(fmt.Sprintf(":%v", cfg.GetInt("PORT")))
}
