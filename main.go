package main

import (
	"encoding/json"
	"fmt"
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
	var rates Rates
	err := json.NewDecoder(r.Body).Decode(&rates)
	if err != nil {
		return
		// TODO:
	}
	c.Rates.Set(rates.Rates)
	c.GetRates(w, r)
}

func (c *RatesController) GetRates(w http.ResponseWriter, r *http.Request) {
	rates := c.Rates.Get()

	w.Header().Set("Content-Type", "application/json")
	out := Rates{Rates: rates}
	json.NewEncoder(w).Encode(out)
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
	var req RateRequest
	if r.Body == nil {
		return
		// TODO:
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return
		// TODO:
	}

	rates := c.Rates.Get()
	price, err := ComputePrice(rates, req.StartDate.Time, req.EndDate.Time)
	if err != nil {
		return
		// TODO:
	}

	var out interface{} = price
	if price == 0 {
		out = "unavailable"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)

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
		if (start.After(rateStart) || start == rateStart) && (end.Before(rateEnd) || end == rateEnd) && IntContains(v.GetDays(), startDay) {
			return v.Price, nil
		} else {
			fmt.Printf("Start: %v, rateStart: %v, end: %v,  rateEnd: %v, Days:%v, startDay: %v \n", start, rateStart, end, rateEnd, v.GetDays(), startDay)
		}
	}

	// rate not found
	return 0, nil
}

func NewServer(store *RateStore) *http.ServeMux {
	metricsStore := NewMetricsStore()

	metricsMiddleware := NewMetricsMiddleware(metricsStore)
	panicMiddleware := NewRecoveryMiddleware()
	ratesController := NewRatesController(store)
	priceController := NewPriceController(store)
	metricsController := NewMetricsController(metricsStore)

	mux := http.NewServeMux()

	mux.Handle("/rates", MiddlewareChain(ratesController, panicMiddleware, metricsMiddleware))
	mux.Handle("/price", MiddlewareChain(priceController, panicMiddleware, metricsMiddleware))
	mux.Handle("/metrics", panicMiddleware(metricsController))

	return mux
}

// TODO: swagger doc
func main() {
	// TODO: env port, dir
	path := "."
	if path == "" || path == "." {
		path = "./rates.json"
	}
	rateStore, err := RateStoreFromFile(path)
	if err != nil {
		panic(err)
	}
	mux := NewServer(rateStore)
	addr := ":3000"
	log.Println("Listening on " + addr)
	http.ListenAndServe(addr, mux)
}

type MetricsController struct {
	Handler
	store *MetricsStore
}

func NewMetricsController(store *MetricsStore) *MetricsController {
	controller := MetricsController{
		Handler: Handler{},
		store:   store,
	}

	controller.Handler[http.MethodGet] = http.HandlerFunc(controller.GetMetrics)
	return &controller
}

func (c *MetricsController) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := c.store.Get()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
