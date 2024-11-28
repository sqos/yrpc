package yrpc_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

func panic_call(yrpc.CallCtx, *interface{}) (interface{}, *yrpc.Status) {
	panic("panic_call")
}

func panic_push(yrpc.PushCtx, *interface{}) *yrpc.Status {
	panic("panic_push")
}

//go:generate go test -v -c -o "${GOPACKAGE}" $GOFILE

func TestPanic(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}

	srv := yrpc.NewPeer(yrpc.PeerConfig{
		CountTime:  true,
		ListenPort: 9090,
	})
	srv.RouteCallFunc(panic_call)
	srv.RoutePushFunc(panic_push)
	go srv.ListenAndServe()

	time.Sleep(2 * time.Second)

	cli := yrpc.NewPeer(yrpc.PeerConfig{})
	defer cli.Close()
	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		t.Fatal(stat)
	}
	stat = sess.Call("/panic/call", nil, nil).Status()
	if stat.OK() {
		t.Fatalf("/panic/call: expect error!")
	}
	t.Logf("/panic/call error: %v", stat)
	stat = sess.Push("/panic/push", nil)
	if !stat.OK() {
		t.Fatalf("/panic/push: expect ok!")
	}
	t.Logf("/panic/push: ok")
}
