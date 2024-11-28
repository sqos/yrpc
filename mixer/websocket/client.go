// Package websocket is an extension package that makes the yRPC framework compatible
// with websocket protocol as specified in RFC 6455.
//
// Copyright 2018-2023 HenryLee. All Rights Reserved.
// Copyright 2024 sqos. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package websocket

import (
	"net"
	"path"
	"strings"

	"github.com/sqos/yrpc"
	"github.com/sqos/yrpc/mixer/websocket/jsonSubProto"
	"github.com/sqos/yrpc/mixer/websocket/pbSubProto"
	ws "github.com/sqos/yrpc/mixer/websocket/websocket"
)

// Client a websocket client
type Client struct {
	yrpc.Peer
}

// NewClient creates a websocket client.
func NewClient(rootPath string, cfg yrpc.PeerConfig, globalLeftPlugin ...yrpc.Plugin) *Client {
	globalLeftPlugin = append([]yrpc.Plugin{NewDialPlugin(rootPath)}, globalLeftPlugin...)
	peer := yrpc.NewPeer(cfg, globalLeftPlugin...)
	return &Client{
		Peer: peer,
	}
}

// DialJSON connects with the JSON protocol.
func (c *Client) DialJSON(addr string) (yrpc.Session, *yrpc.Status) {
	return c.Dial(addr, jsonSubProto.NewJSONSubProtoFunc())
}

// DialProtobuf connects with the Protobuf protocol.
func (c *Client) DialProtobuf(addr string) (yrpc.Session, *yrpc.Status) {
	return c.Dial(addr, pbSubProto.NewPbSubProtoFunc())
}

// Dial connects with the peer of the destination address.
func (c *Client) Dial(addr string, protoFunc ...yrpc.ProtoFunc) (yrpc.Session, *yrpc.Status) {
	if len(protoFunc) == 0 {
		return c.Peer.Dial(addr, defaultProto)
	}
	return c.Peer.Dial(addr, protoFunc...)
}

// NewDialPlugin creates a websocket plugin for client.
func NewDialPlugin(rootPath string) yrpc.Plugin {
	return &clientPlugin{fixRootPath(rootPath)}
}

func fixRootPath(rootPath string) string {
	rootPath = path.Join("/", strings.TrimRight(rootPath, "/"))
	return rootPath
}

type clientPlugin struct {
	rootPath string
}

var (
	_ yrpc.PostDialPlugin = new(clientPlugin)
)

func (*clientPlugin) Name() string {
	return "websocket"
}

func (c *clientPlugin) PostDial(sess yrpc.PreSession, isRedial bool) (stat *yrpc.Status) {
	var location, origin string
	if sess.Peer().TLSConfig() == nil {
		location = "ws://" + sess.RemoteAddr().String() + c.rootPath
		origin = "ws://" + sess.LocalAddr().String() + c.rootPath
	} else {
		location = "wss://" + sess.RemoteAddr().String() + c.rootPath
		origin = "wss://" + sess.LocalAddr().String() + c.rootPath
	}
	cfg, err := ws.NewConfig(location, origin)
	if err != nil {
		return yrpc.NewStatus(yrpc.CodeDialFailed, "upgrade to websocket failed", err.Error())
	}
	sess.ModifySocket(func(conn net.Conn) (net.Conn, yrpc.ProtoFunc) {
		conn, err := ws.NewClient(cfg, conn)
		if err != nil {
			stat = yrpc.NewStatus(yrpc.CodeDialFailed, "upgrade to websocket failed", err.Error())
			return nil, nil
		}
		if isRedial {
			return conn, sess.GetProtoFunc()
		}
		return conn, NewWsProtoFunc(sess.GetProtoFunc())
	})
	return stat
}
