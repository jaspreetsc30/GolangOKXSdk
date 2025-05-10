package business

import (
	"cadenza-market-connector-okx/pkg/go-okx-api/models/ws"
	"encoding/json"
	"fmt"
	"strconv"
)

type CandleEvent struct {
    Arg  ws.Args     `json:"arg"`
    Data []Candlestick `json:"data"`
}


type Candlestick struct {
	Timestamp     int64   `json:"ts"`
	Open          float64 `json:"o"`
	High          float64 `json:"h"`
	Low           float64 `json:"l"`
	Close         float64 `json:"c"`
	Volume        float64 `json:"vol"`
	VolumeCcy     float64 `json:"volCcy"`
	VolumeCcyQuote float64 `json:"volCcyQuote"`
	Confirm       bool    `json:"confirm"`
}

// UnmarshalJSON customizes parsing for Candlestick
func (c *Candlestick) UnmarshalJSON(b []byte) error {
	var raw []string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if len(raw) != 9 {
		return fmt.Errorf("expected 9 fields in candlestick data, got %d", len(raw))
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(raw[0], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %s: %v", raw[0], err)
	}
	c.Timestamp = ts

	for i, field := range []struct {
		target *float64
		name   string
	}{
		{&c.Open, "open"},
		{&c.High, "high"},
		{&c.Low, "low"},
		{&c.Close, "close"},
		{&c.Volume, "volume"},
		{&c.VolumeCcy, "volumeCcy"},
		{&c.VolumeCcyQuote, "volumeCcyQuote"},
	} {
		f, err := strconv.ParseFloat(raw[i+1], 64)
		if err != nil {
			return fmt.Errorf("failed to parse %s %s: %v", field.name, raw[i+1], err)
		}
		*field.target = f
	}

	switch raw[8] {
	case "0":
		c.Confirm = false
	case "1":
		c.Confirm = true
	default:
		return fmt.Errorf("invalid confirm value: %s", raw[8])
	}

	return nil
}