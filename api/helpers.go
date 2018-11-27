package api

import (
	"log"
	"math"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"database/sql"
	"github.com/anyaddres/supermann/config"
	ds "github.com/anyaddres/supermann/datastore"
	"github.com/anyaddres/supermann/geoip"
	"github.com/umahmood/haversine"
)

// LocationCache represents the Ip to Location Map for faster retrieval.
type LocationCache struct {
	ipAddressToLocation map[string]*LoginInfo
	mutex               *sync.RWMutex
}

var locationCache = &LocationCache{ipAddressToLocation: make(map[string]*LoginInfo), mutex: &sync.RWMutex{}}

const (
	// SpeedThreshold ...
	SpeedThreshold = 500
)

func getLatLonForIP(ctx *SrvContext, entry *LoginRequest) (*LoginInfo, error) {
	locationCache.mutex.RLock()
	if rec, ok := locationCache.ipAddressToLocation[entry.IpAddress]; ok {
		locationCache.mutex.RUnlock()
		return rec, nil
	}
	locationCache.mutex.RUnlock()
	ip := net.ParseIP(entry.IpAddress)
	if ctx.gip == nil {
		ctx.gip = geoip.NewGeoIP(ctx.cfg)
	}
	err := ctx.gip.GDB.Lookup(ip, &city)
	if err != nil {
		return nil, err
	}
	loc := Location{Lat: city.Location.Latitude, Lon: city.Location.Longitude}
	rec := &LoginInfo{Location: loc, Radius: city.Location.AccuracyRadius}
	locationCache.mutex.Lock()
	locationCache.ipAddressToLocation[entry.IpAddress] = rec
	locationCache.mutex.Unlock()
	return rec, nil
}

func getDistanceBetweenLocations(l1, l2 Location) (mi, km float64) {
	loc1 := haversine.Coord{Lat: l1.Lat, Lon: l1.Lon}
	loc2 := haversine.Coord{Lat: l2.Lat, Lon: l2.Lon}
	return haversine.Distance(loc1, loc2)
}

func findMin(eventsGreaterThan []Events) *Events {
	// minStream chan<- *Events
	// defer close(minStream)
	var min Events
	if len(eventsGreaterThan) > 0 {
		min = eventsGreaterThan[0]
	}
	for index := 1; index < len(eventsGreaterThan); index++ {
		if eventsGreaterThan[index].TimeStamp < min.TimeStamp {
			min = eventsGreaterThan[index]
		}
	}
	return &min
}

func findMax(eventsLessThan []Events) *Events {
	var max Events
	// defer close(maxStream)
	if len(eventsLessThan) > 0 {
		max = eventsLessThan[0]
	}
	for index := 1; index < len(eventsLessThan); index++ {
		if eventsLessThan[index].TimeStamp > max.TimeStamp {
			max = eventsLessThan[index]
		}
	}
	// maxStream <- &max
	return &max
}

// Method computes the distance between the 2 Coordinates. It assumes that geoip mapping
// will work for all ip addresses. It does not handle a case where in given an IP address
// the geo db does not contain the lat and lon for that ip. In a real world scenario a
// hacker could spoof the originating Ip addresses such that they dont map to a lat/lon.
func isTravelSuspicious(entry *LoginRequest, latLonForReq *LoginInfo, prevsub *Events) (float64, bool) {
	miles, _ := getDistanceBetweenLocations(latLonForReq.Location, prevsub.LoginInfo.Location)
	ts1 := time.Unix(entry.UnixTimeStamp, 0)
	ts2 := time.Unix(prevsub.TimeStamp, 0)
	hours := math.Abs(ts1.Sub(ts2).Hours())
	speed := miles / hours
	if speed > SpeedThreshold {
		return speed, true
	}
	return speed, false
}

func subsequentEvent(db Searcher, entry *LoginRequest, latLonForReq *LoginInfo) (*Events, error) {
	var subsequent *Events
	largerSet, err := db.GetLoginsForUserGreaterThanOrLessThan(entry.UserName, ">", entry.UnixTimeStamp)
	// log.Printf("Got login entries having timestamp larger than current in %v", time.Since(start))
	if err != nil {
		return nil, err
	}
	larger := make([]Events, 0)
	for _, ele := range *largerSet {
		loc := Location{Lat: ele.Lat, Lon: ele.Lon}
		info := LoginInfo{Location: loc, Speed: ele.Speed, Radius: ele.Radius}
		larger = append(larger, Events{Ip: ele.IpAddress, TimeStamp: ele.UnixTimeStamp, LoginInfo: info})
	}
	if len(*largerSet) > 0 {
		// minStream := make(chan *Events, 1)
		// t2 := time.Now()
		subsequent = findMin(larger)
		// subsequent = <-minStream
		subsequent.Speed, subsequent.SuspiciousTravel = isTravelSuspicious(entry, latLonForReq, subsequent)
		return subsequent, nil
	}
	return nil, nil
}

func precedingEvent(db Searcher, entry *LoginRequest, latLonForReq *LoginInfo) (*Events, error) {
	var preceding *Events
	smallerSet, err := db.GetLoginsForUserGreaterThanOrLessThan(entry.UserName, "<", entry.UnixTimeStamp)
	if err != nil {
		return nil, err
	}
	smaller := make([]Events, 0)
	for _, ele := range *smallerSet {
		loc := Location{Lat: ele.Lat, Lon: ele.Lon}
		info := LoginInfo{Location: loc, Speed: ele.Speed, Radius: ele.Radius}
		smaller = append(smaller, Events{Ip: ele.IpAddress, TimeStamp: ele.UnixTimeStamp, LoginInfo: info})
	}
	if len(*smallerSet) > 0 {
		// maxStream := make(chan *Events, 1)
		// t1 := time.Now()
		preceding = findMax(smaller)
		// preceding = <-maxStream
		preceding.Speed, preceding.SuspiciousTravel = isTravelSuspicious(entry, latLonForReq, preceding)
		return preceding, nil
	}
	return nil, nil
}

// Method finds out closest previous login and closest subsequent login if they exist
// It divides the list of logins into 2 parts. Those having a timestamp greater than the current login
// and those having a timestamp less than the current event. It then finds the max timestamp
// among the less than events and min timestamp among the greater than events.
func closestNeighbouringLogins(db Searcher, entry *LoginRequest, latLonForReq *LoginInfo) (*Events, *Events, []error) {
	var wg sync.WaitGroup
	var errs []error
	var serr, perr error
	var subsequent, preceding *Events
	wg.Add(2)
	go func() {
		subsequent, serr = subsequentEvent(db, entry, latLonForReq)
		if serr != nil {
			log.Println(serr)
			errs = append(errs, serr)
		}
		wg.Done()
	}()

	go func() {
		preceding, perr = precedingEvent(db, entry, latLonForReq)
		if perr != nil {
			log.Println(perr)
			errs = append(errs, perr)
		}
		wg.Done()
	}()
	wg.Wait()
	if serr != nil || perr != nil {
		return nil, nil, errs
	}
	return preceding, subsequent, nil
}

// This persists
func persistLoginInfo(dao LoginStore, dp *LoginRequest, li *LoginInfo) error {
	start := time.Now()
	loginInfo := ds.LoginInfoDAO{Lat: li.Lat, Lon: li.Lon, Radius: li.Radius, Speed: li.Speed}
	loginDAO := &ds.LoginEntryDAO{
		LoginRequestDAO: ds.LoginRequestDAO(*dp),
		LoginInfoDAO:    loginInfo,
	}
	err := dao.InsertLogin(loginDAO)
	if err != nil {
		return err
	}
	log.Printf("persistLoginInfo took %s", time.Since(start))
	return nil
}

// NewServer ...
func NewServer() *Server {
	cfg := config.GetConfig()
	handlers := make(map[Route]func(*SrvContext, http.ResponseWriter, *http.Request) (interface{}, *apiErr), NumOfRoutes)
	srvContext := &SrvContext{
		cfg: cfg,
		db:  ds.NewDB(cfg.DatabaseFile),
	}
	// add routes
	handlers[IdentifyLogin] = identifySuspiciousLogins
	server := &Server{
		srvContext: srvContext,
		handle:     handlers,
	}
	runtime.GOMAXPROCS(MaxOsThreads)
	return server
}

// ServerCleanup ...
func (s *Server) ServerCleanup() {
	s.srvContext.db.CloseHandle()
	s.srvContext.gip.CloseGeoIPHandle()
}

// DbPing Initial DB Ping Check on Server Startup
func (s *Server) DbPing() {
	if s.srvContext.db == nil {
		log.Fatal("DB Handle has not been initialized")
	}
	_, err := sql.Open("sqlite3", "./"+s.srvContext.db.GetDBName())
	if err != nil {
		panic(err)
	}
}
