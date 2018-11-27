package datastore

// LoginEntryDAO represents the data that finally gets persisted into the DB
type LoginEntryDAO struct {
	LoginRequestDAO
	LoginInfoDAO
}

// LoginRequestDAO represents the data that comes from the user request
type LoginRequestDAO struct {
	UserName      string `db:"username" json:"username,string"`
	UnixTimeStamp int64  `db:"unix_timestamp" json:"unix_timestamp,int"`
	EventUUID     string `db:"event_uuid" json:"event_uuid,string"`
	IpAddress     string `db:"ip_address" json:"ip_address,string"`
}

// LoginInfoDAO represents the computed latitude, longitude, radius and speed.
type LoginInfoDAO struct {
	Lat    float64 `db:"lat" json:"lat,string"`
	Lon    float64 `db:"lon" json:"lon,string"`
	Radius uint16  `db:"radius" json:"radius,string" `
	Speed  float64 `db:"speed" json:"speed,string"`
}
