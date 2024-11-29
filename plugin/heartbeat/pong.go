// Copyright 2018-2023 HenryLee. All Rights Reserved.
// Copyright 2024 sqos. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package heartbeat

import (
	"strconv"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil"

	"github.com/sqos/goutil/coarsetime"
)

// NewPong returns a heartbeat receiver plugin.
func NewPong() Pong {
	return new(heartPong)
}

type (
	// Pong receive heartbeat.
	Pong interface {
		// Name returns name.
		Name() string
		// PostNewPeer runs ping woker.
		PostNewPeer(peer yrpc.EarlyPeer) error
		// PostWriteCall updates heartbeat information.
		PostWriteCall(ctx yrpc.WriteCtx) *yrpc.Status
		// PostWritePush updates heartbeat information.
		PostWritePush(ctx yrpc.WriteCtx) *yrpc.Status
		// PostReadCallHeader updates heartbeat information.
		PostReadCallHeader(ctx yrpc.ReadCtx) *yrpc.Status
		// PostReadPushHeader updates heartbeat information.
		PostReadPushHeader(ctx yrpc.ReadCtx) *yrpc.Status
	}
	heartPong struct{}
)

var (
	_ yrpc.PostNewPeerPlugin        = Pong(nil)
	_ yrpc.PostWriteCallPlugin      = Pong(nil)
	_ yrpc.PostWritePushPlugin      = Pong(nil)
	_ yrpc.PostReadCallHeaderPlugin = Pong(nil)
	_ yrpc.PostReadPushHeaderPlugin = Pong(nil)
)

func (h *heartPong) Name() string {
	return "heart-pong"
}

func (h *heartPong) PostNewPeer(peer yrpc.EarlyPeer) error {
	peer.RouteCallFunc((*pongCall).heartbeat)
	peer.RoutePushFunc((*pongPush).heartbeat)
	rangeSession := peer.RangeSession
	const initial = time.Second*minRateSecond - 1
	interval := initial
	go func() {
		for {
			time.Sleep(interval)
			rangeSession(func(sess yrpc.Session) bool {
				info, ok := getHeartbeatInfo(sess.Swap())
				if !ok {
					return true
				}
				cp := info.elemCopy()
				if sess.Health() && cp.last.Add(cp.rate*2).Before(coarsetime.CeilingTimeNow()) {
					sess.Close()
				}
				if cp.rate < interval || interval == initial {
					interval = cp.rate
				}
				return true
			})
		}
	}()
	return nil
}

func (h *heartPong) PostReadCallHeader(ctx yrpc.ReadCtx) *yrpc.Status {
	h.update(ctx)
	return nil
}

func (h *heartPong) PostReadPushHeader(ctx yrpc.ReadCtx) *yrpc.Status {
	h.update(ctx)
	return nil
}

func (h *heartPong) PostWriteCall(ctx yrpc.WriteCtx) *yrpc.Status {
	return h.PostWritePush(ctx)
}

func (h *heartPong) PostWritePush(ctx yrpc.WriteCtx) *yrpc.Status {
	sess := ctx.Session()
	if !sess.Health() {
		return nil
	}
	updateHeartbeatInfo(sess.Swap(), 0)
	return nil
}

func (h *heartPong) update(ctx yrpc.ReadCtx) {
	if ctx.ServiceMethod() == HeartbeatServiceMethod {
		return
	}
	sess := ctx.Session()
	if !sess.Health() {
		return
	}
	updateHeartbeatInfo(sess.Swap(), 0)
}

type pongCall struct {
	yrpc.CallCtx
}

func (ctx *pongCall) heartbeat(_ *struct{}) (*struct{}, *yrpc.Status) {
	return nil, handelHeartbeat(ctx.Session(), ctx.PeekMeta)
}

type pongPush struct {
	yrpc.PushCtx
}

func (ctx *pongPush) heartbeat(_ *struct{}) *yrpc.Status {
	return handelHeartbeat(ctx.Session(), ctx.PeekMeta)
}

func handelHeartbeat(sess yrpc.CtxSession, peekMeta func(string) []byte) *yrpc.Status {
	rateStr := goutil.BytesToString(peekMeta(heartbeatMetaKey))
	rateSecond := parseHeartbeatRateSecond(rateStr)
	isFirst := updateHeartbeatInfo(sess.Swap(), time.Second*time.Duration(rateSecond))
	if isFirst && rateSecond == -1 {
		return yrpc.NewStatus(yrpc.CodeBadMessage, "invalid heartbeat rate", rateStr)
	}
	if rateSecond == 0 {
		yrpc.Tracef("heart-pong: %s", sess.ID())
	} else {
		yrpc.Tracef("heart-pong: %s, set rate: %ds", sess.ID(), rateSecond)
	}
	return nil
}

func parseHeartbeatRateSecond(s string) int {
	if len(s) == 0 {
		return 0
	}
	r, err := strconv.Atoi(s)
	if err != nil || r < 0 {
		return -1
	}
	return r
}
