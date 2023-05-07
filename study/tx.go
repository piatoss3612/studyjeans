package study

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Tx interface {
	Store
	Query
	ExecTx(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
}

type TxImpl struct {
	Query
	Store
	client *mongo.Client
}

func NewTx(client *mongo.Client, dbname string) Tx {
	return &TxImpl{
		Query:  NewQuery(client, dbname),
		Store:  NewStore(client, dbname),
		client: client,
	}
}

func (tx *TxImpl) ExecTx(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	sess, err := tx.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer sess.EndSession(ctx)

	timeout := time.Duration(5 * time.Second)

	opts := options.Transaction().SetMaxCommitTime(&timeout) // TODO: set read concern, write concern

	return sess.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return fn(sc)
	}, opts)
}
