package ignorecase_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/plugin/ignorecase"
	"github.com/sqos/goutil"
)

type Home struct {
	yrpc.CallCtx
}

func (h *Home) Test(arg *map[string]string) (map[string]interface{}, *yrpc.Status) {
	h.Session().Push("/push/test", map[string]string{
		"your_id": string(h.PeekMeta("peer_id")),
	})
	return map[string]interface{}{
		"arg": *arg,
	}, nil
}

//go:generate go test -v -c -o "${GOPACKAGE}" $GOFILE

func TestIngoreCase(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	// Server
	srv := yrpc.NewPeer(yrpc.PeerConfig{ListenPort: 9090, Network: "tcp"}, ignorecase.NewIgnoreCase())
	srv.RouteCall(new(Home))
	go srv.ListenAndServe()
	time.Sleep(2e9)

	// Client
	cli := yrpc.NewPeer(yrpc.PeerConfig{Network: "tcp"}, ignorecase.NewIgnoreCase())
	cli.RoutePush(new(Push))
	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result interface{}
	stat = sess.Call("/home/TesT",
		map[string]string{
			"author": "andeya",
		},
		&result,
		yrpc.WithAddMeta("peer_id", "110"),
	).Status()
	if !stat.OK() {
		t.Error(stat)
	}
	t.Logf("result:%v", result)
	time.Sleep(3e9)
}

type Push struct {
	yrpc.PushCtx
}

func (p *Push) Test(arg *map[string]string) *yrpc.Status {
	yrpc.Infof("receive push(%s):\narg: %#v\n", p.IP(), arg)
	return nil
}
