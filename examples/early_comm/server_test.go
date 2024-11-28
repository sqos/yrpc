package early_comm

import (
	"testing"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}_server" $GOFILE

func TestServer(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}

	defer yrpc.FlushLogger()
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{
			PrintDetail: false,
			ListenPort:  9090,
		},
		new(earlyResult),
	)
	srv.ListenAndServe()
}

type earlyResult struct{}

func (e *earlyResult) Name() string {
	return "early_result"
}

func (e *earlyResult) PostAccept(sess yrpc.PreSession) *yrpc.Status {
	var rigthServiceMethod bool
	input := sess.PreReceive(func(header yrpc.Header) interface{} {
		if header.ServiceMethod() == "/early/ping" {
			rigthServiceMethod = true
			return new(map[string]string)
		}
		return nil
	})
	stat := input.Status()
	if !stat.OK() {
		return stat
	}

	var result string
	if !rigthServiceMethod {
		stat = yrpc.NewStatus(10005, "unexpected request", "")
	} else {
		body := *input.Body().(*map[string]string)
		if body["author"] != "andeya" {
			stat = yrpc.NewStatus(10005, "incorrect author", body["author"])
		} else {
			stat = nil
			result = "OK"
		}
	}
	return sess.PreSend(
		yrpc.TypeReply,
		"/early/pong",
		result,
		stat,
	)
}
