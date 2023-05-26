package mongo

import (
	"context"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/study/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TxOptsFn func(*mongoTx)

func WithDBName(dbname string) TxOptsFn {
	return func(tx *mongoTx) {
		tx.dbname = dbname
	}
}

type mongoTx struct {
	repository.Query
	repository.Store
	client *mongo.Client
	dbname string
}

func NewMongoTx(client *mongo.Client, opts ...TxOptsFn) repository.Tx {
	tx := &mongoTx{
		client: client,
		dbname: "default",
	}

	for _, opt := range opts {
		opt(tx)
	}

	tx.Query = NewMongoQuery(client, WithQueryDBName(tx.dbname))
	tx.Store = NewMongoStore(client, WithStoreDBName(tx.dbname))

	return tx
}

func (tx *mongoTx) ExecTx(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
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
