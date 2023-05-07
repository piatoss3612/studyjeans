package study

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store interface {
	StoreManagement(ctx context.Context, m Management) (string, error)
	StoreStudy(ctx context.Context, s Study) (string, error)
	UpdateManagement(ctx context.Context, m Management) error
	UpdateStudy(ctx context.Context, s Study) error
}

type Query interface {
	FindManagement(ctx context.Context, guildID string) (*Management, error)
	FindStudy(ctx context.Context, id string) (*Study, error)
}

type Tx interface {
	Store
	Query
	ExecTx(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
}

type StoreImpl struct {
	client *mongo.Client
}

func NewStore(client *mongo.Client) Store {
	return &StoreImpl{client: client}
}

func (si *StoreImpl) StoreManagement(ctx context.Context, m Management) (string, error) {
	collection := si.client.Database("study").Collection("management")

	res, err := collection.InsertOne(ctx, m)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (si *StoreImpl) StoreStudy(ctx context.Context, s Study) (string, error) {
	collection := si.client.Database("study").Collection("study")

	res, err := collection.InsertOne(ctx, s)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (si *StoreImpl) UpdateManagement(ctx context.Context, m Management) error {
	collection := si.client.Database("study").Collection("management")

	objID, err := primitive.ObjectIDFromHex(m.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}

	update := bson.D{
		{
			Key: "$set", Value: bson.D{
				{Key: "notice_channel_id", Value: m.NoticeChannelID},
				{Key: "manager_id", Value: m.ManagerID},
				{Key: "ongoing_study_id", Value: m.OngoingStudyID},
				{Key: "current_study_stage", Value: m.CurrentStudyStage},
				{Key: "updated_at", Value: m.UpdatedAt},
			},
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

func (si StoreImpl) UpdateStudy(ctx context.Context, s Study) error {
	collection := si.client.Database("study").Collection("study")

	objID, err := primitive.ObjectIDFromHex(s.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}

	update := bson.D{
		{
			Key: "$set", Value: bson.D{
				{Key: "title", Value: s.Title},
				{Key: "members", Value: s.Members},
				{Key: "updated_at", Value: s.UpdatedAt},
			},
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

type QueryImpl struct {
	client *mongo.Client
}

func NewQuery(client *mongo.Client) Query {
	return &QueryImpl{client: client}
}

func (q *QueryImpl) FindManagement(ctx context.Context, guildID string) (*Management, error) {
	collection := q.client.Database("study").Collection("management")

	filter := bson.M{"guild_id": guildID}

	m := NewManagement()

	err := collection.FindOne(ctx, filter).Decode(m)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return m, nil
}

func (q *QueryImpl) FindStudy(ctx context.Context, id string) (*Study, error) {
	collection := q.client.Database("study").Collection("study")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}

	s := New()

	err = collection.FindOne(ctx, filter).Decode(s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return s, nil
}

type TxImpl struct {
	Query
	Store
	client *mongo.Client
}

func NewTx(client *mongo.Client) Tx {
	return &TxImpl{
		Query:  NewQuery(client),
		Store:  NewStore(client),
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
