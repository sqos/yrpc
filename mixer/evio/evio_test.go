package evio_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/mixer/evio"
	"github.com/sqos/yrpc/proto/jsonproto"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}" $GOFILE

func TestEvio(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	// server
	srv := evio.NewServer(1, yrpc.PeerConfig{ListenPort: 9090})
	// use TLS
	srv.SetTLSConfig(yrpc.GenerateTLSConfigForServer())
	srv.RouteCall(new(Home))
	go srv.ListenAndServe(jsonproto.NewJSONProtoFunc())
	time.Sleep(1e9)

	// client
	cli := evio.NewClient(yrpc.PeerConfig{})
	// use TLS
	cli.SetTLSConfig(yrpc.GenerateTLSConfigForClient())
	cli.RoutePush(new(Push))
	sess, stat := cli.Dial(":9090", jsonproto.NewJSONProtoFunc())
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result interface{}
	stat = sess.Call("/home/test",
		map[string]string{
			"author": "andeya",
		},
		&result,
		yrpc.WithAddMeta("peer_id", "110"),
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("result:%v", result)
	time.Sleep(2e9)
}

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

type Push struct {
	yrpc.PushCtx
}

func (p *Push) Test(arg *map[string]string) *yrpc.Status {
	yrpc.Infof("receive push(%s):\narg: %#v\n", p.IP(), arg)
	return nil
}
