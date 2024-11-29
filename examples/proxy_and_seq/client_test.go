package proxy_and_seq

import (
	"fmt"
	"testing"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/socket"
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

	sess, stat := cli.Dial(":8080")
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
	yrpc.Printf("result: %d", result)

	stat = sess.Push(
		"/chat/say",
		fmt.Sprintf("I get result %d", result),
		socket.WithSetMeta("X-ID", "client-001"),
	)
	if !stat.OK() {
		yrpc.Fatalf("%v", stat)
	}
}
