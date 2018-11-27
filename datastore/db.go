package datastore

import (
	"database/sql"
	"log"
	"strconv"
)

// DB ...
type DB struct {
	dbh  *sql.DB
	name string
}

// Db ...
var db *DB

const (
	// JournalMode ...
	JournalMode = "PRAGMA journal_mode=WAL;"
	// Synchronous ...
	Synchronous = "PRAGMA synchronous=NORMAL;"
)

func setupDB(db *sql.DB) {
	_, err := db.Exec(JournalMode)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(Synchronous)
	if err != nil {
		log.Fatal(err)
	}

	createStmt := "CREATE TABLE IF NOT EXISTS logins (id INTEGER PRIMARY KEY, username TEXT, " +
		"unix_timestamp BIGINT, event_uuid TEXT, ip_address TEXT, lat REAL, lon REAL, " +
		"radius INTEGER, speed REAL);"
	statement, _ := db.Prepare(createStmt)
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

// NewDB ...
func NewDB(dbName string) *DB {
	if db != nil {
		return db
	}
	database, err := sql.Open("sqlite3", "./"+dbName)
	if err != nil {
		log.Fatal(err)
	}
	setupDB(database)
	database.SetMaxOpenConns(1)
	db = &DB{dbh: database, name: dbName}
	return db
}

// CloseHandle ...
func (db *DB) CloseHandle() {
	db.dbh.Close()
}

// GetDBName returns the name of the DB
func (db *DB) GetDBName() string {
	return db.name
}

// InsertLogin ...
func (db *DB) InsertLogin(lg *LoginEntryDAO) error {
	InsStmt := "INSERT INTO  LOGINS(username, unix_timestamp, event_uuid, ip_address, lat,lon,radius,speed)" +
		"VALUES($1,$2,$3,$4,$5,$6,$7,$8)"
	_, err := db.dbh.Exec(InsStmt, lg.UserName, lg.UnixTimeStamp, lg.EventUUID, lg.IpAddress,
		lg.Lat, lg.Lon, lg.Radius, lg.Speed)
	if err != nil {
		return err
	}
	return nil
}

// GetLoginsForUserGreaterThanOrLessThan ...
func (db *DB) GetLoginsForUserGreaterThanOrLessThan(username, operator string, ts int64) (*[]LoginEntryDAO, error) {
	log.Println("Inside GetLoginsForUserGreaterThanOrLessThan")
	columns := "username,unix_timestamp,event_uuid,ip_address,lat,lon,radius,speed"
	selectStmt := "SELECT " + columns + " from LOGINS where username=? AND unix_timestamp " +
		operator + " ?;"
	rows, err := db.dbh.Query(selectStmt, username, strconv.FormatInt(ts, 10))
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	results := make([]LoginEntryDAO, 0)
	for rows.Next() {
		lg := LoginEntryDAO{}
		err = rows.Scan(&lg.UserName, &lg.UnixTimeStamp, &lg.EventUUID, &lg.IpAddress, &lg.Lat,
			&lg.Lon, &lg.Radius, &lg.Speed)
		if err != nil {
			return nil, err
		}
		results = append(results, lg)
	}
	return &results, nil
}
