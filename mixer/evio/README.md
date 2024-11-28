## evio

A fast event-loop networking framework that uses the yrpc API layer. (From [evio](https://github.com/tidwall/evio))

It makes direct [epoll](https://en.wikipedia.org/wiki/Epoll) and [kqueue](https://en.wikipedia.org/wiki/Kqueue) syscalls rather than using the standard Go [net](https://golang.org/pkg/net/) package, and works in a similar manner as [libuv](https://github.com/libuv/libuv) and [libevent](https://github.com/libevent/libevent).

### Feature

- [Fast](#performance) single-threaded or [multithreaded](#multithreaded) event loop
- Built-in [load balancing](#load-balancing) options
- yRPC API
- Low memory usage
- Fallback for non-epoll/kqueue operating systems by simulating events with the [net](https://golang.org/pkg/net/) package
- [SO_REUSEPORT](#so_reuseport) socket option
- Support TLS
- (TODO) Support set connection deadline

### Usage
	
`import "github.com/sqos/yrpc/mixer/evio"`

#### Test

```go
package evio_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/mixer/evio"
)

func Test(t *testing.T) {
	// server
	srv := evio.NewServer(1, yrpc.PeerConfig{ListenPort: 9090})
	// use TLS
	srv.SetTLSConfig(yrpc.GenerateTLSConfigForServer())
	srv.RouteCall(new(Home))
	go srv.ListenAndServe()
	time.Sleep(1e9)

	// client
	cli := evio.NewClient(yrpc.PeerConfig{})
	// use TLS
	cli.SetTLSConfig(yrpc.GenerateTLSConfigForClient())
	cli.RoutePush(new(Push))
	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result interface{}
	stat = sess.Call("/home/test",
		map[string]string{
			"author": "andeya",
		},
		&result,
		yrpc.WithAddMeta("peer_id", "110"),
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("result:%v", result)
	time.Sleep(2e9)
}

type Home struct {
	yrpc.CallCtx
}

func (h *Home) Test(arg *map[string]string) (map[string]interface{}, *yrpc.Status) {
	h.Session().Push("/push/test", map[string]string{
		"your_id": string(h.PeekMeta("peer_id")),
	})
	return map[string]interface{}{
		"arg": *arg,
	}, nil
}

type Push struct {
	yrpc.PushCtx
}

func (p *Push) Test(arg *map[string]string) *yrpc.Status {
	yrpc.Infof("receive push(%s):\narg: %#v\n", p.IP(), arg)
	return nil
}
```

test command:

```sh
go test -v -run=Test
```