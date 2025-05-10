package ws

import "cadenza-market-connector-okx/pkg/go-okx-api/common"

type LoginArgs struct {
	ApiKey     string `json:"apiKey"`
	Passphrase string `json:"passphrase"`
	Timestamp  string `json:"timestamp"`
	Sign       string `json:"sign"`
}

func NewLoginArgsFromAuth(auth common.Auth) *LoginArgs {
	signature := auth.Signature("GET", "/users/self/verify", "", true)
	return &LoginArgs{
		ApiKey:     auth.ApiKey,
		Passphrase: auth.Passphrase,
		Sign:       signature.Build(),
		Timestamp:  signature.Timestamp,
	}
}