package cctx

import "context"

type ContextKey string

const (
	RealIpKey ContextKey = "realIP"
)

func RealIP(ctx context.Context) string {
	if val := ctx.Value(RealIpKey); val != nil {
		return val.(string)
	}
	return ""
}
