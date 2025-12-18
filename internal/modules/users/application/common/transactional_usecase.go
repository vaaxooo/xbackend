package common

import "context"

// UseCase is the primary application abstraction executed by transports.
// It deliberately mirrors the Handle signature used by adapters so decorators
// can be shared across layers.
type UseCase[Cmd any, Resp any] interface {
	Execute(ctx context.Context, cmd Cmd) (Resp, error)
}

// UseCaseHandler adapts a UseCase to a Handle interface used by transports.
func UseCaseHandler[Cmd any, Resp any](uc UseCase[Cmd, Resp]) anyHandler[Cmd, Resp] {
	return anyHandler[Cmd, Resp]{next: uc}
}

// Handler mirrors the transport interface and allows decoration.
type Handler[Cmd any, Resp any] interface {
	Handle(ctx context.Context, cmd Cmd) (Resp, error)
}

type anyHandler[Cmd any, Resp any] struct {
	next UseCase[Cmd, Resp]
}

func (h anyHandler[Cmd, Resp]) Handle(ctx context.Context, cmd Cmd) (Resp, error) {
	return h.next.Execute(ctx, cmd)
}

type transactionalUseCase[Cmd any, Resp any] struct {
	uow  UnitOfWork
	next UseCase[Cmd, Resp]
}

// NewTransactionalUseCase wraps a use-case execution in a UnitOfWork boundary.
// It keeps the business logic agnostic of transaction management.
func NewTransactionalUseCase[Cmd any, Resp any](uow UnitOfWork, next UseCase[Cmd, Resp]) UseCase[Cmd, Resp] {
	return transactionalUseCase[Cmd, Resp]{
		uow:  uow,
		next: next,
	}
}

func (t transactionalUseCase[Cmd, Resp]) Execute(ctx context.Context, cmd Cmd) (Resp, error) {
	var (
		out Resp
		err error
	)

	if err = t.uow.Do(ctx, func(ctx context.Context) error {
		out, err = t.next.Execute(ctx, cmd)
		return err
	}); err != nil {
		return out, err
	}

	return out, nil
}
