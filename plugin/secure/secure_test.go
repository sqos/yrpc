package secure_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/plugin/secure"
	"github.com/sqos/goutil"
)

//go:generate go test -v -c -o "${GOPACKAGE}" $GOFILE

type Arg struct {
	A int
	B int
}

type Result struct {
	C int
}

type math struct{ yrpc.CallCtx }

func (m *math) Add(arg *Arg) (*Result, *yrpc.Status) {
	// enforces the body of the encrypted reply message.
	// secure.EnforceSecure(m.Output())
	return &Result{C: arg.A + arg.B}, nil
}

func newSession(t *testing.T, port uint16) yrpc.Session {
	p := secure.NewPlugin(100001, "cipherkey1234567")
	srv := yrpc.NewPeer(yrpc.PeerConfig{
		ListenPort:  port,
		PrintDetail: true,
	})
	srv.RouteCall(new(math), p)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := yrpc.NewPeer(yrpc.PeerConfig{
		PrintDetail: true,
	}, p)
	sess, stat := cli.Dial(":" + strconv.Itoa(int(port)))
	if !stat.OK() {
		t.Fatal(stat)
	}
	return sess
}

func TestSecurePlugin(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	sess := newSession(t, 9090)
	// test secure
	var result Result
	stat := sess.Call(
		"/math/add",
		&Arg{A: 10, B: 2},
		&result,
		secure.WithSecureMeta(),
		// secure.WithAcceptSecureMeta(false),
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	if result.C != 12 {
		t.Fatalf("expect 12, but get %d", result.C)
	}
	t.Logf("test secure10+2=%d", result.C)
}

func TestAcceptSecurePlugin(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	sess := newSession(t, 9091)
	// test accept secure
	var result Result
	stat := sess.Call(
		"/math/add",
		&Arg{A: 20, B: 4},
		&result,
		secure.WithAcceptSecureMeta(true),
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	if result.C != 24 {
		t.Fatalf("expect 24, but get %d", result.C)
	}
	t.Logf("test accept secure: 20+4=%d", result.C)
}
