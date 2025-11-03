package middleware

type ctxKey string

const (
	RealIpKey ctxKey = "realIP"
	RoleKey   ctxKey = "role"
	UserIdKey ctxKey = "userID"
)
