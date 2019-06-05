package app

import (
	"fmt"
	"net/url"

	"ledger.api/pkg/core/diag"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //go dialect has to be imported
)

// OpenGormConnection - opens connection for given url
func OpenGormConnection(rawDBurl string) *gorm.DB {
	dbURL, err := url.Parse(rawDBurl)
	if err != nil {
		panic(err)
	}
	logger.WithData(diag.MsgData{
		"host": dbURL.Host,
		"db":   dbURL.Path,
	}).Info(nil, "Initializing DB connection")
	db, err := gorm.Open("postgres", rawDBurl)
	if err != nil {
		panic(err)
	}
	db.SetLogger(&dbLogger{logger: logger})
	return db
}

type dbLogger struct {
	logger diag.Logger
}

func (dl dbLogger) Print(values ...interface{}) {
	dl.logger.Debug(nil, fmt.Sprint(values...))
}
