package proxy_and_seq

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
	srv := yrpc.NewPeer(yrpc.PeerConfig{
		CountTime:  true,
		ListenPort: 9090,
	})
	srv.RouteCall(new(math))
	srv.RoutePush(new(chat))
	srv.ListenAndServe()
}

type math struct {
	yrpc.CallCtx
}

func (m *math) Add(arg *[]int) (int, *yrpc.Status) {
	var r int
	for _, a := range *arg {
		r += a
	}
	return r, nil
}

type chat struct {
	yrpc.PushCtx
}

func (c *chat) Say(arg *string) *yrpc.Status {
	yrpc.Printf("%s say: %q", c.PeekMeta("X-ID"), *arg)
	return nil
}
