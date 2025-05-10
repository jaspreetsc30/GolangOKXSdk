package public

import "cadenza-market-connector-okx/pkg/go-okx-api/models/ws"

type CandleEvent struct {
    Arg  ws.Args     `json:"arg"`
    Data []Candlestick `json:"data"`
}

// Candlestick represents a single 1-minute candlestick
type Candlestick struct {
    Timestamp string `json:"ts"`    // Unix timestamp in milliseconds
    Open      float64 `json:"open"`  // Opening price
    High      float64 `json:"high"`  // Highest price
    Low       float64 `json:"low"`   // Lowest price
    Close     float64 `json:"close"` // Closing price
    Volume    float64 `json:"vol"`   // Volume (in base currency)
}