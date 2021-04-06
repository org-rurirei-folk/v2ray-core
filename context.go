// +build !confonly

package core

import (
	"context"
)

// V2rayKey is the key type of Instance in Context, exported for test.
type V2rayKey int

const v2rayKey V2rayKey = 1

// FromContext returns an Instance from the given context, or nil if the context doesn't contain one.
func FromContext(ctx context.Context) *Instance {
	if s, ok := ctx.Value(v2rayKey).(*Instance); ok {
		return s
	}
	return nil
}

// MustFromContext returns an Instance from the given context, or panics if not present.
func MustFromContext(ctx context.Context) *Instance {
	v := FromContext(ctx)
	if v == nil {
		panic("V is not in context.")
	}
	return v
}

// ToContext returns ctx from the given context, or creates an Instance if the context doesn't find that.
func ToContext(ctx context.Context, v *Instance) context.Context {
	if FromContext(ctx) != v {
		ctx = context.WithValue(ctx, v2rayKey, v)
	}
	return ctx
}

// MustToContext returns ctx from the given context, or panics if not found that.
func MustToContext(ctx context.Context, v *Instance) context.Context {
	if c := ToContext(ctx, v); c != ctx {
		panic("V is not in context.")
	}
	return ctx
}
