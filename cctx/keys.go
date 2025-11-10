package cctx

import "context"

type contextKey string

const (
	RealIpKey contextKey = "realIP"
)

func RealIP(ctx context.Context) string {
	if val := ctx.Value(RealIpKey); val != nil {
		return val.(string)
	}
	return ""
}
