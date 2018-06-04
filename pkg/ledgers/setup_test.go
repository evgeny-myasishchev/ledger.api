package ledgers

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"ledger.api/pkg/app"
	"ledger.api/pkg/logging"
)

var DB *gorm.DB
var service Service

func TestMain(m *testing.M) {
	cfg := app.GetConfig()
	DB = app.OpenGormConnection(cfg.GetString("DB_URL"), logging.NewTestLogger())
	defer DB.Close()
	service = CreateService(DB)
	ResetSchema(DB)

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
