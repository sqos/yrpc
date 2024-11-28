package tps

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

type Call struct {
	yrpc.CallCtx
}

func (*Call) Test(*struct{}) (*struct{}, *yrpc.Status) {
	return nil, nil
}

type Push struct {
	yrpc.PushCtx
}

func (*Push) Test(*struct{}) *yrpc.Status {
	return nil
}

//go:generate go test -v -c -o "${GOPACKAGE}" ./...

func TestTPS(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	yrpc.SetLoggerLevel("OFF")
	// Server
	srv := yrpc.NewPeer(yrpc.PeerConfig{ListenPort: 9090}, NewTPS(5))
	srv.RouteCall(new(Call))
	srv.RoutePush(new(Push))
	go srv.ListenAndServe()
	time.Sleep(1e9)

	// Client
	cli := yrpc.NewPeer(yrpc.PeerConfig{})
	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		t.Fatal(stat)
	}

	ticker := time.NewTicker(time.Millisecond * 10)
	for i := 0; i < 1<<10; i++ {
		<-ticker.C
		stat = sess.Call("/call/test", nil, nil).Status()
		if !stat.OK() {
			t.Fatal(stat)
		}
		stat = sess.Push("/push/test", nil)
		if !stat.OK() {
			t.Fatal(stat)
		}
	}
}
