package route_func

import (
	"testing"

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

	cli := yrpc.NewPeer(yrpc.PeerConfig{})
	defer cli.Close()

	cli.RoutePushFunc((*pushCtrl).ServerStatus1)
	cli.RoutePushFunc(ServerStatus2)

	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}

	var result int
	stat = sess.Call("/math/add1",
		[]int{1, 2, 3, 4, 5},
		&result,
	).Status()

	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	yrpc.Printf("result1: %d", result)

	stat = sess.Call("/math/add2",
		[]int{1, 2, 3, 4, 5},
		&result,
		yrpc.WithAddMeta("push_status", "yes"),
	).Status()

	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	yrpc.Printf("result2: %d", result)
}

type pushCtrl struct {
	yrpc.PushCtx
}

func (c *pushCtrl) ServerStatus1(arg *string) *yrpc.Status {
	return ServerStatus2(c, arg)
}

func ServerStatus2(ctx yrpc.PushCtx, arg *string) *yrpc.Status {
	yrpc.Printf("server status(%s): %s", ctx.ServiceMethod(), *arg)
	return nil
}
