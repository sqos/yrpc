package session_age

import (
	"context"
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}_client" $GOFILE

func TestClient(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}

	defer yrpc.SetLoggerLevel("INFO")()
	cli := yrpc.NewPeer(yrpc.PeerConfig{PrintDetail: true})
	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}

	var result string
	sess.Call("/test/ok", "test1", &result)
	yrpc.Infof("test sync1: %v", result)
	result = ""
	stat = sess.Call("/test/timeout", "test2", &result).Status()
	yrpc.Infof("test sync2: server context timeout: %v", stat)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

	result = ""
	callCmd := sess.AsyncCall(
		"/test/timeout",
		"test3",
		&result,
		make(chan yrpc.CallCmd, 1),
		yrpc.WithContext(ctx),
	)
	select {
	case <-callCmd.Done():
		cancel()
		yrpc.Infof("test async1: %v", result)
	case <-ctx.Done():
		yrpc.Warnf("test async1: client context timeout: %v", ctx.Err())
	}

	time.Sleep(time.Second * 6)
	result = ""
	stat = sess.Call("/test/ok", "test4", &result).Status()
	yrpc.Warnf("test sync3: disconnect due to server session timeout: %v", stat.Cause())

	sess, stat = cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	sess.AsyncCall(
		"/test/break",
		nil,
		nil,
		make(chan yrpc.CallCmd, 1),
	)
}
