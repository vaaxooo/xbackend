package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
)

// UseCaseHandler defines a transport-agnostic contract for invoking application
// services. It allows decorators (metrics, logging, unit-of-work, etc.) to be
// composed transparently around a use-case.
type UseCaseHandler[Cmd any, Resp any] interface {
	Handle(ctx context.Context, cmd Cmd) (Resp, error)
}

// UseCaseFunc adapts a plain function to the UseCaseHandler interface.
type UseCaseFunc[Cmd any, Resp any] func(ctx context.Context, cmd Cmd) (Resp, error)

func (f UseCaseFunc[Cmd, Resp]) Handle(ctx context.Context, cmd Cmd) (Resp, error) {
	return f(ctx, cmd)
}

// UseCaseMiddleware enriches the request context with deadlines and metadata
// before invoking the application layer.
type UseCaseMiddleware struct {
	Timeout time.Duration
}

func HandleUseCase[Cmd any, Resp any](m UseCaseMiddleware, r *http.Request, handler UseCaseHandler[Cmd, Resp], cmd Cmd) (Resp, error) {
	ctx, cancel := m.contextWithTimeout(r)
	if cancel != nil {
		defer cancel()
	}

	ctx = common.WithRequestMeta(ctx, common.RequestMeta{
		UserAgent: r.UserAgent(),
		IP:        clientIP(r),
	})

	return handler.Handle(ctx, cmd)
}

func (m UseCaseMiddleware) contextWithTimeout(r *http.Request) (context.Context, context.CancelFunc) {
	ctx := r.Context()

	if m.Timeout > 0 {
		return context.WithTimeout(ctx, m.Timeout)
	}
	if deadline, ok := ctx.Deadline(); ok {
		return context.WithDeadline(ctx, deadline)
	}

	return ctx, nil
}

func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
