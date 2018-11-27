package api

import (
	ds "github.com/anyaddres/supermann/datastore"
)

// LoginRequest represents the incoming data
type LoginRequest struct {
	UserName      string `json:"username,omitempty"`
	UnixTimeStamp int64  `json:"unix_timestamp,omitempty"`
	EventUUID     string `json:"event_uuid,omitempty"`
	IpAddress     string `json:"ip_address,omitempty"`
}

// Encloses the Lat Lon Info for a given IP and an error during geoip mapping
type LatLonResult struct {
	*LoginInfo
	Error error
}

// Location struct that consits of the Latitude and Longitude
type Location struct {
	Lat float64 `json:"lat,omitempty"`
	Lon float64 `json:"lon,omitempty"`
}

// LoginInfo struct that includes the Location and also the Accuracy radius and speed
type LoginInfo struct {
	Location
	Radius uint16  `json:"radius,omitempty"`
	Speed  float64 `json:"speed,omitempty"`
}

type LoginEntry struct {
	LoginRequest `json:",omitempty"`
	LoginInfo    `json:",omitempty"`
}

type Events struct {
	LoginInfo
	Ip               string `json:"ip,omitempty"`
	TimeStamp        int64  `json:"timestamp,omitempty"`
	SuspiciousTravel bool   `json:"suspiciousTravel"`
}

type Response struct {
	CurrentGeo         *LoginInfo `json:"currentGeo,omitempty"`
	PrecedingIpAccess  *Events    `json:"precedingIpAccess,omitempty"`
	SubsequentIpAccess *Events    `json:"subsequentIpAccess,omitempty"`
}

type LoginStore interface {
	InsertLogin(loginEntry *ds.LoginEntryDAO) error
	GetLoginsForUserGreaterThanOrLessThan(loginID string, operator string, ts int64) (*[]ds.LoginEntryDAO, error)
}

type Searcher interface {
	GetLoginsForUserGreaterThanOrLessThan(username, operator string, ts int64) (*[]ds.LoginEntryDAO, error)
}
