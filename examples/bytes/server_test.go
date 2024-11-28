package bytes

import (
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
	go yrpc.GraceSignal()
	yrpc.SetShutdown(time.Second*20, nil, nil)
	var peer = yrpc.NewPeer(yrpc.PeerConfig{
		SlowCometDuration: time.Millisecond * 500,
		PrintDetail:       true,
		ListenPort:        9090,
	})
	group := peer.SubRoute("group")
	group.RouteCall(new(Home))
	peer.SetUnknownCall(UnknownCallHandle)
	peer.ListenAndServe()
}

// Home controller
type Home struct {
	yrpc.CallCtx
}

// Test handler
func (h *Home) Test(arg *[]byte) ([]byte, *yrpc.Status) {
	h.Session().Push("/push/test", []byte("test push text"))
	yrpc.Debugf("HomeCallHandle: codec: %d, arg: %s", h.GetBodyCodec(), *arg)
	return []byte("test call result text"), nil
}

// UnknownCallHandle handles unknown call message
func UnknownCallHandle(ctx yrpc.UnknownCallCtx) (interface{}, *yrpc.Status) {
	ctx.Session().Push("/push/test", []byte("test unknown push text"))
	var arg []byte
	codecID, err := ctx.Bind(&arg)
	if err != nil {
		return nil, yrpc.NewStatus(1001, "bind error", err.Error())
	}
	yrpc.Debugf("UnknownCallHandle: codec: %d, arg: %s", codecID, arg)
	return []byte("test unknown call result text"), nil
}
