package app

import (
	"net/url"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //go dialect has to be imported
	"ledger.api/pkg/logging"
)

// OpenGormConnection - opens connection for given url
func OpenGormConnection(rawDBurl string, logger logging.Logger) *gorm.DB {
	dbURL, err := url.Parse(rawDBurl)
	if err != nil {
		panic(err)
	}
	logger.
		WithField("host", dbURL.Host).
		WithField("db", dbURL.Path).
		Info("Initializing DB connection")
	db, err := gorm.Open("postgres", rawDBurl)
	if err != nil {
		panic(err)
	}
	db.SetLogger(&dbLogger{logger: logger})
	return db
}

type dbLogger struct {
	logger logging.Logger
}

func (dl dbLogger) Print(values ...interface{}) {
	dl.logger.Debug(values...)
}
