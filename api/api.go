package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/anyaddres/supermann/config"
	ds "github.com/anyaddres/supermann/datastore"
	"github.com/anyaddres/supermann/geoip"
	_ "github.com/mattn/go-sqlite3" // Registering the SQL Lite Driver
)

// Route ...
type Route string

const (
	// IdentifyLogin ...
	IdentifyLogin Route = "/api/identifylogins/"
	// NumOfRoutes ...
	NumOfRoutes = 1
	// MaxOsThreads ...
	MaxOsThreads = 100
)

type (
	// SrvContext ...
	SrvContext struct {
		cfg *config.Config
		db  *ds.DB
		gip *geoip.GeoIP
	}
	// Server ...
	Server struct {
		srvContext *SrvContext
		handle     map[Route]func(*SrvContext, http.ResponseWriter, *http.Request) (interface{}, *apiErr)
	}
)

var (
	city struct {
		Location struct {
			AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
			Latitude       float64 `maxminddb:"latitude"`
			Longitude      float64 `maxminddb:"longitude"`
			MetroCode      uint    `maxminddb:"metro_code"`
			TimeZone       string  `maxminddb:"time_zone"`
		} `maxminddb:"location"`
	}
)

// ServeHTTP...
func (s *Server) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	switch Route(req.RequestURI) {
	case IdentifyLogin:
		if req.Method != "POST" {
			json.NewEncoder(writer).Encode(newNotImplemented(req.Method, req.RequestURI))
			return
		}
		writer.Header().Set("Content-type", "application/json")
		// start := time.Now()
		apiResp, err := s.handle[IdentifyLogin](s.srvContext, writer, req)
		if err != nil {
			json.NewEncoder(writer).Encode(err)
			return
		}
		// log.Printf("***** identifySuspiciousLogins took %s\n", time.Since(start))
		json.NewEncoder(writer).Encode(apiResp)
	}
}

func identifySuspiciousLogins(ctx *SrvContext, w http.ResponseWriter, r *http.Request) (interface{}, *apiErr) {
	var loginEvent LoginRequest
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, newInternalServerErr(err)
	}
	err = json.Unmarshal(data, &loginEvent)
	if err != nil {
		return nil, newInternalServerErr(err)
	}
	// Input Validation
	if validationErrs := loginEvent.validate(); len(validationErrs) > 0 {
		return nil, newInvalidArgumentErr(validationErrs)
	}

	latLonForEntry, err := getLatLonForIP(ctx, &loginEvent)
	if err != nil {
		return nil, newInternalServerErr(err)

	}

	prev, next, errs := closestNeighbouringLogins(ctx.db, &loginEvent, latLonForEntry)
	if errs != nil {
		serr := ""
		for _, e := range errs {
			serr += e.Error() + "\n"
		}
		return nil, newInternalServerErr(fmt.Errorf(serr))
	}

	err = persistLoginInfo(ctx.db, &loginEvent, latLonForEntry)
	if err != nil {
		return nil, newInternalServerErr(err)
	}
	return &Response{CurrentGeo: latLonForEntry, PrecedingIpAccess: prev, SubsequentIpAccess: next}, nil
}
