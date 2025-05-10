package ws

import (
	"cadenza-market-connector-okx/pkg/go-okx-api/common"
)

type Request struct {
	Op   string      `json:"op"`
	Args interface{} `json:"args"`
}

// new request for subscribe
func NewRequestSubscribe(args interface{}) *Request {
	return &Request{
		Op:   "subscribe",
		Args: args,
	}}

func NewRequestUnsubscribe(args interface{}) *Request {
	return &Request{
		Op:   "unsubscribe",
		Args: args,
	}}


// // new request for login
func NewRequestLogin(auth common.Auth) *Request {
	return &Request{
		Op:   "login",
		Args: []LoginArgs{*NewLoginArgsFromAuth(auth)},
	}}

