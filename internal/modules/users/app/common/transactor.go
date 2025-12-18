package common

import "context"

// Transactor defines a unit-of-work boundary for application usecases.
// Infrastructure (e.g. postgres) should provide an implementation.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
