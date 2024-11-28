package detail

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/xfer/gzip"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}_server" $GOFILE

func TestServer(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}

	defer yrpc.FlushLogger()
	gzip.Reg('g', "gizp", 5)

	go yrpc.GraceSignal()
	// yrpc.SetReadLimit(10)
	yrpc.SetShutdown(time.Second*20, nil, nil)
	var peer = yrpc.NewPeer(yrpc.PeerConfig{
		SlowCometDuration: time.Millisecond * 500,
		PrintDetail:       true,
		CountTime:         true,
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
func (h *Home) Test(arg *map[string]interface{}) (map[string]interface{}, *yrpc.Status) {
	h.Session().Push("/push/test", map[string]interface{}{
		"your_id": string(h.PeekMeta("peer_id")),
	})
	h.VisitMeta(func(k, v []byte) {
		yrpc.Infof("meta: key: %s, value: %s", k, v)
	})
	time.Sleep(5e9)
	return map[string]interface{}{
		"your_arg":    *arg,
		"server_time": time.Now(),
	}, nil
}

// UnknownCallHandle handles unknown call message
func UnknownCallHandle(ctx yrpc.UnknownCallCtx) (interface{}, *yrpc.Status) {
	time.Sleep(1)
	var v = struct {
		RawMessage json.RawMessage
		Bytes      []byte
	}{}
	codecID, err := ctx.Bind(&v)
	if err != nil {
		return nil, yrpc.NewStatus(1001, "bind error", err.Error())
	}
	yrpc.Debugf("UnknownCallHandle: codec: %d, RawMessage: %s, bytes: %s",
		codecID, v.RawMessage, v.Bytes,
	)
	ctx.Session().Push("/push/test", map[string]interface{}{
		"your_id": string(ctx.PeekMeta("peer_id")),
	})
	return map[string]interface{}{
		"your_arg":    v,
		"server_time": time.Now(),
		"meta":        ctx.CopyMeta().String(),
	}, nil
}
