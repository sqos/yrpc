# yRPC

[![tag](https://img.shields.io/github/tag/sqos/yrpc.svg)](https://github.com/sqos/yrpc/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-%23007d9c)
[![GoDoc](https://godoc.org/github.com/sqos/yrpc?status.svg)](https://pkg.go.dev/github.com/sqos/yrpc)
![Build Status](https://github.com/sqos/yrpc/actions/workflows/go-ci.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/sqos)](https://goreportcard.com/report/github.com/sqos/yrpc)
[![License](https://img.shields.io/github/license/sqos/yrpc)](./LICENSE)

## 声明

yRPC是基于[eRPC](https://github.com/andeya/erpc) v7.2.1版本的一个fork版本，yRPC版本号将从v1.0.0开始，如需使用eRPC请移步[原始作者andeya的仓库](https://github.com/andeya/erpc)。

**感谢eRPC作者和相关贡献者的付出。**

## 介绍

yRPC 是一个高效、可扩展且简单易用的 RPC 框架。

适用于 RPC、微服务、点对点长连接、IM 和游戏等领域。


![yRPC-Framework](https://github.com/sqos/yrpc/raw/main/doc/yrpc_module_diagram.png)


## 安装

- go vesion ≥ 1.18

- install
```sh
GO111MODULE=on go get -u -v -insecure github.com/sqos/yrpc
```

- import
```go
import "github.com/sqos/yrpc"
```

## 特性

- 使用 peer 为 server 和 client 提供相同的 API 封装
- 提供多层抽象，如：
  - peer
  - session/socket
  - router
  - handle/context
  - message
  - protocol
  - codec
  - transfer filter
  - plugin
- 支持平滑重启和关闭
- 兼容 HTTP 的消息格式：
  - 由 `Header` 和 `Body` 两部分组成
  - `Header` 包含与 HTTP header 格式相同的 metadata
  - `Body` 支持类似 Content Type 的自定义编解码器，已经实现的：
    - Protobuf
    - Thrift
    - JSON
    - XML
    - Form
    - Plain
  - 支持 push、call-reply 和更多的消息类型
- 支持自定义消息协议，并提供了一些常见实现：
  - `rawproto` - 默认的高性能二进制协议
  - `jsonproto` - JSON 消息协议
  - `pbproto` - Ptotobuf 消息协议
  - `thriftproto` - Thrift 消息协议
  - `httproto` - HTTP 消息协议
- 可优化的高性能传输层
  - 使用 Non-block socket 和 I/O 多路复用技术
  - 支持设置套接字 I/O 的缓冲区大小
  - 支持设置读取消息的大小（如果超过则断开连接）
  - 支持控制连接的文件描述符
- 支持多种网络类型：
  - `tcp`
  - `tcp4`
  - `tcp6`
  - `unix`
  - `unixpacket`
  - `kcp`
  - `quic`
  - 其他
    - websocket
    - evio
- 提供丰富的插件埋点，并已实现：
  - auth
  - binder
  - heartbeat
  - ignorecase(service method)
  - overloader
  - proxy(for unknown service method)
  - secure
- 强大灵活的日志系统：
  - 详细的日志信息，支持打印输入和输出详细信息
  - 支持设置慢操作警报阈值
  - 支持自定义实现日志组件
- 客户端会话支持在断开连接后自动重拨


## 性能测试


- CPU耗时火焰图 yrpc/socket

![yrpc_socket_profile_torch](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_profile_torch.png)

**[svg file](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_profile_torch.svg)**

- 堆栈信息火焰图 yrpc/socket

![yrpc_socket_heap_torch](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_heap_torch.png)

**[svg file](https://github.com/sqos/yrpc/raw/main/doc/yrpc_socket_heap_torch.svg)**


## 代码示例

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

[更多示例](https://github.com/sqos/yrpc/blob/main/examples)

## 用法

**NOTE:**

- 最好设置读包时大小限制: `SetReadLimit`
- 默认读包时大小限制为 1 GB

### Peer端点（服务端或客户端）示例

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

### 自带ServiceMethod映射规则

- 结构体或方法名称到服务方法名称的默认映射（HTTPServiceMethodMapper）：
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

- 结构体或方法名称到服务方法名称的映射（RPCServiceMethodMapper）：
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

### Call-Struct 接口模版

```go
type Aaa struct {
    yrpc.CallCtx
}
func (x *Aaa) XxZz(arg *<T>) (<T>, *yrpc.Status) {
    ...
    return r, nil
}
```

- 注册到根路由：

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

### Call-Function 接口模板

```go
func XxZz(ctx yrpc.CallCtx, arg *<T>) (<T>, *yrpc.Status) {
    ...
    return r, nil
}
```

- 注册到根路由：

```go
// register the call route
// HTTP mapping: /xx_zz
// RPC mapping: XxZz
peer.RouteCallFunc(XxZz)
```

### Push-Struct 接口模板

```go
type Bbb struct {
    yrpc.PushCtx
}
func (b *Bbb) YyZz(arg *<T>) *yrpc.Status {
    ...
    return nil
}
```

- 注册到根路由：

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

### Push-Function 接口模板

```go
// YyZz register the handler
func YyZz(ctx yrpc.PushCtx, arg *<T>) *yrpc.Status {
    ...
    return nil
}
```

- 注册到根路由：

```go
// register the push handler
// HTTP mapping: /yy_zz
// RPC mapping: YyZz
peer.RoutePushFunc(YyZz)
```

### Unknown-Call-Function 接口模板

```go
func XxxUnknownCall (ctx yrpc.UnknownCallCtx) (interface{}, *yrpc.Status) {
    ...
    return r, nil
}
```

- 注册到根路由：

```go
// register the unknown pull route: /*
peer.SetUnknownCall(XxxUnknownCall)
```

### Unknown-Push-Function 接口模板

```go
func XxxUnknownPush(ctx yrpc.UnknownPushCtx) *yrpc.Status {
    ...
    return nil
}
```

- 注册到根路由：

```go
// register the unknown push route: /*
peer.SetUnknownPush(XxxUnknownPush)
```

### 插件示例

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

### 注册以上操作和插件示例到路由

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

### 配置信息

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
    DefaultContextAge  time.Duration `yaml:"default_context_age"  ini:"default_context_age"  comment:"Default PULL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
    SlowCometDuration  time.Duration `yaml:"slow_comet_duration"  ini:"slow_comet_duration"  comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
    PrintDetail        bool          `yaml:"print_detail"         ini:"print_detail"         comment:"Is print body and metadata or not"`
    CountTime          bool          `yaml:"count_time"           ini:"count_time"           comment:"Is count cost time or not"`
}
```

### 通信优化

- SetMessageSizeLimit 设置报文大小的上限，
  如果 maxSize<=0，上限默认为最大 uint32

    ```go
    func SetMessageSizeLimit(maxMessageSize uint32)
    ```

- SetSocketKeepAlive 是否允许操作系统的发送TCP的keepalive探测包

    ```go
    func SetSocketKeepAlive(keepalive bool)
    ```


- SetSocketKeepAlivePeriod 设置操作系统的TCP发送keepalive探测包的频度

    ```go
    func SetSocketKeepAlivePeriod(d time.Duration)
    ```

- SetSocketNoDelay 是否禁用Nagle算法，禁用后将不在合并较小数据包进行批量发送，默认为禁用

    ```go
    func SetSocketNoDelay(_noDelay bool)
    ```

- SetSocketReadBuffer 设置操作系统的TCP读缓存区的大小

    ```go
    func SetSocketReadBuffer(bytes int)
    ```

- SetSocketWriteBuffer 设置操作系统的TCP写缓存区的大小

    ```go
    func SetSocketWriteBuffer(bytes int)
    ```


## 扩展包

### 编解码器
| package                                  | import                                   | description                  |
| ---------------------------------------- | ---------------------------------------- | ---------------------------- |
| [json](https://github.com/sqos/yrpc/blob/main/codec/json_codec.go) | `"github.com/sqos/yrpc/codec"` | JSON codec(yrpc own)     |
| [protobuf](https://github.com/sqos/yrpc/blob/main/codec/protobuf_codec.go) | `"github.com/sqos/yrpc/codec"` | Protobuf codec(yrpc own) |
| [thrift](https://github.com/sqos/yrpc/blob/main/codec/thrift_codec.go) | `"github.com/sqos/yrpc/codec"` | Form(url encode) codec(yrpc own)   |
| [xml](https://github.com/sqos/yrpc/blob/main/codec/xml_codec.go) | `"github.com/sqos/yrpc/codec"` | Form(url encode) codec(yrpc own)   |
| [plain](https://github.com/sqos/yrpc/blob/main/codec/plain_codec.go) | `"github.com/sqos/yrpc/codec"` | Plain text codec(yrpc own)   |
| [form](https://github.com/sqos/yrpc/blob/main/codec/form_codec.go) | `"github.com/sqos/yrpc/codec"` | Form(url encode) codec(yrpc own)   |

### 插件

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [auth](https://github.com/sqos/yrpc/tree/main/plugin/auth) | `"github.com/sqos/yrpc/plugin/auth"` | An auth plugin for verifying peer at the first time |
| [binder](https://github.com/sqos/yrpc/tree/main/plugin/binder) | `"github.com/sqos/yrpc/plugin/binder"` | Parameter Binding Verification for Struct Handler |
| [heartbeat](https://github.com/sqos/yrpc/tree/main/plugin/heartbeat) | `"github.com/sqos/yrpc/plugin/heartbeat"` | A generic timing heartbeat plugin        |
| [proxy](https://github.com/sqos/yrpc/tree/main/plugin/proxy) | `"github.com/sqos/yrpc/plugin/proxy"` | A proxy plugin for handling unknown calling or pushing |
[secure](https://github.com/sqos/yrpc/tree/main/plugin/secure)|`"github.com/sqos/yrpc/plugin/secure"` | Encrypting/decrypting the message body
[overloader](https://github.com/sqos/yrpc/tree/main/plugin/overloader)|`"github.com/sqos/yrpc/plugin/overloader"` | A plugin to protect yrpc from overload

### 协议

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [rawproto](https://github.com/sqos/yrpc/tree/main/proto/rawproto) | `"github.com/sqos/yrpc/proto/rawproto` | 一个高性能的通信协议（yrpc默认）|
| [jsonproto](https://github.com/sqos/yrpc/tree/main/proto/jsonproto) | `"github.com/sqos/yrpc/proto/jsonproto"` | JSON 格式的通信协议     |
| [pbproto](https://github.com/sqos/yrpc/tree/main/proto/pbproto) | `"github.com/sqos/yrpc/proto/pbproto"` | Protobuf 格式的通信协议     |
| [thriftproto](https://github.com/sqos/yrpc/tree/main/proto/thriftproto) | `"github.com/sqos/yrpc/proto/thriftproto"` | Thrift 格式的通信协议     |
| [httproto](https://github.com/sqos/yrpc/tree/main/proto/httproto) | `"github.com/sqos/yrpc/proto/httproto"` | HTTP 格式的通信协议     |

### 传输过滤器

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [gzip](https://github.com/sqos/yrpc/tree/main/xfer/gzip) | `"github.com/sqos/yrpc/xfer/gzip"` | Gzip(yrpc own)                       |
| [md5](https://github.com/sqos/yrpc/tree/main/xfer/md5) | `"github.com/sqos/yrpc/xfer/md5"` | Provides a integrity check transfer filter |

### 其他模块

| package                                  | import                                   | description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| [multiclient](https://github.com/sqos/yrpc/tree/main/mixer/multiclient) | `"github.com/sqos/yrpc/mixer/multiclient"` | Higher throughput client connection pool when transferring large messages (such as downloading files) |
| [websocket](https://github.com/sqos/yrpc/tree/main/mixer/websocket) | `"github.com/sqos/yrpc/mixer/websocket"` | Makes the yRPC framework compatible with websocket protocol as specified in RFC 6455 |
| [evio](https://github.com/sqos/yrpc/tree/main/mixer/evio) | `"github.com/sqos/yrpc/mixer/evio"` | A fast event-loop networking framework that uses the yrpc API layer |

## 基于yRPC的项目

## 企业用户

## 开源协议

yRPC 项目采用商业应用友好的 [Apache2.0](https://github.com/sqos/yrpc/raw/main/LICENSE) 协议发布
