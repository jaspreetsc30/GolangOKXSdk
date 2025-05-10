package okx

import (
	"cadenza-market-connector-okx/pkg/go-okx-api/common"
	"cadenza-market-connector-okx/pkg/go-okx-api/models/ws"
	"cadenza-market-connector-okx/pkg/go-okx-api/models/ws/business"
	"cadenza-market-connector-okx/pkg/go-okx-api/models/ws/public"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
)

const (
	MaxReconnectAttempts = 5
	ConnectTimeout       = 10 * time.Second
)

type OKXWsClient struct {
	Public   *WSClient
	Private  *WSClient
	Business *WSClient
}

type WSClient struct {
	auth         common.Auth
	conn         *websocket.Conn
	ctx          context.Context
	cancel       context.CancelFunc
	endpointType string // "public", "private", or "business"
	debugMode    bool
	endpoint     string
	pingInterval time.Duration
	lastResponse time.Time
	emitter      *emission.Emitter
	mu           sync.Mutex
	connected    bool
	reconnecting bool
}

func NewOKXWsClient(auth common.Auth) *OKXWsClient {
	public := NewWSClient("public", auth)
	private := NewWSClient("private", auth)
	business := NewWSClient("business", auth)

	return &OKXWsClient{
		Public:   public,
		Private:  private,
		Business: business,
	}
}

func NewWSClient(endpointType string, auth common.Auth) *WSClient {
	ctx, cancel := context.WithCancel(context.Background())
	endpoint := determineEndpoint(endpointType, auth.DebugMode)

	client := &WSClient{
		auth:         auth,
		ctx:          ctx,
		cancel:       cancel,
		endpointType: endpointType,
		debugMode:    auth.DebugMode,
		endpoint:     endpoint,
		pingInterval: 20 * time.Second,
		emitter:      emission.NewEmitter(),
		connected:    false,
		reconnecting: false,
	}

	if err := client.Connect(); err != nil {
		log.Printf("Initial connection failed for %s: %v", endpointType, err)
	}

	return client
}

func determineEndpoint(endpointType string, debugMode bool) string {
	switch endpointType {
	case "private":
		if debugMode {
			return "wss://wspap.okx.com:8443/ws/v5/private?brokerId=9999"
		}
		return "wss://ws.okx.com:8443/ws/v5/private"
	case "business":
		if debugMode {
			return "wss://wspap.okx.com:8443/ws/v5/business?brokerId=9999"
		}
		return "wss://ws.okx.com:8443/ws/v5/business"
	case "public":
		fallthrough
	default:
		if debugMode {
			return "wss://wspap.okx.com:8443/ws/v5/public?brokerId=9999"
		}
		return "wss://ws.okx.com:8443/ws/v5/public"
	}
}

func (c *OKXWsClient) Subscribe(args interface{}) error {
	requestArgs, ok := args.([]ws.Args)
	if !ok || len(requestArgs) == 0 {
		return errors.New("invalid subscription args: must be non-empty []ws.Args")
	}

	publicArgs := []ws.Args{}
	privateArgs := []ws.Args{}
	businessNoAuthArgs := []ws.Args{}
	businessAuthArgs := []ws.Args{}

	for _, arg := range requestArgs {
		channel := arg.Channel
		if isPrivateChannel(channel) {
			privateArgs = append(privateArgs, arg)
		} else if isBusinessChannel(channel) {
			if requiresBusinessAuth(channel) {
				businessAuthArgs = append(businessAuthArgs, arg)
			} else {
				businessNoAuthArgs = append(businessNoAuthArgs, arg)
			}
		} else {
			publicArgs = append(publicArgs, arg)
		}
	}

	fmt.Printf(" publicArgs %v\n", publicArgs)
	fmt.Printf(" privateArgs %v\n", privateArgs)
	fmt.Printf(" businessNoAuthArgs %v\n", businessNoAuthArgs)
	fmt.Printf(" businessAuthArgs %v\n", businessAuthArgs)

	var errs []error
	if len(publicArgs) > 0 {
		if err := c.Public.Subscribe(publicArgs); err != nil {
			errs = append(errs, fmt.Errorf("public subscribe failed: %v", err))
		}
	}
	if len(privateArgs) > 0 {
		if err := c.Private.Subscribe(privateArgs); err != nil {
			errs = append(errs, fmt.Errorf("private subscribe failed: %v", err))
		}
	}
	if len(businessNoAuthArgs) > 0 {
		if err := c.Business.Subscribe(businessNoAuthArgs); err != nil {
			errs = append(errs, fmt.Errorf("business no-auth subscribe failed: %v", err))
		}
	}
	if len(businessAuthArgs) > 0 {
		fmt.Println("code arrives here in businessAuthArgs")
		if err := c.Business.SubscribeWithAuth(businessAuthArgs); err != nil {
			errs = append(errs, fmt.Errorf("business with-auth subscribe failed: %v", err))
		}
	}

	if len(errs) > 0 {
		errMsg := "subscription errors: "
		for i, err := range errs {
			errMsg += err.Error()
			if i < len(errs)-1 {
				errMsg += "; "
			}
		}
		return errors.New(errMsg)
	}
	return nil
}

func (client *WSClient) SubscribeWithAuth(args interface{}) error {
	if err := client.Login(); err != nil {
		return err
	}
	return client.sendRequest(ws.NewRequestSubscribe(args))
}

func (client *WSClient) Connect() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	u, err := url.Parse(client.endpoint)
	if err != nil {
		return err
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	client.conn = c
	client.lastResponse = time.Now()
	client.connected = true
	client.reconnecting = false

	go client.handleMessages()
	go client.manageHeartbeat()

	return nil
}

func (client *WSClient) manageHeartbeat() {
	ticker := time.NewTicker(client.pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := client.sendPing(); err != nil {
				log.Printf("Ping failed for %s: %v", client.endpointType, err)
				client.reconnect()
			}
		case <-client.ctx.Done():
			return
		}
	}
}

func (client *WSClient) sendPing() error {
	return client.sendTextMessage("ping")
}

func (client *WSClient) sendTextMessage(message string) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if !client.connected {
		return errors.New("client not connected")
	}

	if err := client.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		client.connected = false
		return err
	}
	return nil
}

func (client *WSClient) reconnect() {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.cancel()
	ctx, cancel := context.WithCancel(context.Background())
	client.ctx = ctx
	client.cancel = cancel

	attempts := 0
	for attempts < MaxReconnectAttempts {
		if err := client.Connect(); err != nil {
			attempts++
			if attempts == MaxReconnectAttempts {
				log.Printf("Max reconnection attempts reached for %s: %v", client.endpointType, err)
				return
			}
			time.Sleep(2 * time.Second)
			continue
		}

		client.connected = true
		client.reconnecting = false

		if client.endpointType == "private" {
			if err := client.Login(); err != nil {
				log.Printf("Login after reconnect failed for %s: %v", client.endpointType, err)
				continue
			}
		}
		log.Printf("Successfully reconnected after %d attempts for %s", attempts+1, client.endpointType)
		return
	}
}

func (client *WSClient) Login() error {
	loginArgs := ws.NewRequestLogin(client.auth)
	return client.sendRequest(loginArgs)
}

func (client *WSClient) Subscribe(args interface{}) error {
	if client.endpointType == "private" {
		if err := client.Login(); err != nil {
			return err
		}
	}
	return client.sendRequest(ws.NewRequestSubscribe(args))
}

func (client *WSClient) Unsubscribe(args interface{}) error {
	if client.endpointType == "private" {
		if err := client.Login(); err != nil {
			return err
		}
	}
	return client.sendRequest(ws.NewRequestUnsubscribe(args))
}

func (client *WSClient) sendRequest(request *ws.Request) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if !client.connected {
		return errors.New("client not connected")
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshalling request: %v", err)
	}
	if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		client.connected = false
		return fmt.Errorf("sending request: %v", err)
	}
	return nil
}

func (client *WSClient) handleMessages() {
	defer client.conn.Close()
	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			_, message, err := client.conn.ReadMessage()
			if err != nil {
				log.Printf("Read error for %s: %v", client.endpointType, err)
				client.reconnect()
				return
			}

			client.lastResponse = time.Now()

			if string(message) == "ping" || string(message) == "pong" {
				continue
			}

			var response map[string]interface{}
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Failed to unmarshal response for %s: %v, message: %s", client.endpointType, err, message)
				client.Emit("raw_message", message)
				continue
			}

			switch {
			case response["event"] == "login":
				fmt.Printf("login response: %v\n", response)
			case response["event"] == "subscribe":
				client.Emit("subscribe", response)
			case response["event"] == "error":
				client.Emit("error", response)
			default:
				if arg, ok := response["arg"].(map[string]interface{}); ok {
					channel, channelOk := arg["channel"].(string)
					instId, instIdOk := arg["instId"].(string)
					if channelOk && instIdOk {
						eventKey := ws.Args{
							Channel: channel,
							InstId:  instId,
						}
						switch channel {
						case "tickers":
							var ticker public.TickerEvent
							if err := json.Unmarshal(message, &ticker); err == nil {
								client.Emit(eventKey, &ticker)
							} else {
								log.Printf("Failed to unmarshal ticker: %v", err)
								client.Emit(eventKey, message)
							}
						case "trades":
							var trade public.TradeEvent
							if err := json.Unmarshal(message, &trade); err == nil {
								client.Emit(eventKey, &trade)
							} else {
								log.Printf("Failed to unmarshal trade: %v", err)
								client.Emit(eventKey, message)
							}
						case "candle1m":
							var candle business.CandleEvent
							if err := json.Unmarshal(message, &candle); err == nil {
								client.Emit(eventKey, &candle)
							} else {
								log.Printf("Failed to unmarshal candle1m: %v", err)
								client.Emit(eventKey, message)
							}
						case "books5":
							var orderBook public.OrderBookEvent
							if err := json.Unmarshal(message, &orderBook); err == nil {
								client.Emit(eventKey, &orderBook)
							} else {
								log.Printf("Failed to unmarshal books5: %v", err)
								client.Emit(eventKey, message)
							}
						default:
							client.Emit(eventKey, message)
						}
					} else {
						log.Printf("Missing channel or instId in arg for %s: %v", client.endpointType, arg)
						client.Emit("message", message)
					}
				} else {
					client.Emit("message", message)
				}
			}

			if client.debugMode {
				log.Printf("Received for %s: %s", client.endpointType, message)
			}
		}
	}
}

func (client *WSClient) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.cancel()
	if client.conn != nil {
		err := client.conn.Close()
		client.connected = false
		return err
	}
	return nil
}

func (c *OKXWsClient) On(event interface{}, listener interface{}) *emission.Emitter {
	wsArgs, ok := event.(ws.Args)
	if !ok {
		log.Printf("Warning: Event must be of args type, got %T", event)
	}

	if isPrivateChannel(wsArgs.Channel) {
		return c.Private.On(event, listener)
	} else if isBusinessChannel(wsArgs.Channel) {
		return c.Business.On(event, listener)
	}
	return c.Public.On(event, listener)
}

func (c *OKXWsClient) Emit(event interface{}, arguments ...interface{}) *emission.Emitter {
	wsArgs, ok := event.(ws.Args)
	if !ok {
		log.Printf("Warning: Event must be of args type, got %T", event)
	}
	if isPrivateChannel(wsArgs.Channel) {
		return c.Private.Emit(event, arguments...)
	} else if isBusinessChannel(wsArgs.Channel) {
		return c.Business.Emit(event, arguments...)
	}
	return c.Public.Emit(event, arguments...)
}

func (c *OKXWsClient) Off(event interface{}, listener interface{}) *emission.Emitter {
	wsArgs, ok := event.(ws.Args)
	if !ok {
		log.Printf("Warning: Event must be of args type, got %T", event)
	}

	if isPrivateChannel(wsArgs.Channel) {
		return c.Private.Off(event, listener)
	} else if isBusinessChannel(wsArgs.Channel) {
		return c.Business.Off(event, listener)
	}
	return c.Public.Off(event, listener)
}

func isPrivateChannel(channel string) bool {
	privateChannels := []string{
		"orders",
		"account",
		"positions",
		"balance_and_position",
		"order-algo",
	}
	eventLower := strings.ToLower(channel)
	for _, ch := range privateChannels {
		if strings.Contains(eventLower, ch) {
			return true
		}
	}
	return false
}

func isBusinessChannel(channel string) bool {
	businessChannels := []string{
		"candle1m",
	}
	eventLower := strings.ToLower(channel)
	for _, ch := range businessChannels {
		if strings.Contains(eventLower, ch) {
			return true
		}
	}
	return false
}

func requiresBusinessAuth(channel string) bool {
	privateBusinessChannels := []string{
		// Add known private /business channels here
	}
	eventLower := strings.ToLower(channel)
	for _, ch := range privateBusinessChannels {
		if strings.Contains(eventLower, ch) {
			return true
		}
	}
	return false
}

func (client *WSClient) On(event interface{}, listener interface{}) *emission.Emitter {
	return client.emitter.On(event, listener)
}

func (client *WSClient) Emit(event interface{}, arguments ...interface{}) *emission.Emitter {
	return client.emitter.Emit(event, arguments...)
}

func (client *WSClient) Off(event interface{}, listener interface{}) *emission.Emitter {
	return client.emitter.Off(event, listener)
}
