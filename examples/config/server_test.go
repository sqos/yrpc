package config

import (
	"testing"

	"github.com/sqos/cfgo"
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
	go yrpc.GraceSignal()
	cfg := yrpc.PeerConfig{
		CountTime:  true,
		ListenPort: 9090,
	}

	// auto create and sync config/config.yaml
	cfgo.MustGet("config/config.yaml", true).MustReg("cfg_srv", &cfg)

	srv := yrpc.NewPeer(cfg)
	srv.RouteCall(new(math))
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
