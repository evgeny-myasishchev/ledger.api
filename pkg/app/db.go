package app

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //go dialect has to be imported
	"ledger.api/pkg/logging"
)

// OpenGormConnection - opens connection for given url
func OpenGormConnection(dbURL string, logger logging.Logger) *gorm.DB {
	db, err := gorm.Open("postgres", dbURL)
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
