# yRPC

[![tag](https://img.shields.io/github/tag/sqos/yrpc.svg)](https://github.com/sqos/yrpc/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-%23007d9c)
[![GoDoc](https://godoc.org/github.com/sqos/yrpc?status.svg)](https://pkg.go.dev/github.com/sqos/yrpc)
![Build Status](https://github.com/sqos/yrpc/actions/workflows/go-ci.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/sqos/yrpc)](https://goreportcard.com/report/github.com/sqos/yrpc)
[![License](https://img.shields.io/github/license/sqos/yrpc)](./LICENSE)

## Declaration

yRPC is a fork of [eRPC](https://github.com/andeya/erpc) version 7.2.1. The version number of yRPC will start from v1.0.0. For using eRPC, please visit [the andeya's repository](https://github.com/andeya/erpc).

**We would like to thank the eRPC author and all related contributors for their efforts.**

## Introduction

yRPC is an efficient, extensible and easy-to-use RPC framework.

Suitable for RPC, Microservice, Peer-to-Peer, IM, Game and other fields.

[简体中文](https://github.com/sqos/yrpc/tree/main/README_ZH.md)


![yRPC-Framework](https://github.com/sqos/yrpc/raw/main/doc/yrpc_module_diagram.png)


## Install


- go vesion ≥ 1.18

- install
```sh
GO111MODULE=on go get -u -v -insecure github.com/sqos/yrpc
```

- import
```go
import "github.com/sqos/yrpc"
```

## Feature

- Use peer to provide the same API package for the server and client
- Provide multi-layout abstractions such as:
  - peer
  - session/socket
  - router
  - handle/context
  - message
  - protocol
  - codec
  - transfer filter
  - plugin
- Support reboot and shutdown gracefully
- HTTP-compatible message format:
  - Composed of two parts, the `Header` and the `Body`
  - `Header` contains metadata in the same format as HTTP header
  - `Body` supports for custom codec of Content Type-Like, already implemented:
    - Protobuf
    - Thrift
    - JSON
    - XML
    - Form
    - Plain
  - Support push, call-reply and more message types
- Support custom message protocol, and provide some common implementations:
  - `rawproto` - Default high performance binary protocol
  - `jsonproto` - JSON message protocol
  - `pbproto` - Ptotobuf message protocol
  - `thriftproto` - Thrift message protocol
  - `httproto` - HTTP message protocol
- Optimized high performance transport layer
  - Use Non-block socket and I/O multiplexing technology
  - Support setting the size of socket I/O buffer
  - Support setting the size of the reading message (if exceed disconnect it)
  - Support controling the connection file descriptor
- Support a variety of network types:
  - `tcp`
  - `tcp4`
  - `tcp6`
  - `unix`
  - `unixpacket`
  - `kcp`
  - `quic`
  - other
    - websocket
    - evio
- Provide a rich plug-in point, and already implemented:
  - auth
  - binder
  - heartbeat
  - ignorecase(service method)
  - overloader
  - proxy(for unknown service method)
  - secure
- Powerful and flexible logging system:
  - Detailed log information, support print input and output details
  - Support setting slow operation alarm threshold
  - Support for custom implementation log component
- Client session support automatically redials after disconnection


## Benchmark

- Profile torch of yrpc/socket

![yrpc_socket_profile_torch](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_profile_torch.png)

**[svg file](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_profile_torch.svg)**

- Heap torch of yrpc/socket

![yrpc_socket_heap_torch](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_heap_torch.png)

**[svg file](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_heap_torch.svg)**


## Example

### server.go

```go
package main

import (
	"fmt"
	"time"

	"github.com/sqos/yrpc"
)

func main() {
	defer yrpc.FlushLogger()
	// graceful
	go yrpc.GraceSignal()

	// server peer
	srv := yrpc.NewPeer(yrpc.PeerConfig{
		CountTime:   true,
		ListenPort:  9090,
		PrintDetail: true,
	})
	// srv.SetTLSConfig(yrpc.GenerateTLSConfigForServer())

	// router
	srv.RouteCall(new(Math))

	// broadcast per 5s
	go func() {
		for {
			time.Sleep(time.Second * 5)
			srv.RangeSession(func(sess yrpc.Session) bool {
				sess.Push(
					"/push/status",
					fmt.Sprintf("this is a broadcast, server time: %v", time.Now()),
				)
				return true
			})
		}
	}()

	// listen and serve
	srv.ListenAndServe()
}

// Math handler
type Math struct {
	yrpc.CallCtx
}

// Add handles addition request
func (m *Math) Add(arg *[]int) (int, *yrpc.Status) {
	// test meta
	yrpc.Infof("author: %s", m.PeekMeta("author"))
	// add
	var r int
	for _, a := range *arg {
		r += a
	}
	// response
	return r, nil
}
```

### client.go

```go
package main

import (
  "time"

  "github.com/sqos/yrpc"
)

func main() {
  defer yrpc.SetLoggerLevel("ERROR")()

  cli := yrpc.NewPeer(yrpc.PeerConfig{})
  defer cli.Close()
  // cli.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})

  cli.RoutePush(new(Push))

  sess, stat := cli.Dial(":9090")
  if !stat.OK() {
    yrpc.Fatalf("%v", stat)
  }

  var result int
  stat = sess.Call("/math/add",
    []int{1, 2, 3, 4, 5},
    &result,
    yrpc.WithAddMeta("author", "andeya"),
  ).Status()
  if !stat.OK() {
    yrpc.Fatalf("%v", stat)
  }
  yrpc.Printf("result: %d", result)
  yrpc.Printf("Wait 10 seconds to receive the push...")
  time.Sleep(time.Second * 10)
}

// Push push handler
type Push struct {
  yrpc.PushCtx
}

// Push handles '/push/status' message
func (p *Push) Status(arg *string) *yrpc.Status {
  yrpc.Printf("%s", *arg)
  return nil
}
```

[More Examples](https://github.com/sqos/yrpc/tree/main/examples)

## Usage

**NOTE:**

- It is best to set the packet size when reading: `SetReadLimit`
- The default packet size limit when reading is 1 GB

### Peer(server or client) Demo

```go
// Start a server
var peer1 = yrpc.NewPeer(yrpc.PeerConfig{
ListenPort: 9090, // for server role
})
peer1.Listen()

...

// Start a client
var peer2 = yrpc.NewPeer(yrpc.PeerConfig{})
var sess, err = peer2.Dial("127.0.0.1:8080")
```

### Call-Struct API template

```go
type Aaa struct {
    yrpc.CallCtx
}
func (x *Aaa) XxZz(arg *<T>) (<T>, *yrpc.Status) {
    ...
    return r, nil
}
```

- register it to root router:

```go
// register the call route
// HTTP mapping: /aaa/xx_zz
// RPC mapping: Aaa.XxZz
peer.RouteCall(new(Aaa))

// or register the call route
// HTTP mapping: /xx_zz
// RPC mapping: XxZz
peer.RouteCallFunc((*Aaa).XxZz)
```

### Service method mapping

- The default mapping(HTTPServiceMethodMapper) of struct(func) name to service methods:
    - `AaBb` -> `/aa_bb`
    - `ABcXYz` -> `/abc_xyz`
    - `Aa__Bb` -> `/aa_bb`
    - `aa__bb` -> `/aa_bb`
    - `ABC__XYZ` -> `/abc_xyz`
    - `Aa_Bb` -> `/aa/bb`
    - `aa_bb` -> `/aa/bb`
    - `ABC_XYZ` -> `/abc/xyz`
    ```go
    yrpc.SetServiceMethodMapper(yrpc.HTTPServiceMethodMapper)
    ```

- The mapping(RPCServiceMethodMapper) of struct(func) name to service methods:
    - `AaBb` -> `AaBb`
    - `ABcXYz` -> `ABcXYz`
    - `Aa__Bb` -> `Aa_Bb`
    - `aa__bb` -> `aa_bb`
    - `ABC__XYZ` -> `ABC_XYZ`
    - `Aa_Bb` -> `Aa.Bb`
    - `aa_bb` -> `aa.bb`
    - `ABC_XYZ` -> `ABC.XYZ`
    ```go
    yrpc.SetServiceMethodMapper(yrpc.RPCServiceMethodMapper)
    ```

### Call-Function API template

```go
func XxZz(ctx yrpc.CallCtx, arg *<T>) (<T>, *yrpc.Status) {
    ...
    return r, nil
}
```

- register it to root router:

```go
// register the call route
// HTTP mapping: /xx_zz
// RPC mapping: XxZz
peer.RouteCallFunc(XxZz)
```

### Push-Struct API template

```go
type Bbb struct {
    yrpc.PushCtx
}
func (b *Bbb) YyZz(arg *<T>) *yrpc.Status {
    ...
    return nil
}
```

- register it to root router:

```go
// register the push handler
// HTTP mapping: /bbb/yy_zz
// RPC mapping: Bbb.YyZz
peer.RoutePush(new(Bbb))

// or register the push handler
// HTTP mapping: /yy_zz
// RPC mapping: YyZz
peer.RoutePushFunc((*Bbb).YyZz)
```

### Push-Function API template

```go
// YyZz register the handler
func YyZz(ctx yrpc.PushCtx, arg *<T>) *yrpc.Status {
    ...
    return nil
}
```

- register it to root router:

```go
// register the push handler
// HTTP mapping: /yy_zz
// RPC mapping: YyZz
peer.RoutePushFunc(YyZz)
```

### Unknown-Call-Function API template

```go
func XxxUnknownCall (ctx yrpc.UnknownCallCtx) (interface{}, *yrpc.Status) {
    ...
    return r, nil
}
```

- register it to root router:

```go
// register the unknown call route: /*
peer.SetUnknownCall(XxxUnknownCall)
```

### Unknown-Push-Function API template

```go
func XxxUnknownPush(ctx yrpc.UnknownPushCtx) *yrpc.Status {
    ...
    return nil
}
```

- register it to root router:

```go
// register the unknown push route: /*
peer.SetUnknownPush(XxxUnknownPush)
```

### Plugin Demo

```go
// NewIgnoreCase Returns a ignoreCase plugin.
func NewIgnoreCase() *ignoreCase {
    return &ignoreCase{}
}

type ignoreCase struct{}

var (
    _ yrpc.PostReadCallHeaderPlugin = new(ignoreCase)
    _ yrpc.PostReadPushHeaderPlugin = new(ignoreCase)
)

func (i *ignoreCase) Name() string {
    return "ignoreCase"
}

func (i *ignoreCase) PostReadCallHeader(ctx yrpc.ReadCtx) *yrpc.Status {
    // Dynamic transformation path is lowercase
    ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
    return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx yrpc.ReadCtx) *yrpc.Status {
    // Dynamic transformation path is lowercase
    ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
    return nil
}
```

### Register above handler and plugin

```go
// add router group
group := peer.SubRoute("test")
// register to test group
group.RouteCall(new(Aaa), NewIgnoreCase())
peer.RouteCallFunc(XxZz, NewIgnoreCase())
group.RoutePush(new(Bbb))
peer.RoutePushFunc(YyZz)
peer.SetUnknownCall(XxxUnknownCall)
peer.SetUnknownPush(XxxUnknownPush)
```

### Config

```go
type PeerConfig struct {
    Network            string        `yaml:"network"              ini:"network"              comment:"Network; tcp, tcp4, tcp6, unix, unixpacket, kcp or quic"`
    LocalIP            string        `yaml:"local_ip"             ini:"local_ip"             comment:"Local IP"`
    ListenPort         uint16        `yaml:"listen_port"          ini:"listen_port"          comment:"Listen port; for server role"`
    DialTimeout time.Duration `yaml:"dial_timeout" ini:"dial_timeout" comment:"Default maximum duration for dialing; for client role; ns,µs,ms,s,m,h"`
    RedialTimes        int32         `yaml:"redial_times"         ini:"redial_times"         comment:"The maximum times of attempts to redial, after the connection has been unexpectedly broken; Unlimited when <0; for client role"`
	RedialInterval     time.Duration `yaml:"redial_interval"      ini:"redial_interval"      comment:"Interval of redialing each time, default 100ms; for client role; ns,µs,ms,s,m,h"`
    DefaultBodyCodec   string        `yaml:"default_body_codec"   ini:"default_body_codec"   comment:"Default body codec type id"`
    DefaultSessionAge  time.Duration `yaml:"default_session_age"  ini:"default_session_age"  comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
    DefaultContextAge  time.Duration `yaml:"default_context_age"  ini:"default_context_age"  comment:"Default CALL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
    SlowCometDuration  time.Duration `yaml:"slow_comet_duration"  ini:"slow_comet_duration"  comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
    PrintDetail        bool          `yaml:"print_detail"         ini:"print_detail"         comment:"Is print body and metadata or not"`
    CountTime          bool          `yaml:"count_time"           ini:"count_time"           comment:"Is count cost time or not"`
}
```

### Optimize

- SetMessageSizeLimit sets max message size.
  If maxSize<=0, set it to max uint32.

    ```go
    func SetMessageSizeLimit(maxMessageSize uint32)
    ```

- SetSocketKeepAlive sets whether the operating system should send
  keepalive messages on the connection.

    ```go
    func SetSocketKeepAlive(keepalive bool)
    ```

- SetSocketKeepAlivePeriod sets period between keep alives.

    ```go
    func SetSocketKeepAlivePeriod(d time.Duration)
    ```

- SetSocketNoDelay controls whether the operating system should delay
  message transmission in hopes of sending fewer messages (Nagle's
  algorithm).  The default is true (no delay), meaning that data is
  sent as soon as possible after a Write.

    ```go
    func SetSocketNoDelay(_noDelay bool)
    ```

- SetSocketReadBuffer sets the size of the operating system's
  receive buffer associated with the connection.

    ```go
    func SetSocketReadBuffer(bytes int)
    ```

- SetSocketWriteBuffer sets the size of the operating system's
  transmit buffer associated with the connection.

    ```go
    func SetSocketWriteBuffer(bytes int)
    ```


## Extensions

### Codec

| package                                  | import                                   | description                  |
| ---------------------------------------- | ---------------------------------------- | ---------------------------- |
| [json](https://github.com/sqos/yrpc/blob/main/codec/json_codec.go) | `"github.com/sqos/yrpc/codec"` | JSON codec(yrpc own)     |
| [protobuf](https://github.com/sqos/yrpc/blob/main/codec/protobuf_codec.go) | `"github.com/sqos/yrpc/codec"` | Protobuf codec(yrpc own) |
| [thrift](https://github.com/sqos/yrpc/blob/main/codec/thrift_codec.go) | `"github.com/sqos/yrpc/codec"` | Form(url encode) codec(yrpc own)   |
| [xml](https://github.com/sqos/yrpc/blob/main/codec/xml_codec.go) | `"github.com/sqos/yrpc/codec"` | Form(url encode) codec(yrpc own)   |
| [plain](https://github.com/sqos/yrpc/blob/main/codec/plain_codec.go) | `"github.com/sqos/yrpc/codec"` | Plain text codec(yrpc own)   |
| [form](https://github.com/sqos/yrpc/blob/main/codec/form_codec.go) | `"github.com/sqos/yrpc/codec"` | Form(url encode) codec(yrpc own)   |

### Plugin

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [auth](https://github.com/sqos/yrpc/tree/main/plugin/auth) | `"github.com/sqos/yrpc/plugin/auth"` | An auth plugin for verifying peer at the first time |
| [binder](https://github.com/sqos/yrpc/tree/main/plugin/binder) | `"github.com/sqos/yrpc/plugin/binder"` | Parameter Binding Verification for Struct Handler |
| [heartbeat](https://github.com/sqos/yrpc/tree/main/plugin/heartbeat) | `"github.com/sqos/yrpc/plugin/heartbeat"` | A generic timing heartbeat plugin        |
| [proxy](https://github.com/sqos/yrpc/tree/main/plugin/proxy) | `"github.com/sqos/yrpc/plugin/proxy"` | A proxy plugin for handling unknown calling or pushing |
[secure](https://github.com/sqos/yrpc/tree/main/plugin/secure)|`"github.com/sqos/yrpc/plugin/secure"` | Encrypting/decrypting the message body
[overloader](https://github.com/sqos/yrpc/tree/main/plugin/overloader)|`"github.com/sqos/yrpc/plugin/overloader"` | A plugin to protect yrpc from overload

### Protocol

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [rawproto](https://github.com/sqos/yrpc/tree/main/proto/rawproto) | `"github.com/sqos/yrpc/proto/rawproto` | A fast socket communication protocol(yrpc default protocol) |
| [jsonproto](https://github.com/sqos/yrpc/tree/main/proto/jsonproto) | `"github.com/sqos/yrpc/proto/jsonproto"` | A JSON socket communication protocol     |
| [pbproto](https://github.com/sqos/yrpc/tree/main/proto/pbproto) | `"github.com/sqos/yrpc/proto/pbproto"` | A Protobuf socket communication protocol     |
| [thriftproto](https://github.com/sqos/yrpc/tree/main/proto/thriftproto) | `"github.com/sqos/yrpc/proto/thriftproto"` | A Thrift communication protocol     |
| [httproto](https://github.com/sqos/yrpc/tree/main/proto/httproto) | `"github.com/sqos/yrpc/proto/httproto"` | A HTTP style socket communication protocol     |

### Transfer-Filter

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [gzip](https://github.com/sqos/yrpc/tree/main/xfer/gzip) | `"github.com/sqos/yrpc/xfer/gzip"` | Gzip(yrpc own)                       |
| [md5](https://github.com/sqos/yrpc/tree/main/xfer/md5) | `"github.com/sqos/yrpc/xfer/md5"` | Provides a integrity check transfer filter |

### Mixer

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [multiclient](https://github.com/sqos/yrpc/tree/main/mixer/multiclient) | `"github.com/sqos/yrpc/mixer/multiclient"` | Higher throughput client connection pool when transferring large messages (such as downloading files) |
| [websocket](https://github.com/sqos/yrpc/tree/main/mixer/websocket) | `"github.com/sqos/yrpc/mixer/websocket"` | Makes the yRPC framework compatible with websocket protocol as specified in RFC 6455 |
| [evio](https://github.com/sqos/yrpc/tree/main/mixer/evio) | `"github.com/sqos/yrpc/mixer/evio"` | A fast event-loop networking framework that uses the yrpc API layer |

## Projects based on yRPC

## Business Users

## License

yRPC is under Apache v2 License. See the [LICENSE](https://github.com/sqos/yrpc/raw/main/LICENSE) file for the full license text
