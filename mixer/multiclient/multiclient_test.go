package multiclient_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/mixer/multiclient"
	"github.com/sqos/goutil"
)

type Arg struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ yrpc.CallCtx }

func (p *P) Divide(arg *Arg) (int, *yrpc.Status) {
	return arg.A / arg.B, nil
}

//go:generate go test -v -c -o "${GOPACKAGE}" $GOFILE

func TestMultiClient(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	srv := yrpc.NewPeer(yrpc.PeerConfig{
		ListenPort: 9090,
	})
	srv.RouteCall(new(P))
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := multiclient.New(
		yrpc.NewPeer(yrpc.PeerConfig{}),
		":9090",
		100,
		time.Second*5,
	)
	go func() {
		for {
			t.Logf("%+v", cli.Stats())
			time.Sleep(time.Millisecond * 500)
		}
	}()
	go func() {
		var result int
		for i := 0; ; i++ {
			stat := cli.Call("/p/divide", &Arg{
				A: i,
				B: 2,
			}, &result).Status()
			if !stat.OK() {
				t.Log(stat)
			} else {
				t.Logf("%d/2=%v", i, result)
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
	time.Sleep(time.Second * 6)
	cli.Close()
	time.Sleep(time.Second * 3)
}
