# SuperMan Suspicious Login Detector
A RESTful API to check for suspicious login events by users based on GeoIP Mapping.

## Prerequisites
You need to have go1.11.1 setup. Dependencies are managed via go mod vendor and are already
packaged in the vendor/ folder.

## Installation & Run
```bash
# Download the project either to the GOPATH directory or outside the GOPATH if you prefer.
# If you want to clone the package outside the GOPATH then use the following command
git clone https://github.com/anyaddres/supermann ~/supermann

# If you want to clone inside the GOPATH(GOPATH environment variable should be set)
go get -d github.com/anyaddres/supermann. The project does not have .go files outside folders so you might 
get the `no Go files in` error. The package will be downloaded but not built.
```
You can choose to run the server without a docker image. Before running the API server outside the docker container, make sure to download the [GeoLite2-City.tar.gz](http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz) file and unzip the mmdb file in the `/GeoLite2/` location. That is the default location. The value can be overridden by setting the `GEO_IP_DB` environment variable. The SQLite  db file location
can also be overriden by setting the `DATABASE_FILE` environment variable.

The docker build adds GeoLite2-City.mmdb from the maxmind site to the docker container. This was one way to do it. The other way would 
be to mount a file from the localhost to the docker container. However, for this demo I chose to go with the self contained geoip file.

```bash
# Build, Run & Access
cd superman
go build cmd/logins/superman.go

MMDB_FILE_PATH is the path on your system where you have untarred the GeoLite mmdb file. 

To run the superman binary you can run the following command: GEO_IP_DB=MMDB_FILE_PATH ./superman

To run the tests you can run the following command : GEO_IP_DB=MMDB_FILE_PATH go test -v ./...

API Endpoint : http://127.0.0.1:8080/api/identifylogins/ Handles only POST Method.

curl -X POST -d \
    '{"username": "bob",
      "unix_timestamp": 590729457,
      "event_uuid": "85ad929a-db03-4bf4-9541-8f728fa12e42",
      "ip_address": "82.233.123.117"}' http://127.0.0.1:8080/api/identifylogins/
```
## Docker Image build.
1. Run the build command from the same folder where the Dockerfile resides.
```bash
docker build -t sworks/superman-1.0 .
```
## Executing the Docker image
``` bash
docker run -d -p 8080:8080 sworks/superman-1.0:latest
```
## Structure
```
├── api
│   ├── api.go            // Core API Handler
│   ├── errors.go         // API Error handling
│   ├── helpers.go        // API Helper functions.
│   ├── logins.go         // API Request/Response Objects
│   └── validator.go      // API Validation
├── cmd
│   ├── logins
│   │   └── superman.go   // The main command 
│   └── perf
│       └── login_perf.go // A Test Program for the API
├── config
│   └── settings.go       // Configuration
├── datastore
│   ├── dao.go            // Data Access Objects
│   └── db.go             // DB Functions
├── Dockerfile
├── geoip
│   └── geoip.go          // Geoip Setup.
├── go.mod
├── go.sum
├── README.md
└── vendor                // Vendored Dependencies
```

## API

#### /api/identifylogins/ 
* `POST` : Detects a suspicious login and reponds with previous and subsequents events to the current event if they exist. The Suspicious travel attribute denotes whether the user could travel from one location to another in the time between the occurrence of the 2 events such that he/she would need to travel more than 500 miles per hour.

## External Libraries

* [MaxMind DB Reader](https://github.com/oschwald/maxminddb-golang) Go Reader for MaxMind DB
* [SQLite3 Driver](https://github.com/mattn/go-sqlite3) SQLite Driver
* [Haversine](https://github.com/umahmood/haversine) Haversine Calculations
* [EnvDecode](https://github.com/joeshaw/envdecode) Populate Struct from Env Vars


## Output
* First Request for a new user
```json
{
  "currentGeo": {
    "lat": 38.317,
    "lon": -88.9105,
    "radius": 50
  }
}
```
* Second Request for same user.
```json
{
  "currentGeo": {
    "lat": 37.751,
    "lon": -97.822,
    "radius": 1000
  },
  "subsequentIpAccess": {
    "lat": 38.317,
    "lon": -88.9105,
    "radius": 50,
    "speed": 0.003004705539353452,
    "ip": "66.232.172.64",
    "timestamp": 819456674,
    "suspiciousTravel": false
  }
}
```
*Third request for the same user
```json
{
  "currentGeo": {
    "lat": 34.7725,
    "lon": 113.7266,
    "radius": 50
  },
  "precedingIpAccess": {
    "lat": 37.751,
    "lon": -97.822,
    "radius": 1000,
    "speed": 0.06455826393121096,
    "ip": "75.203.160.217",
    "timestamp": 236854267,
    "suspiciousTravel": false
  },
  "subsequentIpAccess": {
    "lat": 38.317,
    "lon": -88.9105,
    "radius": 50,
    "speed": 0.13571846927395814,
    "ip": "66.232.172.64",
    "timestamp": 819456674,
    "suspiciousTravel": false
  }
}
```

## Performance

* I ran the application on C4,R4 and T2 AWS instances. I then called the API from 3 other C5.4xLarge machines. I got response times of 20ms max. Putting log messages of execution time the DB scans were the bottle neck. I will see if I can use a different db or some other caching(memcache) to minimize db hits. I ran a million requests against the api running on the vm directly. I did notice that the response times went up as the number of events hit a million.

## Todo

- [ ] Use gorilla if more sophisticated handlers are needed. 
      The current url matcher is a simple solution for a direct mapping.
- [ ] Support Authentication with user for securing the APIs.
- [ ] Improve performance of the response times. DB Scan's are the bottle neck
- [ ] Does not handle duplicate events. Duplicate event would be an event with the all the same value
     except the UUID of the event and TimeStamp. Same IPAddress, Username. The user is trying to login from the IP Address again and again. Basically could be brute forcing.