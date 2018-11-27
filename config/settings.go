package config

import "github.com/joeshaw/envdecode"

// Config ...
type Config struct {
	DatabaseFile string `env:"DATABASE_FILE,default=logins.db"`
	GeoIPDB      string `env:"GEO_IP_DB,default=/GeoLite2/GeoLite2-City.mmdb"`
}

// GetConfig ...
func GetConfig() *Config {
	cfg := &Config{}
	if err := envdecode.Decode(cfg); err != nil {
		return nil
	}
	return cfg
}
