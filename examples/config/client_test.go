package config

import (
	"testing"

	"github.com/sqos/cfgo"
	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}_client" $GOFILE

func TestClient(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	defer yrpc.FlushLogger()
	cfg := yrpc.PeerConfig{}

	// auto create and sync config/config.yaml
	cfgo.MustGet("config/config.yaml", true).MustReg("cfg_cli", &cfg)

	cli := yrpc.NewPeer(cfg)
	defer cli.Close()

	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}

	var result int
	stat = sess.Call("/math/add",
		[]int{1, 2, 3, 4, 5},
		&result,
	).Status()

	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
	yrpc.Printf("result: 1+2+3+4+5 = %d", result)
}
