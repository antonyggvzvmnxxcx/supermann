package api

import (
	"testing"
	"time"

	"github.com/anyaddres/supermann/config"
	ds "github.com/anyaddres/supermann/datastore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDistanceBetweenLocations(t *testing.T) {
	loc1 := Location{Lat: 22.55, Lon: 43.12}
	loc2 := Location{Lat: 13.45, Lon: 100.28}
	miles, kms := getDistanceBetweenLocations(loc1, loc2)
	assert.Equal(t, 3786.251258825624, miles, "The two distances in miles should be equal")
	assert.Equal(t, 6094.544408786774, kms, "The two distances in kms should equal")
}

func TestFindMin(t *testing.T) {
	events := make([]Events, 0)
	evOne := Events{Ip: "187.34.174.242", TimeStamp: 1342378499}
	evTwo := Events{Ip: "18.118.60.44", TimeStamp: 845719794}
	events = append(events, []Events{evOne, evTwo}...)
	// minStream := make(chan *Events, 0)
	min := findMin(events)
	// min := <-minStream
	assert.Equal(t, evTwo, *min, "The second event should be the right event")
}

func TestFindMax(t *testing.T) {
	events := make([]Events, 0)
	evOne := Events{Ip: "18.118.60.44", TimeStamp: 845719794}
	evTwo := Events{Ip: "187.34.174.242", TimeStamp: 1342378499}
	events = append(events, []Events{evOne, evTwo}...)
	// maxStream := make(chan *Events, 0)
	max := findMax(events)
	// max := <-maxStream
	assert.Equal(t, evTwo, *max, "The second event should be the one with the larger timestamp")
}

func TestGetLatLonForIp(t *testing.T) {
	ctx := &SrvContext{cfg: config.GetConfig()}
	entry := &LoginRequest{IpAddress: "123.192.212.224"}
	latLon, err := getLatLonForIP(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 25.0478, latLon.Lat, "The two lats should be equal")
	assert.Equal(t, 121.5318, latLon.Lon, "The two lons should be equal")
}

func TestIsTravelSuspiciousTrue(t *testing.T) {
	entry := &LoginRequest{UnixTimeStamp: time.Now().Unix()}
	loc1 := Location{Lat: 38.291962, Lon: -122.458000} // Sonoma,CA
	latLonReq := LoginInfo{Location: loc1}
	loc2 := Location{Lat: 39.952583, Lon: -75.165222} // Philadelphia, PA
	loginInfo := LoginInfo{Location: loc2}
	event := &Events{LoginInfo: latLonReq, TimeStamp: time.Now().Add(-10 * time.Minute).Unix()}
	_, isSuspiciousTravel := isTravelSuspicious(entry, &loginInfo, event)
	assert.True(t, isSuspiciousTravel, "Travel should be suspicious")
}

func TestIsTravelSuspiciousFalse(t *testing.T) {
	entry := &LoginRequest{UnixTimeStamp: time.Now().Unix()}
	loc1 := Location{Lat: 38.291962, Lon: -122.458000} // Sonoma,CA
	latLonReq := LoginInfo{Location: loc1}
	loc2 := Location{Lat: 39.952583, Lon: -75.165222} // Philadelphia, PA
	loginInfo := LoginInfo{Location: loc2}
	event := &Events{LoginInfo: latLonReq, TimeStamp: time.Now().Add(48 * time.Hour).Unix()}
	_, isSuspiciousTravel := isTravelSuspicious(entry, &loginInfo, event)
	assert.False(t, isSuspiciousTravel, "Travel should not be suspicious")
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetLoginsForUserGreaterThanOrLessThan(username, operator string, ts int64) (*[]ds.LoginEntryDAO, error) {
	args := m.Called(username, operator, ts)
	return args.Get(0).(*[]ds.LoginEntryDAO), args.Error(1)
}

func TestClosestNeighbouringLogins(t *testing.T) {
	testObj := new(MockDB)

	// Login entry 1st of January 2017
	lr := &LoginRequest{UserName: "bob", UnixTimeStamp: 1483246800, IpAddress: "18.118.60.44",
		EventUUID: "85ad929a-db03-4bf4-9541-8f728fa12e42"}

	// 2 entries after the login entry date
	loginGreaterEntries := make([]ds.LoginEntryDAO, 0)
	lrg1 := ds.LoginRequestDAO{UserName: "bob", EventUUID: "85ad929a-db03-4bf4-9541-8f728fa12e42",
		UnixTimeStamp: 1483419600} // 3rd Jan 2017
	lrg2 := ds.LoginRequestDAO{UserName: "bob", EventUUID: "85ad929a-db03-4bf4-9541-8f728fa12e42",
		UnixTimeStamp: 1483333200} // 2nd Jan 2017
	leg1 := ds.LoginEntryDAO{LoginRequestDAO: lrg1, LoginInfoDAO: ds.LoginInfoDAO{}}
	leg2 := ds.LoginEntryDAO{LoginRequestDAO: lrg2, LoginInfoDAO: ds.LoginInfoDAO{}}

	loginGreaterEntries = append(loginGreaterEntries, []ds.LoginEntryDAO{leg1, leg2}...)
	testObj.On("GetLoginsForUserGreaterThanOrLessThan", "bob", ">", int64(1483246800)).Return(&loginGreaterEntries, nil)

	// 2 entries before the login entry date.
	loginSmallerEntries := make([]ds.LoginEntryDAO, 0)
	lrs1 := ds.LoginRequestDAO{UserName: "bob", EventUUID: "85ad929a-db03-4bf4-9541-8f728fa12e42",
		UnixTimeStamp: 1483160400} // 31st of December 2016
	lrs2 := ds.LoginRequestDAO{UserName: "bob", EventUUID: "85ad929a-db03-4bf4-9541-8f728fa12e42",
		UnixTimeStamp: 1483074000} // 30th of December 2016
	les1 := ds.LoginEntryDAO{LoginRequestDAO: lrs1, LoginInfoDAO: ds.LoginInfoDAO{}}
	les2 := ds.LoginEntryDAO{LoginRequestDAO: lrs2, LoginInfoDAO: ds.LoginInfoDAO{}}
	loginSmallerEntries = append(loginSmallerEntries, []ds.LoginEntryDAO{les1, les2}...)
	testObj.On("GetLoginsForUserGreaterThanOrLessThan", "bob", "<", int64(1483246800)).Return(&loginSmallerEntries, nil)

	latLong := &LoginInfo{}
	prev, next, _ := closestNeighbouringLogins(testObj, lr, latLong)
	assert.Equal(t, int64(1483160400), prev.TimeStamp, "Previous login entry should be equal to 1483160400")
	assert.Equal(t, int64(1483333200), next.TimeStamp, "Next login entry should be equal to 1483333200")
}
