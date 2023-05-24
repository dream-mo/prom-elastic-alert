package xtime

import (
	"time"
)

var Zone *time.Location

const (
	ISO8601 = "2006-01-02T15:04:05.000Z07:00"
)

type TimeLimit struct {
	Seconds uint `json:"seconds" yaml:"seconds"`
	Minutes uint `json:"minutes" yaml:"minutes"`
	Days    uint `json:"days" yaml:"days"`
}

func (t TimeLimit) GetTimeDuration() time.Duration {
	s := time.Duration(t.Seconds) * time.Second
	s += time.Duration(t.Minutes) * time.Minute
	s += time.Duration(t.Days*24) * time.Hour
	return s
}

func (t TimeLimit) GetSeconds() int {
	s := t.Seconds
	s += t.Minutes * 60
	s += t.Days * 3600 * 24
	return int(s)
}

func FixedZone(zoneStr string) {
	z, err := time.LoadLocation(zoneStr)
	if err != nil {
		panic(err)
	} else {
		Zone = z
	}
}

func Now() time.Time {
	t := time.Now()
	t.In(Zone)
	return t
}

func TimeFormatISO8601(t time.Time) string {
	return t.In(Zone).Format(ISO8601)
}

func Parse(s string) time.Time {
	t, _ := time.ParseInLocation(time.RFC3339, s, Zone)
	return t
}
