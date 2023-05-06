package study

import (
	"context"
)

type Store interface {
	StoreManagement(ctx context.Context, m Management) (string, error)
	StoreStudy(ctx context.Context, s Study) (string, error)
	UpdateManagement(ctx context.Context, m Management) error
	UpdateStudy(ctx context.Context, s Study) error
}

type Query interface {
	FindManagement(ctx context.Context, id string) (Management, error)
	FindStudy(ctx context.Context, id string) (Study, error)
}

type Tx interface {
	Store
	Query
	ExecTx(ctx context.Context, fn func(Store) error) error
}
