package common

import "context"

type requestMetaKey struct{}

type RequestMeta struct {
	UserAgent string
	IP        string
}

func WithRequestMeta(ctx context.Context, meta RequestMeta) context.Context {
	return context.WithValue(ctx, requestMetaKey{}, meta)
}

func RequestMetaFromContext(ctx context.Context) (RequestMeta, bool) {
	v := ctx.Value(requestMetaKey{})
	if v == nil {
		return RequestMeta{}, false
	}
	m, ok := v.(RequestMeta)
	return m, ok
}
