package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Controller map[string]http.Handler

func (c Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, found := c[r.Method]
	if !found {
		http.NotFound(w, r)
		return
	}
	handler.ServeHTTP(w, r)
}

type RateRequest struct {
	StartDate time.Time
	EndDate   time.Time
}

/*
Example:
	{
    	"days": "mon,tues,thurs",
        "times": "0900-2100",
        "tz": "America/Chicago",
        "price": 1500
    },
*/
type Rate struct {
	Price    int    `json:"price"`
	Timezone string `json:"tz"`
	Times    string `json:"times"`
	Days     string `json:"days"`
}

// Returns the start and end time offsets respectively
func (r Rate) GetTimes() (time.Duration, time.Duration, error) {
	timesRaw := strings.Split(r.Times, "-")
	if len(timesRaw) != 2 {
		return 0, 0, errors.New("Invalid 'times' format expected 0000-0000")
	}
	startHour, err := strconv.ParseInt(timesRaw[0][0:2], 10, 8)
	if err != nil {
		return 0, 0, err
	}
	startMinute, err := strconv.ParseInt(timesRaw[0][2:], 10, 8)
	if err != nil {
		return 0, 0, err
	}
	endHour, err := strconv.ParseInt(timesRaw[1][0:2], 10, 8)
	if err != nil {
		return 0, 0, err
	}
	endMinute, err := strconv.ParseInt(timesRaw[1][2:], 10, 8)
	if err != nil {
		return 0, 0, err
	}
	startOffset := time.Duration(startHour)*time.Hour + time.Duration(startMinute)*time.Minute
	endOffset := time.Duration(endHour)*time.Hour + time.Duration(endMinute)*time.Minute
	return startOffset, endOffset, nil
}

var days = map[string]int{
	"sun":   0,
	"mon":   1,
	"tues":  2,
	"wed":   3,
	"thurs": 4,
	"fri":   5,
	"sat":   6,
}

func (r Rate) GetDays() []int {
	daysRaw := strings.Split(r.Days, ",")
	out := make([]int, 0, len(daysRaw))
	for _, v := range daysRaw {
		out = append(out, days[v])
	}
	return out
}
