package ctx

type key string

const (
	userIDKey    key = "T-User-Id"
	requestIDKey key = "T-Request-Id"
	traceIDKey   key = "T-Trace-Id"
	spanIDKey    key = "T-Span-Id"
	clientIDKey  key = "T-Client-Id"
)

func GetSupportedCustomHeaders() []string {
	return []string{string(requestIDKey)}
}
