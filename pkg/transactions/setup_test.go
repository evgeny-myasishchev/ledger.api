package transactions

import (
	"os"
	"testing"

	"ledger.api/config"

	"github.com/jinzhu/gorm"
	"ledger.api/pkg/app"
)

var DB *gorm.DB

func TestMain(m *testing.M) {
	cfg := config.Load()
	DB = app.OpenGormConnection(cfg.StringParam(config.DbURL).Value())
	defer DB.Close()

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
