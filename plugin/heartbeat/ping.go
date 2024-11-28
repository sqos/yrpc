// Heartbeat is a generic timing heartbeat plugin.
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
package heartbeat

import (
	"strconv"
	"sync"
	"time"

	"github.com/sqos/yrpc"
	"github.com/sqos/goutil/coarsetime"
)

const (
	// HeartbeatServiceMethod heartbeat service method
	HeartbeatServiceMethod = "/heartbeat"
	heartbeatMetaKey       = "hb_"
)

// NewPing returns a heartbeat(CALL or PUSH) sender plugin.
func NewPing(rateSecond int, useCall bool) Ping {
	p := new(heartPing)
	p.useCall = useCall
	p.SetRate(rateSecond)
	return p
}

type (
	// Ping send heartbeat.
	Ping interface {
		// SetRate sets heartbeat rate.
		SetRate(rateSecond int)
		// UseCall uses CALL method to ping.
		UseCall()
		// UsePush uses PUSH method to ping.
		UsePush()
		// Name returns name.
		Name() string
		// PostNewPeer runs ping woker.
		PostNewPeer(peer yrpc.EarlyPeer) error
		// PostDial initializes heartbeat information.
		PostDial(sess yrpc.PreSession, isRedial bool) *yrpc.Status
		// PostAccept initializes heartbeat information.
		PostAccept(sess yrpc.PreSession) *yrpc.Status
		// PostWriteCall updates heartbeat information.
		PostWriteCall(ctx yrpc.WriteCtx) *yrpc.Status
		// PostWritePush updates heartbeat information.
		PostWritePush(ctx yrpc.WriteCtx) *yrpc.Status
		// PostReadCallHeader updates heartbeat information.
		PostReadCallHeader(ctx yrpc.ReadCtx) *yrpc.Status
		// PostReadPushHeader updates heartbeat information.
		PostReadPushHeader(ctx yrpc.ReadCtx) *yrpc.Status
	}
	heartPing struct {
		peer           yrpc.Peer
		pingRate       time.Duration
		mu             sync.RWMutex
		once           sync.Once
		pingRateSecond string
		useCall        bool
	}
)

var (
	_ yrpc.PostNewPeerPlugin        = Ping(nil)
	_ yrpc.PostDialPlugin           = Ping(nil)
	_ yrpc.PostAcceptPlugin         = Ping(nil)
	_ yrpc.PostWriteCallPlugin      = Ping(nil)
	_ yrpc.PostWritePushPlugin      = Ping(nil)
	_ yrpc.PostReadCallHeaderPlugin = Ping(nil)
	_ yrpc.PostReadPushHeaderPlugin = Ping(nil)
)

// SetRate sets heartbeat rate.
func (h *heartPing) SetRate(rateSecond int) {
	if rateSecond < minRateSecond {
		rateSecond = minRateSecond
	}
	h.mu.Lock()
	h.pingRate = time.Second * time.Duration(rateSecond)
	h.pingRateSecond = strconv.Itoa(rateSecond)
	h.mu.Unlock()
	yrpc.Infof("set heartbeat rate: %ds", rateSecond)
}

func (h *heartPing) getRate() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.pingRate
}

func (h *heartPing) getPingRateSecond() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.pingRateSecond
}

// UseCall uses CALL method to ping.
func (h *heartPing) UseCall() {
	h.mu.Lock()
	h.useCall = true
	h.mu.Unlock()
}

// UsePush uses PUSH method to ping.
func (h *heartPing) UsePush() {
	h.mu.Lock()
	h.useCall = false
	h.mu.Unlock()
}

func (h *heartPing) isCall() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.useCall
}

// Name returns name.
func (h *heartPing) Name() string {
	return "heart-ping"
}

// PostNewPeer runs ping worker.
func (h *heartPing) PostNewPeer(peer yrpc.EarlyPeer) error {
	rangeSession := peer.RangeSession
	go func() {
		var isCall bool
		for {
			time.Sleep(h.getRate())
			isCall = h.isCall()
			rangeSession(func(sess yrpc.Session) bool {
				if !sess.Health() {
					return true
				}
				info, ok := getHeartbeatInfo(sess.Swap())
				if !ok {
					return true
				}
				cp := info.elemCopy()
				if cp.last.Add(cp.rate).After(coarsetime.CeilingTimeNow()) {
					return true
				}
				if isCall {
					h.goCall(sess)
				} else {
					h.goPush(sess)
				}
				return true
			})
		}
	}()
	return nil
}

// PostDial initializes heartbeat information.
func (h *heartPing) PostDial(sess yrpc.PreSession, _ bool) *yrpc.Status {
	return h.PostAccept(sess)
}

// PostAccept initializes heartbeat information.
func (h *heartPing) PostAccept(sess yrpc.PreSession) *yrpc.Status {
	rate := h.getRate()
	initHeartbeatInfo(sess.Swap(), rate)
	return nil
}

// PostWriteCall updates heartbeat information.
func (h *heartPing) PostWriteCall(ctx yrpc.WriteCtx) *yrpc.Status {
	return h.PostWritePush(ctx)
}

// PostWritePush updates heartbeat information.
func (h *heartPing) PostWritePush(ctx yrpc.WriteCtx) *yrpc.Status {
	h.update(ctx)
	return nil
}

// PostReadCallHeader updates heartbeat information.
func (h *heartPing) PostReadCallHeader(ctx yrpc.ReadCtx) *yrpc.Status {
	return h.PostReadPushHeader(ctx)
}

// PostReadPushHeader updates heartbeat information.
func (h *heartPing) PostReadPushHeader(ctx yrpc.ReadCtx) *yrpc.Status {
	h.update(ctx)
	return nil
}

func (h *heartPing) goCall(sess yrpc.Session) {
	yrpc.Go(func() {
		if sess.Call(
			HeartbeatServiceMethod, nil, nil,
			yrpc.WithSetMeta(heartbeatMetaKey, h.getPingRateSecond()),
		).Status() != nil {
			sess.Close()
		}
	})
}

func (h *heartPing) goPush(sess yrpc.Session) {
	yrpc.Go(func() {
		if sess.Push(
			HeartbeatServiceMethod,
			nil,
			yrpc.WithSetMeta(heartbeatMetaKey, h.getPingRateSecond()),
		) != nil {
			sess.Close()
		}
	})
}

func (h *heartPing) update(ctx yrpc.PreCtx) {
	sess := ctx.Session()
	if !sess.Health() {
		return
	}
	updateHeartbeatInfo(sess.Swap(), h.getRate())
}
