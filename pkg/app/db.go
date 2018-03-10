package app

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //go dialect has to be imported
)

// OpenGormConnection - opens connection for given url
func OpenGormConnection(dbURL string) *gorm.DB {
	db, err := gorm.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	return db
}
