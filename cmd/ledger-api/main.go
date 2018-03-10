/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
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
	r.Run() // listen and serve on 0.0.0.0:8080
}
