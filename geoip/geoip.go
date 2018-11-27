package geoip

import (
	"log"

	"github.com/anyaddres/supermann/config"

	"github.com/oschwald/maxminddb-golang"
)

// GeoIP Struct encompassing the maxmind db reader
type GeoIP struct {
	GDB *maxminddb.Reader
}

var gip *GeoIP

// NewGeoIP Singleton Pattern for the GeoIP Handle
func NewGeoIP(cfg *config.Config) *GeoIP {
	if gip != nil {
		return gip
	}
	database, err := maxminddb.Open(cfg.GeoIPDB)
	if err != nil {
		log.Fatalf("GeoIP File not present %s", cfg.GeoIPDB)
	}
	gip = &GeoIP{GDB: database}
	return gip
}

// CloseGeoIPHandle ...
func (geoip *GeoIP) CloseGeoIPHandle() {
	geoip.GDB.Close()
}
