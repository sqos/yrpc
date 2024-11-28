package detail

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/xfer/gzip"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}_client" $GOFILE

func TestClient(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	defer yrpc.FlushLogger()
	gzip.Reg('g', "gizp", 5)

	go yrpc.GraceSignal()
	yrpc.SetShutdown(time.Second*20, nil, nil)
	var peer = yrpc.NewPeer(yrpc.PeerConfig{
		SlowCometDuration: time.Millisecond * 500,
		// DefaultBodyCodec:    "json",
		// DefaultContextAge: time.Second * 5,
		PrintDetail:    true,
		CountTime:      true,
		RedialTimes:    3,
		RedialInterval: time.Second * 3,
	})
	defer peer.Close()
	peer.RoutePush(new(Push))

	var sess, stat = peer.Dial("127.0.0.1:9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	sess.SetID("testId")

	var result interface{}
	for {
		if stat = sess.Call(
			"/group/home/test",
			map[string]interface{}{
				"bytes": []byte("test bytes"),
			},
			&result,
			yrpc.WithBodyCodec('j'),
			yrpc.WithAcceptBodyCodec('j'),
			yrpc.WithXferPipe('g'),
			yrpc.WithSetMeta("peer_id", "call-1"),
			yrpc.WithAddMeta("add", "1"),
			yrpc.WithAddMeta("add", "2"),
		).Status(); !stat.OK() {
			yrpc.Errorf("call error: %v", stat)
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}
	yrpc.Infof("test: %#v", result)

	// sess.Close()

	stat = sess.Call(
		"/group/home/test_unknown",
		struct {
			RawMessage json.RawMessage
			Bytes      []byte
		}{
			json.RawMessage(`{"RawMessage":"test_unknown"}`),
			[]byte("test bytes"),
		},
		&result,
		yrpc.WithXferPipe('g'),
	).Status()
	if yrpc.IsConnError(stat) {
		yrpc.Fatalf("has conn error: %v", stat)
	}
	if !stat.OK() {
		yrpc.Fatalf("call error: %v", stat)
	}
	yrpc.Infof("test_unknown: %#v", result)
}

// Push controller
type Push struct {
	yrpc.PushCtx
}

// Test handler
func (p *Push) Test(arg *map[string]interface{}) *yrpc.Status {
	yrpc.Infof("receive push(%s):\narg: %#v\n", p.IP(), arg)
	return nil
}
