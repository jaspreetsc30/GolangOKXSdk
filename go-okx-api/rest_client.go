package okx

import (
	"encoding/json"
	"fmt"
	"time"
	"github.com/google/go-querystring/query"
	"github.com/valyala/fasthttp"
    "cadenza-market-connector-okx/pkg/go-okx-api/common"
	"cadenza-market-connector-okx/pkg/go-okx-api/models/rest"

)

var (
	DefaultFastHttpClient = &fasthttp.Client{
		Name:                "go-okx",
		MaxConnsPerHost:     16,
		MaxIdleConnDuration: 20 * time.Second,
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
	}
)

type RestClient struct {
	Host string
	Auth common.Auth
	C    *fasthttp.Client
}

// new *Client
func NewRestClient(host string, auth common.Auth, c *fasthttp.Client) *RestClient {
	if host == "" {
		host = "https://www.okx.com"
	}
	if c == nil {
		c = DefaultFastHttpClient
	}

	return &RestClient{
		Host: host,
		Auth: auth,
		C:    c,
	}
}

// do request
func (c *RestClient) Do(req rest.IRequest, resp rest.IResponse) error {
	data, err := c.do(req)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	if !resp.IsOk() {
		return rest.NewOKXError(resp.GetCode(), resp.GetMessage())
	}

	return nil
}

// do request
func (c *RestClient) do(r rest.IRequest) ([]byte, error) {
	req := c.newRequest(r)
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	if err := c.C.Do(req, resp); err != nil {
		return nil, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("http status code:%d, desc:%s", resp.StatusCode(), string(resp.Body()))
	}

	return resp.Body(), nil
}

// new *fasthttp.Request
func (c *RestClient) newRequest(r rest.IRequest) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	sign := c.newSignature(r)

	headers := map[string]string{
		fasthttp.HeaderContentType: "application/json;charset=utf-8",
		fasthttp.HeaderAccept:      "application/json",
		"OK-ACCESS-KEY":            c.Auth.ApiKey,
		"OK-ACCESS-PASSPHRASE":     c.Auth.Passphrase,
		"OK-ACCESS-SIGN":           sign.Build(),
		"OK-ACCESS-TIMESTAMP":      sign.Timestamp,
	}
	if c.Auth.DebugMode {
		headers["x-simulated-trading"] = "1"
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.SetMethod(sign.Method)

	req.SetRequestURI(c.Host + sign.Path)
	if sign.Body != "" {
		req.SetBodyString(sign.Body)
	}

	return req
}

// new *Signature
func (c *RestClient) newSignature(r rest.IRequest) *common.Signature {
	var body []byte
	path := r.GetPath()

	if r.IsPost() {
		body, _ = json.Marshal(r.GetParam())
	} else if values, _ := query.Values(r.GetParam()); len(values) > 0 {
		path += "?" + values.Encode()
	}

	return c.Auth.Signature(r.GetMethod(), path, string(body), false)
}
