package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

/* cases TODO:
   start and end date in diff time zone
   start and end date are different dates
   start and end date span multiple rates
   historical handle zone that changes
   test minutes
   test edge of exact rate request == rate start
*/
type ComputePriceCase struct {
	Request     RateRequest
	Expected    int
	Description string
}

func TestComputePrice(t *testing.T) {
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

	cases := []ComputePriceCase{
		ComputePriceCase{
			Request: RateRequest{
				StartDate: ISO8601Time{time.Date(2015, 7, 1, 7, 0, 0, 0, chicago)},
				EndDate:   ISO8601Time{time.Date(2015, 7, 1, 12, 0, 0, 0, chicago)},
			},
			Expected:    1750,
			Description: "Normal Case 1",
		},
		ComputePriceCase{
			Request: RateRequest{
				StartDate: ISO8601Time{time.Date(2015, 7, 4, 15, 0, 0, 0, time.UTC)},
				EndDate:   ISO8601Time{time.Date(2015, 7, 4, 20, 0, 0, 0, time.UTC)},
			},
			Expected:    2000,
			Description: "Normal Case 2",
		},
		ComputePriceCase{
			Request: RateRequest{
				StartDate: ISO8601Time{time.Date(2015, 7, 4, 7, 0, 0, 0, karachi)},
				EndDate:   ISO8601Time{time.Date(2015, 7, 4, 20, 0, 0, 0, karachi)},
			},
			Expected:    0,
			Description: "Not within range",
		},
		ComputePriceCase{
			Request: RateRequest{
				StartDate: ISO8601Time{time.Date(2015, 7, 4, 15, 0, 0, 0, time.UTC)},
				EndDate:   ISO8601Time{time.Date(2015, 7, 5, 20, 0, 0, 0, time.UTC)},
			},
			Expected:    0,
			Description: "Spans Multiple days",
		},
	}

	for _, v := range cases {
		out, err := GetRate(rates, v.Request.StartDate.Time, v.Request.EndDate.Time)

		if err != nil {
			msg := "ERROR testing %v: %v"
			fmt.Printf(msg, v, err)
			t.Error(err)
		}

		assertEqual(t, v.Description, v.Expected, out)
	}
}

func assertEqual(t *testing.T, msg string, expected interface{}, found interface{}) {
	if found != expected {
		t.Fatalf("Note: %v\n Expected: %v\n Found: %v\n", msg, expected, found)
	}
}

func TestRatesEndpoint(t *testing.T) {
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
	}
	store := &RateStore{
		rates: rates,
	}
	metricsStore := NewMetricsStore()
	server := NewServer(store, metricsStore)

	t.Run("Get Rates", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/rates", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))

		expectedBod, _ := json.Marshal(Rates{Rates: rates})
		foundBod := strings.TrimSpace(response.Body.String())
		assertEqual(t, "Response Body", string(expectedBod), foundBod)
	})

	t.Run("Set Rates", func(t *testing.T) {
		rates := Rates{
			Rates: []Rate{
				Rate{
					Days:     "wed",
					Times:    "0600-1800",
					Timezone: "America/Chicago",
					Price:    1750,
				},
			},
		}
		bod, _ := json.Marshal(rates)
		postRequest, _ := http.NewRequest(http.MethodPost, "/rates", bytes.NewBuffer(bod))
		postResponse := httptest.NewRecorder()

		server.ServeHTTP(postResponse, postRequest)

		assertEqual(t, "Status Code", http.StatusOK, postResponse.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", postResponse.Header().Get("content-type"))

		getRequest, _ := http.NewRequest(http.MethodGet, "/rates", nil)
		getResponse := httptest.NewRecorder()

		server.ServeHTTP(getResponse, getRequest)
		expectedBod, _ := json.Marshal(rates)
		foundBod := strings.TrimSpace(getResponse.Body.String())
		assertEqual(t, "Response Body", string(expectedBod), foundBod)
	})
}

func TestPriceEndpoint(t *testing.T) {
	store := &RateStore{
		rates: []Rate{
			Rate{
				Days:     "wed",
				Times:    "0100-0200",
				Timezone: "America/Chicago",
				Price:    1930,
			},
		},
	}
	metricsStore := NewMetricsStore()
	server := NewServer(store, metricsStore)

	t.Run("Compute Price Unavailable", func(t *testing.T) {
		chicago, _ := time.LoadLocation("America/Chicago")
		rateRequest := RateRequest{
			StartDate: ISO8601Time{time.Date(2015, 7, 1, 1, 0, 0, 0, chicago)},
			EndDate:   ISO8601Time{time.Date(2015, 7, 1, 1, 30, 0, 0, chicago)},
		}
		bod, _ := json.Marshal(rateRequest)
		request, _ := http.NewRequest(http.MethodPost, "/rate", bytes.NewBuffer(bod))
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))
		foundBod := strings.TrimSpace(response.Body.String())
		assertEqual(t, "Response Body", `"unavailable"`, foundBod)
	})

	t.Run("Compute Price", func(t *testing.T) {
		chicago, _ := time.LoadLocation("America/Chicago")
		rateRequest := RateRequest{
			StartDate: ISO8601Time{time.Date(2015, 7, 1, 1, 1, 0, 0, chicago)},
			EndDate:   ISO8601Time{time.Date(2015, 7, 1, 1, 30, 0, 0, chicago)},
		}
		bod, _ := json.Marshal(rateRequest)
		request, _ := http.NewRequest(http.MethodPost, "/rate", bytes.NewBuffer(bod))
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))
		var found interface{}
		err := json.Unmarshal(response.Body.Bytes(), &found)
		if err != nil {
			t.Error(err)
		}
		assertEqual(t, "Response Body", float64(1930), found)
	})
}

func TestMetricsEndpoint(t *testing.T) {
	store := &RateStore{
		rates: []Rate{
			Rate{
				Days:     "mon,tues,thurs",
				Times:    "0900-2100",
				Timezone: "America/Chicago",
				Price:    1500,
			},
		},
	}
	metricsStore := NewMetricsStore()
	server := NewServer(store, metricsStore)
	t.Run("Get Metrics", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/rates", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		request, _ = http.NewRequest(http.MethodGet, "/rates", nil)
		response = httptest.NewRecorder()
		server.ServeHTTP(response, request)

		request, _ = http.NewRequest(http.MethodGet, "/metrics", nil)
		response = httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))
	})
}

var defaultRates = []Rate{
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

func TestStoreFromFile(t *testing.T) {
	store, err := RateStoreFromFile("./rates.json")
	if err != nil {
		t.Error(err)
	}
	expected := defaultRates
	assertEqual(t, "Store Reads JSON", fmt.Sprintf("%v", expected), fmt.Sprintf("%v", store.Get()))
}

func TestRatesJSON(t *testing.T) {
	jsonRaw := `{"startDate":"2015-07-01T07:00:00-05:00","endDate":"2015-07-01T12:00:00-05:00"}`

	var obj RateRequest
	err := json.Unmarshal([]byte(jsonRaw), &obj)
	if err != nil {
		t.Error(err)
	}

	loc, _ := time.LoadLocation("America/Chicago")
	expectedStart := fmt.Sprintf("%v", time.Date(2015, 7, 1, 7, 0, 0, 0, loc))
	assertEqual(t, "Deserialize Rate Request", expectedStart, fmt.Sprintf("%v", obj.StartDate))

	out, err := json.Marshal(obj)
	if err != nil {
		t.Error(err)
	}
	assertEqual(t, "Serialize Rate Request", string(out), jsonRaw)
}

func TestMetricsRecord(t *testing.T) {
	store := NewMetricsStore()

	store.Record("test", "test", 200, 100)
	store.Record("test", "test", 201, 300)
	store.Record("test", "test2", 201, 300)

	metrics := store.Get()
	expected := map[string]Metrics{
		"all|all": Metrics{
			AvgResponseTime: 233,
			RequestCount:    3,
			StatusCodeCount: map[int]int{
				200: 1,
				201: 2,
			},
		},
		"test|test": Metrics{
			AvgResponseTime: 200,
			RequestCount:    2,
			StatusCodeCount: map[int]int{
				200: 1,
				201: 1,
			},
		},
		"test|test2": Metrics{
			AvgResponseTime: 300,
			RequestCount:    1,
			StatusCodeCount: map[int]int{
				201: 1,
			},
		},
	}

	assertEqual(t, "Metrics Record Results", fmt.Sprintf("%v", expected), fmt.Sprintf("%v", metrics.Metrics))
}
