package app

import (
	"github.com/spf13/viper"
)

func setDefaults(cfg *viper.Viper) *viper.Viper {
	cfg.SetDefault("DB_URL", "host=localhost port=5432 user=postgres dbname=ledger-dev sslmode=disable")
	return cfg
}

// GetConfig - return config instance
func GetConfig() *viper.Viper {
	return setDefaults(viper.New())
}
