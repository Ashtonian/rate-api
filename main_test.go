package main

import (
	"fmt"
	"testing"
	"time"
)

/* cases
   start and end date in diff time zone
   start and end date are different dates
   start and end date span multiple rates
   historical handle zone that changes
   rate has different day
*/
type ComputePriceCase struct {
	Request  RateRequest
	Expected int
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
				StartDate: time.Date(2015, 7, 1, 7, 0, 0, 0, chicago),
				EndDate:   time.Date(2015, 7, 1, 12, 0, 0, 0, chicago),
			},
			Expected: 1750,
		},
		ComputePriceCase{
			Request: RateRequest{
				StartDate: time.Date(2015, 7, 4, 15, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2015, 7, 4, 20, 0, 0, 0, time.UTC),
			},
			Expected: 2000,
		},
		ComputePriceCase{
			Request: RateRequest{
				StartDate: time.Date(2015, 7, 4, 7, 0, 0, 0, karachi),
				EndDate:   time.Date(2015, 7, 4, 20, 0, 0, 0, karachi),
			},
			Expected: 0,
		},
	}

	for _, v := range cases {
		out, err := ComputePrice(rates, v.Request.StartDate, v.Request.EndDate)

		if err != nil {
			fmt.Printf("ERROR testing %v: %v", v, err)
			t.Error(err)
		}

		assertEqual(t, out, v.Expected)
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
