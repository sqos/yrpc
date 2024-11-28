## proxy

A plugin for handling unknown calling or pushing.

#### Demo

```go
package main

import (
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/plugin/proxy"
)

func main() {
	defer yrpc.FlushLogger()
	srv := yrpc.NewPeer(
		yrpc.PeerConfig{
			ListenPort: 8080,
		},
		newUnknownProxy(),
	)
	srv.ListenAndServe()
}

func newUnknownProxy() yrpc.Plugin {
	cli := yrpc.NewPeer(yrpc.PeerConfig{RedialTimes: 3})
	var sess yrpc.Session
	var stat *yrpc.Status
DIAL:
	sess, stat = cli.Dial(":9090")
	if !stat.OK() {
		yrpc.Warnf("%v", stat)
		time.Sleep(time.Second * 3)
		goto DIAL
	}
	return proxy.NewPlugin(func(*proxy.Label) proxy.Forwarder {
		return sess
	})
}
```
