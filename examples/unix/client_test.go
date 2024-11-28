package unix

import (
	"net"
	"os"
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

	cli := yrpc.NewPeer(
		yrpc.PeerConfig{
			Network:   "unix",
			LocalPort: 9091,
		},
		&yrpc.PluginImpl{
			PluginName: "clean-local-unix-file",
			OnPreDial: func(localAddr net.Addr, remoteAddr string) (stat *yrpc.Status) {
				addr := localAddr.String()
				if _, err := os.Stat(addr); err == nil {
					if err := os.RemoveAll(addr); err != nil {
						return yrpc.NewStatusByCodeText(yrpc.CodeDialFailed, err, false)
					}
				}
				return nil
			},
		},
	)
	defer cli.Close()

	sess, stat := cli.Dial("./0.0.0.0:9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}

	var result string
	stat = sess.Call(
		"/echo/parrot",
		"this is request",
		&result,
	).Status()

	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	yrpc.Printf("result: %s", result)
}
