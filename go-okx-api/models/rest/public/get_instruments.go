package public

import "cadenza-market-connector-okx/pkg/go-okx-api/models/rest"



func NewGetInstruments(param *GetInstrumentsParam) (rest.IRequest, rest.IResponse) {
	return &rest.Request{
		Path:   "/api/v5/public/instruments",
		Method: rest.MethodGet,
		Param:  param,
	}, &GetInstrumentsResponse{}
}

type GetInstrumentsParam struct {
	InstType string `url:"instType,omitempty"` // Instrument type (e.g., SPOT, SWAP, FUTURES, OPTION)
	Uly      string `url:"uly,omitempty"`      // Underlying index, applicable to FUTURES/SWAP/OPTION
	InstId   string `url:"instId,omitempty"`   // Instrument ID, e.g., BTC-USD-190927
}

type GetInstrumentsResponse struct {
	rest.Response
	Data []Instrument `json:"data"`
}

type Instrument struct {
	InstId        string `json:"instId"`
	InstType      string `json:"instType"`
	Uly           string `json:"uly,omitempty"`
	Category      string `json:"category"`
	BaseCcy       string `json:"baseCcy"`
	QuoteCcy      string `json:"quoteCcy"`
	SettleCcy     string `json:"settleCcy"`
	CtVal         string `json:"ctVal"`
	CtMult        string `json:"ctMult"`
	CtValCcy      string `json:"ctValCcy"`
	OptType       string `json:"optType,omitempty"`
	Stk           string `json:"stk,omitempty"`
	ListTime      string `json:"listTime"`
	ExpTime       string `json:"expTime,omitempty"`
	Lever         string `json:"lever,omitempty"`
	TickSz        string `json:"tickSz"`
	LotSz         string `json:"lotSz"`
	MinSz         string `json:"minSz"`
	CtType        string `json:"ctType"`
	Alias         string `json:"alias,omitempty"`
	State         string `json:"state"`
}
