package early_comm

import (
	"testing"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}_client" $GOFILE

func TestClient(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	defer yrpc.FlushLogger()
	cli := yrpc.NewPeer(
		yrpc.PeerConfig{
			PrintDetail: false,
		},
		new(earlyCall),
	)
	defer cli.Close()
	_, stat := cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
}

type earlyCall struct{}

func (e *earlyCall) Name() string {
	return "early_call"
}

func (e *earlyCall) PostDial(sess yrpc.PreSession, isRedial bool) *yrpc.Status {
	stat := sess.PreSend(
		yrpc.TypeCall,
		"/early/ping",
		map[string]string{
			"author": "andeya",
		},
		nil,
	)
	if !stat.OK() {
		return stat
	}

	input := sess.PreReceive(func(header yrpc.Header) interface{} {
		if header.ServiceMethod() == "/early/pong" {
			return new(string)
		}
		yrpc.Panicf("Received an unexpected response: %s", header.ServiceMethod())
		return nil
	})
	stat = input.Status()
	if !stat.OK() {
		return stat
	}
	yrpc.Infof("result: %v", input.String())
	return nil
}
