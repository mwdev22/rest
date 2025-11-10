package cctx

import "context"

type contextKey string

type ctxKey string

const (
	RealIpKey ctxKey = "realIP"
)

func RealIP(ctx context.Context) string {
	if val := ctx.Value(RealIpKey); val != nil {
		return val.(string)
	}
	return ""
}
