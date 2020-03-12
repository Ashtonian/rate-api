package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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

// PostRates - updates the current active rates based on user input.
// @Summary Updates the current active rates based on user input.
// @Description Updates the current active rates based on user input.
// @Tags rates
// @Accept json
// @Produce json
// @Param Rates body Rates true "Update Rates"
// @Success 200 {object} Rates
// @Failure 400 {object} ErrorResponse
// @Failure 404 ""
// @Failure 500 {object} ErrorResponse
// @Router /rates/ [post]
func (c *RatesController) PostRates(w http.ResponseWriter, r *http.Request) {
	var rates Rates
	if r.Body == nil {
		webError(w, http.StatusBadRequest, ErrMissingBody)
	}
	err := json.NewDecoder(r.Body).Decode(&rates)
	if err != nil {
		webError(w, http.StatusBadRequest, ErrBadBody)
		return
	}
	c.Rates.Set(rates.Rates)
	c.GetRates(w, r)
}

// GetRates - Gets the current active rates.
// @Summary Gets the current active rates.
// @Description Gets the current active rates.
// @Tags rates
// @Accept json
// @Produce json
// @Success 200 {object} Rates
// @Failure 400 {object} ErrorResponse
// @Failure 404 ""
// @Failure 500 {object} ErrorResponse
// @Router /rates/ [get]
func (c *RatesController) GetRates(w http.ResponseWriter, r *http.Request) {
	rates := c.Rates.Get()

	w.Header().Set("Content-Type", "application/json")
	out := Rates{Rates: rates}
	json.NewEncoder(w).Encode(out)
}

type RateController struct {
	Handler
	Rates *RateStore
}

func NewRateController(store *RateStore) *RateController {
	controller := RateController{
		Handler: Handler{},
		Rates:   store,
	}
	controller.Handler[http.MethodPost] = http.HandlerFunc(controller.GetRate)

	return &controller
}

// GetRate - Given the time range input this returns the rate as int, or "unavailable".
// @Summary Given the time range input this returns the rate as int, or "unavailable".
// @Description Given the time range input this returns the rate as int, or "unavailable".
// @Tags rates
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 404 ""
// @Failure 500 {object} ErrorResponse
// @Router /rate [post]
func (c *RateController) GetRate(w http.ResponseWriter, r *http.Request) {
	var req RateRequest
	if r.Body == nil {
		webError(w, http.StatusBadRequest, ErrMissingBody)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		webError(w, http.StatusBadRequest, ErrBadBody)
		return
	}

	rates := c.Rates.Get()
	rate, err := GetRate(rates, req.StartDate.Time, req.EndDate.Time)
	if err != nil {
		webError(w, http.StatusInternalServerError, ErrInternal)
		return
	}

	var out interface{} = rate
	if rate == 0 {
		out = "unavailable"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)

}

// given a start date and time, end date and time, and rates - this returns a valid rate
// returns 0 if rates is unavailable or input spans multiple rates or days.
// otherwise returns rate offset ie if rate is $9.25 this returns 925
func GetRate(rates []Rate, start, end time.Time) (int, error) {
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
		if (start.After(rateStart) || start.Equal(rateStart)) && (end.Before(rateEnd) || end.Equal(rateEnd)) && IntContains(v.GetDays(), startDay) {
			return v.Price, nil
		}
	}

	// rate not found
	return 0, nil
}

func NewServer(rateStore *RateStore, metricsStore *MetricsStore) *http.ServeMux {
	metricsMiddleware := NewMetricsMiddleware(metricsStore)
	panicMiddleware := NewRecoveryMiddleware()
	ratesController := NewRatesController(rateStore)
	rateController := NewRateController(rateStore)
	metricsController := NewMetricsController(metricsStore)

	mux := http.NewServeMux()

	mux.Handle("/rates", MiddlewareChain(ratesController, panicMiddleware, metricsMiddleware))
	mux.Handle("/rate", MiddlewareChain(rateController, panicMiddleware, metricsMiddleware))
	mux.Handle("/metrics", panicMiddleware(metricsController))

	return mux
}

// @title Rate API
// @version 1.0
// @description Rate-Api allows a user to enter a date time range and get back the rate at which they would be charged to park for that time span built for spot hero.

// @contact.name API Support
// @contact.email support@todo.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /
func main() {
	path := os.Getenv("RATE_API_RATES_PATH")
	if path == "" || path == "." {
		path = "./rates.json"
	}
	portStr := os.Getenv("RATE_API_PORT")
	if len(portStr) < 1 {
		portStr = "3000"
	}
	rateStore, err := RateStoreFromFile(path)
	if err != nil {
		panic(err)
	}
	metricsStore := NewMetricsStore()
	mux := NewServer(rateStore, metricsStore)
	addr := ":" + portStr
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

// GetMetrics - Gets the api health metrics available.
// @Summary Gets the api health metrics available.
// @Description Gets the api health metrics available.
// @Tags metrics
// @Accept json
// @Produce json
// @Success 200 {object} EndpointMetrics
// @Failure 400 {object} ErrorResponse
// @Failure 404 ""
// @Failure 500 {object} ErrorResponse
// @Router /metrics [get]
func (c *MetricsController) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := c.store.Get()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
