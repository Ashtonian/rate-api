package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler map[string]http.Handler

func (c Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, found := c[r.Method]
	if !found {
		http.NotFound(w, r)
		return
	}
	handler.ServeHTTP(w, r)
}

type RateRequest struct {
	StartDate ISO8601Time `json:"startDate"`
	EndDate   ISO8601Time `json:"endDate"`
}

type Rates struct {
	Rates []Rate `json:"rates"`
}

/*
Example Rate:
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

var ISO8601 = "2006-01-02T15:04:05-07:00"

type ISO8601Time struct {
	time.Time
}

func (t *ISO8601Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	out, err := time.Parse(ISO8601, s)
	if err != nil {
		return err
	}
	*t = ISO8601Time{out}
	return nil
}

func (t ISO8601Time) MarshalJSON() ([]byte, error) {
	out := fmt.Sprintf("\"%s\"", t.Format(ISO8601))
	return []byte(out), nil
}

type RateStore struct {
	rates []Rate
}

func (store *RateStore) Get() []Rate {
	return store.rates
}

func (store *RateStore) Set(rates []Rate) {
	store.rates = rates
}

func RateStoreFromFile(path string) (*RateStore, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rates Rates

	err = json.Unmarshal(file, &rates)
	if err != nil {
		return nil, err
	}
	store := &RateStore{
		rates: rates.Rates,
	}
	return store, nil
}

func IntContains(ints []int, toFind int) bool {
	for _, v := range ints {
		if v == toFind {
			return true
		}
	}
	return false
}

type EndpointMetrics struct {
	Metrics map[string]Metrics `json:"metrics"`
}
type Metrics struct {
	AvgResponseTime int         `json:"ms"`
	RequestCount    int         `json:"requestCount"`
	StatusCodeCount map[int]int `json:"statusCodeCount"`
}

type MetricsStore struct {
	metrics map[string]Metrics
	allKey  string
}

func NewMetricsStore() *MetricsStore {
	allKey := "all|all"
	return &MetricsStore{
		allKey: allKey,
		metrics: map[string]Metrics{
			allKey: Metrics{
				StatusCodeCount: map[int]int{},
			},
		},
	}
}

func (_ *MetricsStore) getKey(method, path string) string {
	return fmt.Sprintf("%s|%s", method, path)
}

func (store *MetricsStore) Get() EndpointMetrics {
	return EndpointMetrics{
		Metrics: store.metrics,
	}
}

func (store *MetricsStore) Record(method, path string, statusCode, responseMs int) {
	allMetrics, _ := store.metrics[store.allKey]
	allMetrics.RequestCount++
	allMetrics.AvgResponseTime = allMetrics.AvgResponseTime + (responseMs-allMetrics.AvgResponseTime)/allMetrics.RequestCount
	allMetrics.StatusCodeCount[statusCode]++
	store.metrics[store.allKey] = allMetrics

	mKey := store.getKey(method, path)
	v, found := store.metrics[mKey]
	if !found {
		store.metrics[mKey] = Metrics{
			AvgResponseTime: responseMs,
			RequestCount:    1,
			StatusCodeCount: map[int]int{statusCode: 1},
		}
		return
	}

	v.RequestCount++
	v.AvgResponseTime = v.AvgResponseTime + (responseMs-v.AvgResponseTime)/v.RequestCount
	v.StatusCodeCount[statusCode]++
	store.metrics[mKey] = v
}

type ErrorResponse struct {
	Error string `json:"error"`
}
