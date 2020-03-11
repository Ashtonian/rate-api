package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type RatesController struct {
	Handler
	Rates *RateStore
}

func NewRatesController(store *RateStore) *RatesController {
	controller := RatesController{
		Handler: Handler{},
		Rates:   store,
	}
	controller.Handler[http.MethodPost] = http.HandlerFunc(controller.PostRates)
	controller.Handler[http.MethodGet] = http.HandlerFunc(controller.GetRates)
	return &controller
}

func (c *RatesController) PostRates(w http.ResponseWriter, r *http.Request) {
	println("PostRates")
	w.Header().Set("Content-Type", "application/json")
}

func (c *RatesController) GetRates(w http.ResponseWriter, r *http.Request) {
	rates := c.Rates.Get()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rates)
}

type PriceController struct {
	Handler
	Rates *RateStore
}

func NewPriceController(store *RateStore) *PriceController {
	controller := PriceController{
		Handler: Handler{},
		Rates:   store,
	}
	controller.Handler[http.MethodPost] = http.HandlerFunc(controller.ComputeRate)

	return &controller
}

func (c *PriceController) ComputeRate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req RateRequest
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		// TODO:
	}
	price, err := ComputePrice([]Rate{}, req.StartDate.Time, req.EndDate.Time)
	if err != nil {
		// TODO:
	}

	var out interface{} = price
	if price == 0 {
		out = "unavailable"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
	return
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

func NewServer(store *RateStore) *http.ServeMux {
	// TODO: panic handler middleware
	// TODO: metric middleware

	ratesController := NewRatesController(store)
	priceController := NewPriceController(store)

	metricsController := Handler{}
	metricsController[http.MethodGet] = http.HandlerFunc(GetMetrics)

	mux := http.NewServeMux()

	mux.Handle("/rates", ratesController)
	mux.Handle("/price", priceController)
	mux.Handle("/metrics", metricsController)

	return mux
}

// TODO: swagger doc
func main() {
	rateStore, err := RateStoreFromFile(".")
	if err != nil {
		panic(err)
	}
	mux := NewServer(rateStore)
	addr := ":3000"
	log.Println("Listening on " + addr)
	http.ListenAndServe(addr, mux)
}

// TODO: metric store
// TODO: status code counts, response time
func GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	println("GetMetrics")
}
