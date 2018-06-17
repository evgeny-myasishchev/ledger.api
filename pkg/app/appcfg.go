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
	cfg.SetDefault("APP_ENV", "dev")
	cfg.SetDefault("DB_URL", "host=localhost port=5432 user=postgres dbname=ledger-dev sslmode=disable")
	cfg.SetDefault("PORT", 3000)
	cfg.SetDefault("AUTH0_AUD", "https://staging.api.my-ledger.com/")
	cfg.SetDefault("AUTH0_ISS", "https://ledger-staging.eu.auth0.com/")
	return cfg
}

// GetConfig - return config instance
func GetConfig() Config {
	return setDefaults(viper.New())
}
