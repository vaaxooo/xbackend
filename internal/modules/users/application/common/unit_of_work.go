package common

import "context"

// UnitOfWork defines a transactional boundary for application usecases.
// Infrastructure (e.g. postgres) should provide an implementation.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
