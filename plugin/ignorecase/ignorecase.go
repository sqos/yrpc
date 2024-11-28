// Package ignorecase dynamically ignoring the case of path
package ignorecase

import (
	"strings"

	"github.com/sqos/yrpc"
)

// NewIgnoreCase Returns a ignoreCase plugin.
func NewIgnoreCase() *ignoreCase {
	return &ignoreCase{}
}

type ignoreCase struct{}

var (
	_ yrpc.PostReadCallHeaderPlugin = new(ignoreCase)
	_ yrpc.PostReadPushHeaderPlugin = new(ignoreCase)
)

func (i *ignoreCase) Name() string {
	return "ignoreCase"
}

func (i *ignoreCase) PostReadCallHeader(ctx yrpc.ReadCtx) *yrpc.Status {
	// Dynamic transformation path is lowercase
	ctx.ResetServiceMethod(strings.ToLower(ctx.ServiceMethod()))
	return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx yrpc.ReadCtx) *yrpc.Status {
	// Dynamic transformation path is lowercase
	ctx.ResetServiceMethod(strings.ToLower(ctx.ServiceMethod()))
	return nil
}
