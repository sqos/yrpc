package unix

import (
	"os"
	"testing"

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
		Network:    "unix",
	}, &yrpc.PluginImpl{
		PluginName: "clean-listen-unix-file",
		OnPreNewPeer: func(config *yrpc.PeerConfig, _ *yrpc.PluginContainer) error {
			socketFile := config.ListenAddr().String()
			if _, err := os.Stat(socketFile); err == nil {
				if err := os.RemoveAll(socketFile); err != nil {
					return err
				}
			}
			return nil
		},
	})
	srv.RouteCall(&Echo{})
	srv.ListenAndServe()
}

type Echo struct {
	yrpc.CallCtx
}

func (echo *Echo) Parrot(arg *string) (string, *yrpc.Status) {
	return *arg, nil
}
