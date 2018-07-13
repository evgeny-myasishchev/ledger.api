package ledgers

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"ledger.api/pkg/app"
	"ledger.api/pkg/logging"
)

var DB *gorm.DB

func TestMain(m *testing.M) {
	cfg := app.GetConfig()
	DB = app.OpenGormConnection(cfg.GetString("DB_URL"), logging.NewTestLogger()).LogMode(true)
	defer DB.Close()

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
