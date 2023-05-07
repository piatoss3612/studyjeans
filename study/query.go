package study

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Query interface {
	FindManagement(ctx context.Context, guildID string) (*Management, error)
	FindStudy(ctx context.Context, id string) (*Study, error)
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
