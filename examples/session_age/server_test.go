package session_age

import (
	"context"
	"testing"
	"time"

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
	srv := yrpc.NewPeer(yrpc.PeerConfig{
		PrintDetail:       true,
		CountTime:         true,
		ListenPort:        9090,
		DefaultSessionAge: time.Second * 7,
		DefaultContextAge: time.Second * 2,
	})
	srv.RouteCall(new(test))
	srv.ListenAndServe()
}

type test struct {
	yrpc.CallCtx
}

func (t *test) Ok(arg *string) (string, *yrpc.Status) {
	return *arg + " -> OK", nil
}

func (t *test) Timeout(arg *string) (string, *yrpc.Status) {
	tCtx, _ := context.WithTimeout(t.Context(), time.Second)
	time.Sleep(time.Second)
	select {
	case <-tCtx.Done():
		return "", yrpc.NewStatus(
			yrpc.CodeHandleTimeout,
			yrpc.CodeText(yrpc.CodeHandleTimeout),
			tCtx.Err().Error(),
		)
	default:
		return *arg + " -> Not Timeout", nil
	}
}

func (t *test) Break(*struct{}) (*struct{}, *yrpc.Status) {
	time.Sleep(time.Second * 3)
	select {
	case <-t.Session().CloseNotify():
		yrpc.Infof("the connection has gone away!")
		return nil, yrpc.NewStatus(yrpc.CodeConnClosed, "", "")
	default:
		return nil, nil
	}
}
