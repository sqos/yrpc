package evio_bench

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"testing"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/examples/bench/msg"
	"github.com/sqos/yrpc/mixer/evio"
	"github.com/sqos/goutil"
)

var delay *time.Duration

type Hello struct {
	yrpc.CallCtx
}

func (t *Hello) Say(args *msg.BenchmarkMessage) (*msg.BenchmarkMessage, *yrpc.Status) {
	s := "OK"
	var i int32 = 100
	args.Field1 = s
	args.Field2 = i
	if *delay > 0 {
		time.Sleep(*delay)
	} else {
		runtime.Gosched()
	}
	return args, nil
}

//go:generate go test -v -c -o "${GOPACKAGE}_server" $GOFILE

func TestServer(t *testing.T) {
	if goutil.IsGoTest() {
		t.Log("skip test in go test")
		return
	}

	var (
		port      = flag.Int64("p", 8972, "listened port")
		debugAddr = flag.String("d", "127.0.0.1:9981", "server ip and port")
	)
	delay = flag.Duration("delay", 0, "delay to mock business processing")
	flag.Parse()

	defer yrpc.SetLoggerLevel("ERROR")()
	yrpc.SetGopool(1024*1024*100, time.Minute*10)

	go func() {
		log.Println(http.ListenAndServe(*debugAddr, nil))
	}()

	yrpc.SetServiceMethodMapper(yrpc.RPCServiceMethodMapper)
	server := evio.NewServer(1, yrpc.PeerConfig{
		DefaultBodyCodec: "protobuf",
		ListenPort:       uint16(*port),
	})
	server.RouteCall(new(Hello))
	server.ListenAndServe()
}
