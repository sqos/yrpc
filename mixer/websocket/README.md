## websocket

Websocket is an extension package that makes the yRPC framework compatible with websocket protocol as specified in RFC 6455.

### Usage

`import ws "github.com/sqos/yrpc/mixer/websocket"`

#### Test

```go

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sqos/yrpc"
	ws "github.com/sqos/yrpc/mixer/websocket"
	"github.com/sqos/yrpc/mixer/websocket/jsonSubProto"
	"github.com/sqos/yrpc/mixer/websocket/pbSubProto"
	"github.com/sqos/yrpc/plugin/auth"
)

type Arg struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ yrpc.CallCtx }

func (p *P) Divide(arg *Arg) (int, *yrpc.Status) {
	return arg.A / arg.B, nil
}

func TestJSONWebsocket(t *testing.T) {
	srv := ws.NewServer("/", yrpc.PeerConfig{ListenPort: 9090})
	srv.RouteCall(new(P))
	go srv.ListenAndServe()

	time.Sleep(time.Second * 1)

	cli := ws.NewClient("/", yrpc.PeerConfig{})
	sess, stat := cli.Dial(":9090")
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result int
	stat = sess.Call("/p/divide", &Arg{
		A: 10,
		B: 2,
	}, &result,
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/2=%d", result)
	time.Sleep(time.Second)
}

func TestPbWebsocketTLS(t *testing.T) {
	srv := ws.NewServer("/abc", yrpc.PeerConfig{ListenPort: 9091})
	srv.RouteCall(new(P))
	srv.SetTLSConfig(yrpc.GenerateTLSConfigForServer())
	go srv.ListenAndServeProtobuf()

	time.Sleep(time.Second * 1)

	cli := ws.NewClient("/abc", yrpc.PeerConfig{})
	cli.SetTLSConfig(yrpc.GenerateTLSConfigForClient())
	sess, err := cli.DialProtobuf(":9091")
	if err != nil {
		t.Fatal(err)
	}
	var result int
	stat := sess.Call("/p/divide", &Arg{
		A: 10,
		B: 2,
	}, &result,
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/2=%d", result)
	time.Sleep(time.Second)
}

func TestCustomizedWebsocket(t *testing.T) {
	srv := yrpc.NewPeer(yrpc.PeerConfig{})
	http.Handle("/ws", ws.NewPbServeHandler(srv, nil))
	go http.ListenAndServe(":9092", nil)
	srv.RouteCall(new(P))
	time.Sleep(time.Second * 1)

	cli := yrpc.NewPeer(yrpc.PeerConfig{}, ws.NewDialPlugin("/ws"))
	sess, stat := cli.Dial(":9092", pbSubProto.NewPbSubProtoFunc())
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result int
	stat = sess.Call("/p/divide", &Arg{
		A: 10,
		B: 2,
	}, &result,
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/2=%d", result)
	time.Sleep(time.Second)
}

func TestJSONWebsocketAuth(t *testing.T) {
	srv := ws.NewServer(
		"/auth",
		yrpc.PeerConfig{ListenPort: 9093},
		authChecker,
	)
	srv.RouteCall(new(P))
	go srv.ListenAndServe()

	time.Sleep(time.Second * 1)

	cli := ws.NewClient(
		"/auth",
		yrpc.PeerConfig{},
		authBearer,
	)
	sess, stat := cli.Dial(":9093")
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result int
	stat = sess.Call("/p/divide", &Arg{
		A: 10,
		B: 2,
	}, &result,
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/2=%d", result)
	time.Sleep(time.Second)
}

const clientAuthInfo = "client-auth-info-12345"

var authBearer = auth.NewBearerPlugin(
	func(sess auth.Session, fn auth.SendOnce) (stat *yrpc.Status) {
		var ret string
		stat = fn(clientAuthInfo, &ret)
		if !stat.OK() {
			return
		}
		yrpc.Infof("auth info: %s, result: %s", clientAuthInfo, ret)
		return
	},
	yrpc.WithBodyCodec('s'),
)

var authChecker = auth.NewCheckerPlugin(
	func(sess auth.Session, fn auth.RecvOnce) (ret interface{}, stat *yrpc.Status) {
		var authInfo string
		stat = fn(&authInfo)
		if !stat.OK() {
			return
		}
		yrpc.Infof("auth info: %v", authInfo)
		if clientAuthInfo != authInfo {
			return nil, yrpc.NewStatus(403, "auth fail", "auth fail detail")
		}
		return "pass", nil
	},
	yrpc.WithBodyCodec('s'),
)

func TestHandshakeWebsocketAuth(t *testing.T) {
	srv := yrpc.NewPeer(yrpc.PeerConfig{}, handshakePlugin)
	http.Handle("/token", ws.NewJSONServeHandler(srv, nil))
	go http.ListenAndServe(":9094", nil)
	srv.RouteCall(new(P))
	time.Sleep(time.Millisecond * 200)

	rawQuery := fmt.Sprintf("/token?%s=%s", clientAuthKey, clientAuthInfo)
	cli := yrpc.NewPeer(yrpc.PeerConfig{}, ws.NewDialPlugin(rawQuery))
	sess, stat := cli.Dial(":9094", jsonSubProto.NewJSONSubProtoFunc())
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result int
	stat = sess.Call("/p/divide", &Arg{
		A: 10,
		B: 2,
	}, &result,
	).Status()
	if !stat.OK() {
		t.Fatal(stat)
	}
	t.Logf("10/2=%d", result)
	time.Sleep(time.Second)
}

const clientAuthKey = "access_token"
const clientUserID = "user-1234"

var handshakePlugin = ws.NewHandshakeAuthPlugin(
	func(r *http.Request) (sessionId string, status *yrpc.Status) {
		token := ws.QueryToken(clientAuthKey, r)
		yrpc.Infof("auth token: %v", token)
		if token != clientAuthInfo {
			return "", yrpc.NewStatus(yrpc.CodeUnauthorized, yrpc.CodeText(yrpc.CodeUnauthorized))
		}
		return clientUserID, nil
	},
	func(sess yrpc.Session) *yrpc.Status {
		yrpc.Infof("login userID: %v", sess.ID())
		return nil
	},
)
```

test command:

```sh
go test -v -run=TestJSONWebsocket
go test -v -run=TestPbWebsocketTLS
go test -v -run=TestCustomizedWebsocket
go test -v -run=TestJSONWebsocketAuth
```

Among them, TestJSONWebsocket's request body is:

```json
{
  "seq": 0,
  "mtype": 1,
  "serviceMethod": "/p/divide",
  "meta": "",
  "bodyCodec": 106,
  "body": "{\"A\":10,\"B\":2}",
  "xferPipe": []
}
```

TestJSONWebsocket's response body is:

```json
{
  "seq": 0,
  "mtype": 2,
  "serviceMethod": "/p/divide",
  "meta": "",
  "bodyCodec": 106,
  "body": "5",
  "xferPipe": []
}
```
