package quic

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
	defer yrpc.SetLoggerLevel("ERROR")()

	cli := yrpc.NewPeer(yrpc.PeerConfig{Network: "quic"})
	defer cli.Close()
	e := cli.SetTLSConfigFromFile("cert.pem", "key.pem", true)
	if e != nil {
		yrpc.Fatalf("%v", e)
	}

	cli.RoutePush(new(Push))

	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}

	var result int
	stat = sess.Call("/math/add",
		[]int{1, 2, 3, 4, 5},
		&result,
		yrpc.WithAddMeta("author", "andeya"),
	).Status()
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	yrpc.Printf("result: %d", result)

	yrpc.Printf("wait for 10s...")
	time.Sleep(time.Second * 10)
}

// Push push handler
type Push struct {
	yrpc.PushCtx
}

// Push handles '/push/status' message
func (p *Push) Status(arg *string) *yrpc.Status {
	yrpc.Printf("%s", *arg)
	return nil
}
