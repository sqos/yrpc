package binder_test

import (
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/plugin/binder"
	"github.com/sqos/goutil"
)

type (
	Arg struct {
		A int
		B int    `param:"<range:1:100>"`
		C string `param:"<regexp:^[1-9]\\d*$>"`
		Query
		XyZ       string  `param:"<meta><nonzero><stat: 100002: Parameter cannot be empty>"`
		SwapValue float32 `param:"<swap><nonzero>"`
	}
	Query struct {
		X string `param:"<meta:_x>"`
	}
)

type P struct{ yrpc.CallCtx }

func (p *P) Divide(arg *Arg) (int, *yrpc.Status) {
	yrpc.Infof("meta arg _x: %s, xy_z: %s, swap_value: %v", arg.Query.X, arg.XyZ, arg.SwapValue)
	return arg.A / arg.B, nil
}

type SwapPlugin struct{}

func (s *SwapPlugin) Name() string {
	return "swap_plugin"
}
func (s *SwapPlugin) PostReadCallBody(ctx yrpc.ReadCtx) *yrpc.Status {
	ctx.Swap().Store("swap_value", 123)
	return nil
}

//go:generate go test -v -c -o "${GOPACKAGE}" $GOFILE

func TestBinder(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}
	bplugin := binder.NewStructArgsBinder(nil)
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{ListenPort: 9090},
	)
	srv.PluginContainer().AppendRight(bplugin)
	srv.RouteCall(new(P), new(SwapPlugin))
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := yrpc.NewPeer(yrpc.PeerConfig{})
	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result int
	stat = sess.Call("/p/divide", &Arg{
		A: 10,
		B: 2,
		C: "3",
	},
		&result,
		yrpc.WithSetMeta("_x", "testmeta_x"),
		yrpc.WithSetMeta("xy_z", "testmeta_xy_z"),
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/2=%d", result)
	stat = sess.Call("/p/divide", &Arg{
		A: 10,
		B: 5,
		C: "3",
	}, &result).Status()
	if stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/5 error:%v", stat)
	stat = sess.Call("/p/divide", &Arg{
		A: 10,
		B: 0,
		C: "3",
	}, &result).Status()
	if stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/0 error:%v", stat)
}
