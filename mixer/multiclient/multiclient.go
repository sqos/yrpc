// Package multiclient is a higher throughput client connection pool when transferring large messages (such as downloading files).
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
package multiclient

import (
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil/pool"
)

// MultiClient client session which is has connection pool
type MultiClient struct {
	addr string
	peer yrpc.Peer
	pool *pool.Workshop
}

// New creates a client session which is has connection pool.
func New(peer yrpc.Peer, addr string, sessMaxQuota int, sessMaxIdleDuration time.Duration, protoFunc ...yrpc.ProtoFunc) *MultiClient {
	newWorkerFunc := func() (pool.Worker, error) {
		sess, stat := peer.Dial(addr, protoFunc...)
		return sess, stat.Cause()
	}
	return &MultiClient{
		addr: addr,
		peer: peer,
		pool: pool.NewWorkshop(sessMaxQuota, sessMaxIdleDuration, newWorkerFunc),
	}
}

// Addr returns the address.
func (c *MultiClient) Addr() string {
	return c.addr
}

// Peer returns the peer.
func (c *MultiClient) Peer() yrpc.Peer {
	return c.peer
}

// Close closes the session.
func (c *MultiClient) Close() {
	c.pool.Close()
}

// Stats returns the current session pool stats.
func (c *MultiClient) Stats() pool.WorkshopStats {
	return c.pool.Stats()
}

// AsyncCall sends a message and receives reply asynchronously.
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name.
func (c *MultiClient) AsyncCall(
	uri string,
	arg interface{},
	result interface{},
	callCmdChan chan<- yrpc.CallCmd,
	setting ...yrpc.MessageSetting,
) yrpc.CallCmd {
	_sess, err := c.pool.Hire()
	if err != nil {
		callCmd := yrpc.NewFakeCallCmd(uri, arg, result, yrpc.NewStatusByCodeText(yrpc.CodeWrongConn, err, false))
		if callCmdChan != nil && cap(callCmdChan) == 0 {
			yrpc.Panicf("*MultiClient.AsyncCall(): callCmdChan channel is unbuffered")
		}
		callCmdChan <- callCmd
		return callCmd
	}
	sess := _sess.(yrpc.Session)
	defer c.pool.Fire(sess)
	return sess.AsyncCall(uri, arg, result, callCmdChan, setting...)
}

// Call sends a message and receives reply.
// NOTE:
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *MultiClient) Call(uri string, arg interface{}, result interface{}, setting ...yrpc.MessageSetting) yrpc.CallCmd {
	callCmd := c.AsyncCall(uri, arg, result, make(chan yrpc.CallCmd, 1), setting...)
	<-callCmd.Done()
	return callCmd
}

// Push sends a message, but do not receives reply.
// NOTE:
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *MultiClient) Push(uri string, arg interface{}, setting ...yrpc.MessageSetting) *yrpc.Status {
	_sess, err := c.pool.Hire()
	if err != nil {
		return yrpc.NewStatusByCodeText(yrpc.CodeWrongConn, err, false)
	}
	sess := _sess.(yrpc.Session)
	defer c.pool.Fire(sess)
	return sess.Push(uri, arg, setting...)
}
