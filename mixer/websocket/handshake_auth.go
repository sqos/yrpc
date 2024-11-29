// Copyright 2021-2023 HenryLee. All Rights Reserved.
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
	"net/http"

	"github.com/sqos/yrpc"
	ws "github.com/sqos/yrpc/mixer/websocket/websocket"
)

// NewHandshakeAuthPlugin creates a handshake auth plugin for server.
func NewHandshakeAuthPlugin(ckFn Checker, apFn Acceptor) *HandshakeAuthPlugin {
	return &HandshakeAuthPlugin{
		CheckFunc:  ckFn,
		AcceptFunc: apFn,
	}
}

// Acceptor provide authenticated yrpc.Session
// you can get the sessionID that your return by Checker()
type Acceptor func(sess yrpc.Session) *yrpc.Status

// Checker deal with http.Request and your authenticate logic,
// the a sessionID returned will used by yrpc.Session.SetID(), if auth succeeded.
type Checker func(r *http.Request) (sessionID string, status *yrpc.Status)

type HandshakeAuthPlugin struct {
	CheckFunc  Checker
	AcceptFunc Acceptor
}

var (
	_ PostWebsocketAcceptPlugin   = new(HandshakeAuthPlugin)
	_ PreWebsocketHandshakePlugin = new(HandshakeAuthPlugin)
)

func (p *HandshakeAuthPlugin) Name() string {
	return "handshake-auth-plugin"
}

// Used by store sessionID in http.Header
// note, the Header, it may covered the user's request.
const sessionHeader = "Erpc-Session-Id"

func (p *HandshakeAuthPlugin) PreHandshake(r *http.Request) *yrpc.Status {
	if p.CheckFunc == nil {
		return nil
	}
	id, stat := p.CheckFunc(r)
	r.Header.Set(sessionHeader, id)
	return stat
}

func (p *HandshakeAuthPlugin) PostAccept(sess yrpc.Session, conn *ws.Conn) *yrpc.Status {
	if p.AcceptFunc == nil {
		return nil
	}
	id := conn.Request().Header.Get(sessionHeader)
	sess.SetID(id)
	stat := p.AcceptFunc(sess)
	return stat
}

func QueryToken(tokenKey string, r *http.Request) (token string) {
	queryParams := r.URL.Query()
	if values, ok := queryParams[tokenKey]; ok {
		token = values[0]
	}
	return token
}
