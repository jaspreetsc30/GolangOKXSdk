# OKX Go Client

The `okx` package provides a Go-based client for interacting with the OKX cryptocurrency exchange API. It supports both REST and WebSocket connections, enabling you to fetch market data and subscribe to real-time updates.

## Features
- **REST Client**: Fetch market data using OKX's REST API (e.g., instruments, tickers).
- **WebSocket Client**: Subscribe to real-time market data streams, including:
  - Order books (`books5`)
  - Tickers (`tickers`)
  - Trades (`trades`)
  - Candlesticks (`candle1m`)
- **Authentication**: Supports API key authentication with OKX's passphrase for secure access.
- **Reconnection Logic**: Automatically reconnects WebSocket clients in case of disconnection, with configurable retry attempts.
- **Debug Mode**: Supports OKX's simulated trading environment for testing.

## Prerequisites
- Go 1.16 or later
- OKX API credentials (API Key, Secret Key, and Passphrase)
- Dependencies:
  - `github.com/valyala/fasthttp`
  - `github.com/gorilla/websocket`
  - `github.com/chuckpreslar/emission`
  - `github.com/google/go-querystring`

## Installation
1. Add the package to your Go project:
   ```bash
   go get github.com/your-org/okx
   ```
   Replace `github.com/your-org/okx` with the actual repository path where the `okx` package is hosted.
2. Install the required dependencies:
   ```bash
   go mod tidy
   ```

## Configuration
The package uses a `Configuration` struct to initialize the client with your OKX API credentials.

### Example Configuration

```go
package main

import (
    "github.com/your-org/okx"
)

func main() {
    config := &okx.Configuration{
        ApiKey:        "your-api-key",
        SecretKey:     "your-secret-key",
        OkxPassphrase: "your-passphrase",
        DebugMode:     false, // Set to true for simulated trading
        AutoReconnect: true,
    }

    client := okx.NewClient(config)
    // Use the client for REST or WebSocket operations
}
```

## Usage
The package provides two main components:
1. **REST Client**: For interacting with OKX's REST API.
2. **WebSocket Client**: For subscribing to real-time market data.

### 1. Using the REST Client
The REST client (`RestClient`) allows you to fetch market data, such as a list of instruments.

#### Example: Fetch Instruments

```go
package main

import (
    "fmt"
    "github.com/your-org/okx"
    "github.com/your-org/okx/models/rest/public"
)

func main() {
    config := &okx.Configuration{
        ApiKey:        "your-api-key",
        SecretKey:     "your-secret-key",
        OkxPassphrase: "your-passphrase",
        DebugMode:     false,
    }
    client := okx.NewClient(config)

    // Fetch SPOT instruments
    param := &public.GetInstrumentsParam{
        InstType: "SPOT",
    }
    req, resp := public.NewGetInstruments(param)
    if err := client.Rest.Do(req, resp); err != nil {
        fmt.Printf("Error fetching instruments: %v\n", err)
        return
    }

    instruments := resp.(*public.GetInstrumentsResponse).Data
    for _, inst := range instruments {
        fmt.Printf("Instrument: %s, Tick Size: %s\n", inst.InstId, inst.TickSz)
    }
}
```

### 2. Using the WebSocket Client
The WebSocket client (`OKXWsClient`) allows you to subscribe to real-time market data streams.

#### Example: Subscribe to Market Data

```go
package main

import (
    "fmt"
    "github.com/your-org/okx"
    "github.com/your-org/okx/models/ws"
    "github.com/your-org/okx/models/ws/public"
)

func main() {
    config := &okx.Configuration{
        ApiKey:        "your-api-key",
        SecretKey:     "your-secret-key",
        OkxPassphrase: "your-passphrase",
        DebugMode:     false,
        AutoReconnect: true,
    }
    client := okx.NewClient(config)

    // Subscribe to ticker updates for BTC-USDT
    args := []ws.Args{
        {
            Channel: "tickers",
            InstId:  "BTC-USDT",
        },
    }
    if err := client.Ws.Subscribe(args); err != nil {
        fmt.Printf("Failed to subscribe: %v\n", err)
        return
    }

    // Listen for ticker updates
    client.Ws.On(args[0], func(e *public.TickerEvent) {
        for _, ticker := range e.Data {
            fmt.Printf("Ticker Update for %s: Last Price: %s\n", ticker.InstId, ticker.Last)
        }
    })

    // Keep the application running to receive WebSocket messages
    select {}
}
```

#### Supported WebSocket Channels
The WebSocket client supports the following channels:
- **Order Book (`books5`)**: Real-time order book updates (top 5 levels).
- **Tickers (`tickers`)**: Real-time ticker data (price, volume, etc.).
- **Trades (`trades`)**: Real-time trade data.
- **Candlesticks (`candle1m`)**: 1-minute candlestick data.

### 3. Customizing Event Handling
You can attach custom listeners to WebSocket events using the `On` method. The `Emit` method is used internally to dispatch events to your listeners.

#### Example: Handle Order Book Updates

```go
client.Ws.On(ws.Args{Channel: "books5", InstId: "BTC-USDT"}, func(e *public.OrderBookEvent) {
    for _, orderBook := range e.Data {
        fmt.Printf("Order Book Update for %s: Asks: %v, Bids: %v\n", orderBook.InstId, orderBook.Asks, orderBook.Bids)
    }
})
```

## Debugging

- **Debug Mode**: Set `DebugMode: true` in the `Configuration` to use OKX's simulated trading environment. This is useful for testing without affecting real funds.
- **Logging**: The WebSocket client logs connection and subscription events. You can enable debug logging by setting `DebugMode: true`.

## Error Handling
- The WebSocket client automatically reconnects if the connection is lost, with a maximum of 5 retry attempts (`MaxReconnectAttempts`).
- REST API errors are returned as `rest.OKXError` with a code and message.

## Contributing
Contributions are welcome! Please submit a pull request with your changes or open an issue to discuss improvements.

## License
This project is licensed under the MIT License. See the `LICENSE` file for details.
