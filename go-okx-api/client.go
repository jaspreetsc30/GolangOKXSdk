package okx

import (
	"cadenza-market-connector-okx/pkg/go-okx-api/common"
)





type Configuration struct {
	ApiKey string 
	SecretKey string 
	OkxPassphrase string
	AutoReconnect bool
	DebugMode bool
}

type Client struct {
	Configuration *Configuration
	Rest *RestClient
	Ws *OKXWsClient
}



func NewClient(configuration *Configuration) *Client{
	auth := common.NewAuth(configuration.ApiKey,configuration.SecretKey , configuration.OkxPassphrase , configuration.DebugMode)
	restClient := NewRestClient("", auth ,nil)
	wsClient := NewOKXWsClient(auth)



	return &Client{
		Configuration: configuration,
		Rest: restClient,
		Ws: wsClient,
	}

}

