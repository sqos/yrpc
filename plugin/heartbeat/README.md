## heartbeat

A generic timing heartbeat plugin.

During a heartbeat, if there is no communication, send a heartbeat message;
When the connection is idle more than 3 times the heartbeat time, take the initiative to disconnect.

### Usage

`import "github.com/sqos/yrpc/plugin/heartbeat"`

#### Test

```go
package heartbeat_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/plugin/heartbeat"
)

func TestHeartbeatCall1(t *testing.T) {
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{ListenPort: 9090, PrintDetail: true},
		heartbeat.NewPong(),
	)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := yrpc.NewPeer(
		yrpc.PeerConfig{PrintDetail: true},
		heartbeat.NewPing(3, true),
	)
	cli.Dial(":9090")
	time.Sleep(time.Second * 10)
}

func TestHeartbeatCall2(t *testing.T) {
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{ListenPort: 9090, PrintDetail: true},
		heartbeat.NewPong(),
	)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := yrpc.NewPeer(
		yrpc.PeerConfig{PrintDetail: true},
		heartbeat.NewPing(3, true),
	)
	sess, _ := cli.Dial(":9090")
	for i := 0; i < 8; i++ {
		sess.Call("/", nil, nil)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second * 5)
}

func TestHeartbeatPush1(t *testing.T) {
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{ListenPort: 9090, PrintDetail: true},
		heartbeat.NewPing(3, false),
	)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := yrpc.NewPeer(
		yrpc.PeerConfig{PrintDetail: true},
		heartbeat.NewPong(),
	)
	cli.Dial(":9090")
	time.Sleep(time.Second * 10)
}

func TestHeartbeatPush2(t *testing.T) {
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{ListenPort: 9090, PrintDetail: true},
		heartbeat.NewPing(3, false),
	)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := yrpc.NewPeer(
		yrpc.PeerConfig{PrintDetail: true},
		heartbeat.NewPong(),
	)
	sess, _ := cli.Dial(":9090")
	for i := 0; i < 8; i++ {
		sess.Push("/", nil)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second * 5)
}
```

test command:

```sh
go test -v -run=TestHeartbeatCall1
go test -v -run=TestHeartbeatCall2
go test -v -run=TestHeartbeatPush1
go test -v -run=TestHeartbeatPush2
```