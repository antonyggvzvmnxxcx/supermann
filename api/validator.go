package api

import (
	"net/url"
)

func (l *LoginRequest) validate() url.Values {
	errs := url.Values{}
	if l.IpAddress == "" {
		errs.Add("IpAddress", "Ip Address is missing from the event")
	}
	if l.UserName == "" {
		errs.Add("UserName", "Username is missing from the event")
	}
	if l.UnixTimeStamp <= 0 {
		errs.Add("UnixTimeStamp", "UnixTimeStamp is missing from the event")
	}
	if l.EventUUID == "" {
		errs.Add("EventUUID", "EventUUID is missing from the event")
	}
	return errs
}
