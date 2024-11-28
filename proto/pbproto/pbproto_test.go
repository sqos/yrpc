package pbproto_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/proto/pbproto"
	"github.com/sqos/yrpc/xfer/gzip"
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

func TestPbProto(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	gzip.Reg('g', "gizp-5", 5)

	// server
	srv := yrpc.NewPeer(yrpc.PeerConfig{ListenPort: 9090})
	srv.RouteCall(new(Home))
	go srv.ListenAndServe(pbproto.NewPbProtoFunc())
	time.Sleep(1e9)

	// client
	cli := yrpc.NewPeer(yrpc.PeerConfig{})
	cli.RoutePush(new(Push))
	sess, stat := cli.Dial(":9090", pbproto.NewPbProtoFunc())
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
		yrpc.WithXferPipe('g'),
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
