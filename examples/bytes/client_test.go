package bytes

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}_client" $GOFILE

func TestClient(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}

	defer yrpc.FlushLogger()
	go yrpc.GraceSignal()
	yrpc.SetShutdown(time.Second*20, nil, nil)
	var peer = yrpc.NewPeer(yrpc.PeerConfig{
		SlowCometDuration: time.Millisecond * 500,
		PrintDetail:       true,
	})
	defer peer.Close()
	peer.RoutePush(new(Push))

	var sess, stat = peer.Dial("127.0.0.1:9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	var result []byte
	for {
		if stat = sess.Call(
			"/group/home/test",
			[]byte("call text"),
			&result,
			yrpc.WithAddMeta("peer_id", "call-1"),
		).Status(); !stat.OK() {
			yrpc.Errorf("call error: %v", stat)
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}
	yrpc.Infof("test result: %s", result)

	stat = sess.Call(
		"/group/home/test_unknown",
		[]byte("unknown call text"),
		&result,
		yrpc.WithAddMeta("peer_id", "call-2"),
	).Status()
	if yrpc.IsConnError(stat) {
		yrpc.Fatalf("has conn error: %v", stat)
	}
	if !stat.OK() {
		yrpc.Fatalf("call error: %v", stat)
	}
	yrpc.Infof("test_unknown: %s", result)
}

// Push controller
type Push struct {
	yrpc.PushCtx
}

// Test handler
func (p *Push) Test(arg *[]byte) *yrpc.Status {
	yrpc.Infof("receive push(%s):\narg: %s\n", p.IP(), *arg)
	return nil
}
