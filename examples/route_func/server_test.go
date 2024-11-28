package route_func

import (
	"fmt"
	"testing"
	"time"

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
	srv := yrpc.NewPeer(yrpc.PeerConfig{
		CountTime:  true,
		ListenPort: 9090,
	})
	srv.RouteCallFunc((*callCtrl).math_add1)
	srv.RouteCallFunc(math_add2)
	srv.ListenAndServe()
}

type callCtrl struct {
	yrpc.CallCtx
}

func (c *callCtrl) math_add1(arg *[]int) (int, *yrpc.Status) {
	return math_add2(c, arg)
}

func math_add2(ctx yrpc.CallCtx, arg *[]int) (int, *yrpc.Status) {
	if string(ctx.PeekMeta("push_status")) == "yes" {
		ctx.Session().Push(
			"/server/status1",
			fmt.Sprintf("%d numbers are being added...", len(*arg)),
		)
		ctx.Session().Push(
			"/server/status2",
			fmt.Sprintf("%d numbers are being added...", len(*arg)),
		)
		time.Sleep(time.Millisecond * 10)
	}
	var r int
	for _, a := range *arg {
		r += a
	}
	return r, nil
}
