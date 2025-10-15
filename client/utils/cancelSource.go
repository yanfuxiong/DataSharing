package utils

import (
	"context"
	rtkCommon "rtk-cross-share/client/common"
	"sync"
)

type cancelSourceKey struct{}

type sourceHolder struct {
	mu     sync.Mutex
	src    rtkCommon.CancelBusinessSource
	parent *sourceHolder
}

func WithCancelSource(parent context.Context) (context.Context, func(rtkCommon.CancelBusinessSource)) {
	base, cancel := context.WithCancel(parent)

	var parentHolder *sourceHolder
	if v := parent.Value(cancelSourceKey{}); v != nil {
		if ph, ok := v.(*sourceHolder); ok && ph != nil {
			parentHolder = ph
		}
	}

	h := &sourceHolder{
		parent: parentHolder,
	}

	ctxWithVal := context.WithValue(base, cancelSourceKey{}, h)

	return ctxWithVal, func(src rtkCommon.CancelBusinessSource) {
		h.mu.Lock()
		if h.src == 0 {
			h.src = src
		}
		h.mu.Unlock()
		cancel()
	}
}

func GetCancelSource(ctx context.Context) (rtkCommon.CancelBusinessSource, bool) {
	v := ctx.Value(cancelSourceKey{})
	if v == nil {
		return 0, false
	}
	h, ok := v.(*sourceHolder)
	if !ok || h == nil {
		return 0, false
	}

	for cur := h; cur != nil; cur = cur.parent {
		cur.mu.Lock()
		src := cur.src
		cur.mu.Unlock()
		if src != 0 {
			return src, true
		}
	}
	return 0, false
}
