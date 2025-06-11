package service

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ExchangeResponse struct {
	Rates  map[string]float64 `json:"rates"`
	Result string             `json:"result"`
}

func setupRate() {
	url := "https://open.er-api.com/v6/latest/USD"

	resp, err := http.Get(url)
	if err != nil {
		log.Println("err setupRate: ", err)
	}
	defer resp.Body.Close()

	var data ExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println("err setupRate Decoder: ", err)
	}

	if data.Result != "success" {
		log.Println("err setupRate API error: ", data.Result)
	}

	vndRate, ok := data.Rates["VND"]
	if !ok {
		log.Println("VND rate not found")
	}
	now := time.Now().UTC()
	// Get midnight of the next day
	midnight := time.Date(
		now.Year(),
		now.Month(),
		now.Day()+1, // next day
		0, 0, 0, 0,
		now.Location(),
	)

	duration := midnight.Sub(now)

	rate = rateUSDtoVND{
		rate: vndRate,
		at:   now,
		ttl:  duration,
	}
}

type rateUSDtoVND struct {
	rate float64
	at   time.Time
	ttl  time.Duration
}

var rate = rateUSDtoVND{
	rate: 25995.239187,
	at:   time.Now().Add(-time.Minute * 30),
	ttl:  time.Minute * 30,
}

func CalculateVND(usd float64) float64 {
	if rate.rate == 0 || time.Now().After(rate.at.Add(rate.ttl)) {
		setupRate()
	}
	return usd * rate.rate
}
