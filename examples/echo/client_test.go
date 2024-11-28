package echo

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

	cli := yrpc.NewPeer(
		yrpc.PeerConfig{},
	)
	defer cli.Close()

	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}

	var result string
	stat = sess.Call(
		"/echo/add_suffix",
		"this is request",
		&result,
	).Status()

	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	yrpc.Printf("result: %s", result)
}
