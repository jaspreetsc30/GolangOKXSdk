package public

import "cadenza-market-connector-okx/pkg/go-okx-api/models/ws"

type OrderBookEvent struct {
	Arg    ws.Args     `json:"arg"`
	Data   []OrderBook `json:"data"`
}

type OrderBook struct {
	Bids      [][]string `json:"bids"`      
	Asks      [][]string `json:"asks"`      
	Timestamp string     `json:"ts"`        
	Checksum  int32      `json:"checksum"`  
	SeqID     int64      `json:"seqId"`     
}