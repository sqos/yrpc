package twoway

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
	defer yrpc.SetLoggerLevel("DEBUG")()

	cli := yrpc.NewPeer(yrpc.PeerConfig{RedialTimes: -1, RedialInterval: time.Second})
	defer cli.Close()
	cli.SetTLSConfig(yrpc.GenerateTLSConfigForClient())

	cli.RoutePush(new(Push))

	cli.SubRoute("/cli").
		RoutePush(new(Push))

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
	yrpc.Printf("Wait 10 seconds to receive the push...")
	time.Sleep(time.Second * 10)

	stat = sess.Call("/srv/math/v2/add_2",
		[]int{10, 20, 30, 40, 50},
		&result,
		yrpc.WithSetMeta("push_status", "yes"),
	).Status()
}

// Push push handler
type Push struct {
	yrpc.PushCtx
}

// Push handles '/push/status' message
func (p *Push) Status(arg *string) *yrpc.Status {
	yrpc.Printf("server status: %s", *arg)
	return nil
}
