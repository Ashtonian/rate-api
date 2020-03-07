package main

import (
	"fmt"
	"time"
)

func main() {
	// `2015-07-01T07:00:00-05:00` to `2015-07-01T12:00:00-05:00` should yield `1750`
	out := ComputePrice(time.Date(2015, 7, 1, 7, 0, 0, 0, time.UTC), time.Date(2015, 7, 1, 12, 0, 0, 0, time.UTC))
	println(out)
}

/* cases
   start and end date in diff time zone
   start and end date are different dates
   start and end date span multiple rates
   historical handle zone that changes
*/

/** `2015-07-01T07:00:00-05:00` to `2015-07-01T12:00:00-05:00` should yield `1750`
* `2015-07-04T15:00:00+00:00` to `2015-07-04T20:00:00+00:00` should yield `2000`
* `2015-07-04T07:00:00+05:00` to `2015-07-04T20:00:00+05:00` should yield `unavailable`
 */

// given a start date and time and end date and time this returns a valid rate
// returns 0 if prices is unavailable or input spans multiple rates or days.
// otherwise returns price offset ie if rate is $9.25 this returns 925
func ComputePrice(start, end time.Time) int {
	// TODO: handle tz

	// if the times span multiple days
	if start.UTC().Year() != end.UTC().Year() || start.UTC().YearDay() != end.UTC().YearDay() {
		return 0
	}

	startHour := 6
	startMinute := 0
	rate := 1750
	endHour := 18
	endMinute := 0
	rateLocations := start.Location()
	startDuration := time.Duration(startHour)*time.Hour + time.Duration(startMinute)*time.Minute
	endDuration := time.Duration(endHour)*time.Hour + time.Duration(endMinute)*time.Minute

	rateStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	rateStart = rateStart.In(rateLocations).Add(startDuration)

	rateEnd := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	rateEnd = rateEnd.In(rateLocations).Add(endDuration)

	fmt.Printf("rateStart: %v, rateEnd: %v, start: %v, end: %v\n", rateStart, rateEnd, start, end)
	if start.After(rateStart) && end.Before(rateEnd) {
		return rate
	}

	return 0
}
