package echo

import (
	"fmt"
	"strconv"
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
	srv.RouteCall(func() yrpc.CtrlStructPtr { return &Echo{Suffix: strconv.FormatInt(time.Now().UnixMilli(), 10)} })
	srv.ListenAndServe()
}

type Echo struct {
	yrpc.CallCtx
	Suffix string
}

func (echo *Echo) AddSuffix(arg *string) (string, *yrpc.Status) {
	return fmt.Sprintf("%s ------ %s", *arg, echo.Suffix), nil
}
