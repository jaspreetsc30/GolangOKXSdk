package public

import "cadenza-market-connector-okx/pkg/go-okx-api/models/ws"

type TradeEvent struct {
	Arg  ws.Args `json:"arg"`
	Data []Trade `json:"data"`
}

type Trade struct {
	InstId  string `json:"instId"`
	TradeId string `json:"tradeId"`
	Px      string `json:"px"`
	Sz      string `json:"sz"`
	Side    string `json:"side"`
	Ts      int64  `json:"ts,string"`
}