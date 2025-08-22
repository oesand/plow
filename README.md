# Plow

[![tag](https://img.shields.io/github/tag/oesand/giglet.svg)](https://github.com/oesand/giglet/releases)
![Test Status](https://github.com/oesand/plow/actions/workflows/test.yml/badge.svg)
[![GoDoc](https://godoc.org/github.com/oesand/giglet?status.svg)](https://pkg.go.dev/github.com/oesand/giglet)
[![License](https://img.shields.io/github/license/oesand/giglet)](./LICENSE)

🐦‍⬛ **`plow` - All in One HTTP package for Go. Tuned for high performance, easy to use and speed up writing.**

## 📦 Installation

```sh
go get github.com/oesand/plow
```

## 🎯 Example

### Client

```go
import (
    "github.com/oesand/plow"
    "github.com/oesand/plow/specs"
)

url := specs.MustParseUrl("http://example.com")
req := plow.TextRequest(specs.HttpMethodPost, url, specs.ContentTypePlain, "Hello")

client := plow.Client{}
// Recommended (with optimal parameters)
// client := plow.DefaultClient()

resp, err := client.Make(req)
if err != nil {
    panic(err)
}
```

```go
client := plow.Client{}

// You can use proxy
transport := &plow.Transport{}
transport.Proxy = func(url *specs.Url) (*specs.Url, error) {
    return specs.ParseUrl("socks5://127.0.0.1:1080")
}

// Fixed proxy url
transport.Proxy = plow.FixedProxyUrl(specs.MustParseUrl("https://127.0.0.1"))
client.Transport = transport
```

### Server
```go
import (
    "github.com/oesand/plow"
    "github.com/oesand/plow/specs"
)

handler := plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
    return plow.TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "hello", func(resp plow.Response) {
        resp.Header().Set("X-Hello", "Value")
        resp.Header().SetCookieValue("Name", "Value")
    })
})

server := plow.Server{
    Handler: handler,
}

// Recommended (with optimal parameters)
// server := plow.DefaultServer(handler)

err := server.ListenAndServe(":http")
if err != nil {
    panic(err)
}
```

