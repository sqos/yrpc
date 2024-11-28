package async

import (
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
		CountTime:  true,
		ListenPort: 9090,
	})
	srv.RouteCall(new(test))
	srv.ListenAndServe()
}

type test struct {
	yrpc.CallCtx
}

func (t *test) Wait3s(arg *string) (string, *yrpc.Status) {
	time.Sleep(3 * time.Second)
	return *arg + " -> OK", nil
}
