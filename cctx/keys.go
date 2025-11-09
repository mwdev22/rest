package cctx

import "context"

type contextKey string

type ctxKey string

const (
	RealIpKey ctxKey = "realIP"
	RoleKey   ctxKey = "role"
	UserIdKey ctxKey = "userID"
)

func RealIP(ctx context.Context) string {
	if val := ctx.Value(RealIpKey); val != nil {
		return val.(string)
	}
	return ""
}

func Role(ctx context.Context) string {
	if val := ctx.Value(RoleKey); val != nil {
		return val.(string)
	}
	return ""
}

func UserID(ctx context.Context) string {
	if val := ctx.Value(UserIdKey); val != nil {
		return val.(string)
	}
	return ""
}
