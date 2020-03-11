package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

/* cases TODO:
   start and end date in diff time zone
   start and end date are different da tes
   start and end date span multiple rates
   historical handle zone that changes
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
		out, err := ComputePrice(rates, v.Request.StartDate.Time, v.Request.EndDate.Time)

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
		t.Fatalf("Note: %v Expected: %v Found: %v", msg, expected, found)
	}
}

func TestRatesEndpoint(t *testing.T) {
	store := &RateStore{
		rates: []Rate{
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
		},
	}
	server := NewServer(store)

	t.Run("Get Rates", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/rates", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))
		// TODO: response body
	})

	t.Run("Set Rates", func(t *testing.T) {
		// TODO: req body
		request, _ := http.NewRequest(http.MethodPost, "/rates", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))
		// TODO: response body
	})
}

func TestPriceEndpoint(t *testing.T) {
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
	server := NewServer(store)

	t.Run("Compute Price", func(t *testing.T) {
		// TODO: req body
		request, _ := http.NewRequest(http.MethodPost, "/price", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))
		// TODO: response body
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
	server := NewServer(store)

	t.Run("Get Metrics", func(t *testing.T) {
		// TODO: req body
		request, _ := http.NewRequest(http.MethodPost, "/price", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertEqual(t, "Status Code", http.StatusOK, response.Result().StatusCode)
		assertEqual(t, "Content Type", "application/json", response.Header().Get("content-type"))
		// TODO: response body
	})
}
