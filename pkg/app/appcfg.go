package app

import (
	"github.com/spf13/viper"
)

// Config - Application config interface
type Config interface {
	GetString(key string) string
	GetInt(key string) int
}

func setDefaults(cfg *viper.Viper) *viper.Viper {
	cfg.SetDefault("DB_URL", "host=localhost port=5432 user=postgres dbname=ledger-dev sslmode=disable")
	cfg.SetDefault("PORT", 3000)
	return cfg
}

// GetConfig - return config instance
func GetConfig() Config {
	return setDefaults(viper.New())
}
