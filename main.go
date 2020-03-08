package main

import (
	"log"
	"net/http"
	"time"
)

func PostRates(w http.ResponseWriter, r *http.Request) {
	println("PostRates")
}

func GetRates(w http.ResponseWriter, r *http.Request) {
	println("GetRates")
}

func ComputeRate(w http.ResponseWriter, r *http.Request) {
	println("ComputeRate")
}

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	println("GetMetrics")
}

func main() {
	ratesController := Controller{}
	ratesController[http.MethodPost] = http.HandlerFunc(PostRates)
	ratesController[http.MethodGet] = http.HandlerFunc(GetRates)

	priceController := Controller{}
	priceController[http.MethodPost] = http.HandlerFunc(ComputeRate)

	metricsController := Controller{}
	metricsController[http.MethodGet] = http.HandlerFunc(GetMetrics)

	mux := http.NewServeMux()

	mux.Handle("/rates", ratesController)
	mux.Handle("/price", priceController)
	mux.Handle("/metrics", metricsController)

	addr := ":3000"
	log.Println("Listening on " + addr)
	http.ListenAndServe(addr, mux)
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

func IntContains(ints []int, toFind int) bool {
	for _, v := range ints {
		if v == toFind {
			return true
		}
	}
	return false
}
