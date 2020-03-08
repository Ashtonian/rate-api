package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RateRequest struct {
	StartDate time.Time
	EndDate   time.Time
}

func main() {
	rates := []Rate{
		Rate{
			Days:     "mon,tues,thurs",
			Times:    "0900-2100",
			Timezone: "America/Chicago",
			Price:    1500,
		},
		Rate{
			Days:     "fri,sat,sun",
			Times:    "0900-2100",
			Timezone: "America/Chicago",
			Price:    2000,
		},
		Rate{
			Days:     "wed",
			Times:    "0600-1800",
			Timezone: "America/Chicago",
			Price:    1750,
		},
		Rate{
			Days:     "mon,wed,sat",
			Times:    "0100-0500",
			Timezone: "America/Chicago",
			Price:    1000,
		},
		Rate{
			Days:     "sun,tues",
			Times:    "0100-0700",
			Timezone: "America/Chicago",
			Price:    925,
		},
	}

	chicago, _ := time.LoadLocation("America/Chicago")
	karachi, _ := time.LoadLocation("Asia/Karachi")
	requests := []RateRequest{
		RateRequest{
			StartDate: time.Date(2015, 7, 1, 7, 0, 0, 0, chicago),
			EndDate:   time.Date(2015, 7, 1, 12, 0, 0, 0, chicago),
			// Expected: 1750
		},
		RateRequest{
			StartDate: time.Date(2015, 7, 4, 15, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2015, 7, 4, 20, 0, 0, 0, time.UTC),
			// Expected: 2000
		},
		RateRequest{
			StartDate: time.Date(2015, 7, 4, 7, 0, 0, 0, karachi),
			EndDate:   time.Date(2015, 7, 4, 20, 0, 0, 0, karachi),
			// Expected: 0
		},
	}

	for _, v := range requests {
		out, err := ComputePrice(rates, v.StartDate, v.EndDate)
		if err != nil {
			fmt.Printf("ERROR: %v", err)
		}
		fmt.Printf("StartDate: %v, EndDate: %v, Out: %v \n", v.StartDate, v.EndDate, out)
	}
}

/* cases
   start and end date in diff time zone
   start and end date are different dates
   start and end date span multiple rates
   historical handle zone that changes
   rate has different day
*/

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

func IntContains(ints []int, toFind int) bool {
	for _, v := range ints {
		if v == toFind {
			return true
		}
	}
	return false
}

// given a start date and time, end date and time, and rates - this returns a valid rate
// returns 0 if prices is unavailable or input spans multiple rates or days.
// otherwise returns price offset ie if rate is $9.25 this returns 925
func ComputePrice(rates []Rate, start, end time.Time) (int, error) {
	// does the start / end span multiple days
	if start.UTC().Year() != end.UTC().Year() || start.UTC().YearDay() != end.UTC().YearDay() {
		return 0, nil
	}

	for _, v := range rates {
		rateLocation, err := time.LoadLocation(v.Timezone)
		if err != nil {
			return 0, err
		}
		startDay := int(start.In(rateLocation).Weekday())

		rateStartOffset, rateEndOffset, err := v.GetTimes()
		if err != nil {
			return 0, err
		}

		// calculate rate starts based on input start/end dates to account for historical timezone offsets and daylight savings
		rateStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		rateStart = rateStart.In(rateLocation).Add(rateStartOffset)

		rateEnd := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
		rateEnd = rateEnd.In(rateLocation).Add(rateEndOffset)

		// is input within rate range and day?
		if start.After(rateStart) && end.Before(rateEnd) && IntContains(v.GetDays(), startDay) {
			return v.Price, nil
		}
	}

	// rate not found
	return 0, nil
}
