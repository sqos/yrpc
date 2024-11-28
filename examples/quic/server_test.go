package quic

import (
	"fmt"
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
	// graceful
	go yrpc.GraceSignal()

	// server peer
	srv := yrpc.NewPeer(yrpc.PeerConfig{
		Network:     "quic",
		CountTime:   true,
		ListenPort:  9090,
		PrintDetail: true,
	})
	e := srv.SetTLSConfigFromFile("cert.pem", "key.pem")
	if e != nil {
		yrpc.Fatalf("%v", e)
	}

	// router
	srv.RouteCall(new(Math))

	// broadcast per 5s
	go func() {
		for {
			time.Sleep(time.Second * 5)
			srv.RangeSession(func(sess yrpc.Session) bool {
				sess.Push(
					"/push/status",
					fmt.Sprintf("this is a broadcast, server time: %v", time.Now()),
				)
				return true
			})
		}
	}()

	// listen and serve
	srv.ListenAndServe()
	select {}
}

// Math handler
type Math struct {
	yrpc.CallCtx
}

// Add handles addition request
func (m *Math) Add(arg *[]int) (int, *yrpc.Status) {
	// test query parameter
	yrpc.Infof("author: %s", m.PeekMeta("author"))
	// add
	var r int
	for _, a := range *arg {
		r += a
	}
	// response
	return r, nil
}
