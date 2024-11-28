package proxy_and_seq

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/plugin/proxy"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}" $GOFILE

func TestProxy(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	defer yrpc.FlushLogger()
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{
			ListenPort: 8080,
		},
		newProxyPlugin(),
	)
	srv.ListenAndServe()
}

func newProxyPlugin() yrpc.Plugin {
	cli := yrpc.NewPeer(yrpc.PeerConfig{RedialTimes: 3})
	var sess yrpc.Session
	var stat *yrpc.Status
DIAL:
	sess, stat = cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Warnf("%v", stat)
		time.Sleep(time.Second * 3)
		goto DIAL
	}
	return proxy.NewPlugin(func(*proxy.Label) proxy.Forwarder {
		return sess
	})
}
