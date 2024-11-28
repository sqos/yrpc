// Package auth is a plugin for verifying peer at the first time.
//
// Copyright 2017-2023 HenryLee. All Rights Reserved.
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
package auth

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"
)

// NewBearerPlugin creates a auth bearer plugin for client.
func NewBearerPlugin(fn Bearer, infoSetting ...yrpc.MessageSetting) yrpc.Plugin {
	return &authBearerPlugin{
		bearerFunc: fn,
		msgSetting: infoSetting,
	}
}

// NewCheckerPlugin creates a auth checker plugin for server.
func NewCheckerPlugin(fn Checker, retSetting ...yrpc.MessageSetting) yrpc.Plugin {
	return &authCheckerPlugin{
		checkerFunc: fn,
		msgSetting:  retSetting,
	}
}

type (
	// Bearer initiates an authorization request and handles the response.
	Bearer func(sess Session, fn SendOnce) *yrpc.Status
	// SendOnce sends authorization request once.
	SendOnce func(info, retRecv interface{}) *yrpc.Status

	// Checker checks the authorization request.
	Checker func(sess Session, fn RecvOnce) (ret interface{}, stat *yrpc.Status)
	// RecvOnce receives authorization request once.
	RecvOnce func(infoRecv interface{}) *yrpc.Status

	// Session auth session provides SetID, RemoteAddr and Swap methods in base session
	Session interface {
		// Peer returns the peer.
		Peer() yrpc.Peer
		// SetID sets the session id.
		SetID(newID string)
		// LocalAddr returns the local network address.
		LocalAddr() net.Addr
		// RemoteAddr returns the remote network address.
		RemoteAddr() net.Addr
		// Swap returns custom data swap of the session(socket).
		Swap() goutil.Map
	}
)

type authBearerPlugin struct {
	bearerFunc Bearer
	msgSetting []yrpc.MessageSetting
}

type authCheckerPlugin struct {
	checkerFunc Checker
	msgSetting  []yrpc.MessageSetting
}

var (
	_ yrpc.PostDialPlugin   = new(authBearerPlugin)
	_ yrpc.PostAcceptPlugin = new(authCheckerPlugin)
)

func (a *authBearerPlugin) Name() string {
	return "auth-bearer"
}

func (a *authCheckerPlugin) Name() string {
	return "auth-checker"
}

// MultiSendErr the error of multiple call SendOnce function
var MultiSendErr = yrpc.NewStatus(
	yrpc.CodeWriteFailed,
	"auth-bearer plugin usage is incorrect",
	"multiple call SendOnce function",
)

// MultiRecvErr the error of multiple call RecvOnce function
var MultiRecvErr = yrpc.NewStatus(
	yrpc.CodeInternalServerError,
	"auth-checker plugin usage is incorrect",
	"multiple call RecvOnce function",
)

func (a *authBearerPlugin) PostDial(sess yrpc.PreSession, _ bool) *yrpc.Status {
	if a.bearerFunc == nil {
		return nil
	}
	var called int32
	return a.bearerFunc(sess, func(info, retRecv interface{}) *yrpc.Status {
		if !atomic.CompareAndSwapInt32(&called, 0, 1) {
			return MultiSendErr
		}
		stat := sess.PreSend(yrpc.TypeAuthCall, "", info, nil, a.msgSetting...)
		if !stat.OK() {
			return stat
		}
		retMsg := sess.PreReceive(func(header yrpc.Header) interface{} {
			if header.Mtype() != yrpc.TypeAuthReply {
				return nil
			}
			return retRecv
		})
		if !retMsg.StatusOK() {
			return retMsg.Status()
		}
		if retMsg.Mtype() != yrpc.TypeAuthReply {
			return yrpc.NewStatus(
				yrpc.CodeUnauthorized,
				yrpc.CodeText(yrpc.CodeUnauthorized),
				fmt.Sprintf("auth message(1st) expect: AUTH_REPLY, but received: %s",
					yrpc.TypeText(retMsg.Mtype())),
			)
		}
		return nil
	})
}

func (a *authCheckerPlugin) PostAccept(sess yrpc.PreSession) *yrpc.Status {
	if a.checkerFunc == nil {
		return nil
	}
	var called int32
	ret, stat := a.checkerFunc(sess, func(infoRecv interface{}) *yrpc.Status {
		if !atomic.CompareAndSwapInt32(&called, 0, 1) {
			return MultiRecvErr
		}
		infoMsg := sess.PreReceive(func(header yrpc.Header) interface{} {
			if header.Mtype() != yrpc.TypeAuthCall {
				return nil
			}
			return infoRecv
		})
		if !infoMsg.StatusOK() {
			return infoMsg.Status()
		}
		if infoMsg.Mtype() != yrpc.TypeAuthCall {
			return yrpc.NewStatus(
				yrpc.CodeUnauthorized,
				yrpc.CodeText(yrpc.CodeUnauthorized),
				fmt.Sprintf("auth message(1st) expect: AUTH_CALL, but received: %s",
					yrpc.TypeText(infoMsg.Mtype())),
			)
		}
		return nil
	})
	if stat == MultiRecvErr {
		sess.PreSend(yrpc.TypeAuthReply, "", nil, stat, a.msgSetting...)
		return stat
	}
	stat2 := sess.PreSend(yrpc.TypeAuthReply, "", ret, stat, a.msgSetting...)
	if !stat2.OK() {
		return stat2
	}
	return stat
}
