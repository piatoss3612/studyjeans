package study

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Query interface {
	FindManagement(ctx context.Context, guildID string) (*Management, error)
	FindStudy(ctx context.Context, id string) (*Study, error)
	FindStudies(ctx context.Context, guildID string) ([]*Study, error)
}

type QueryImpl struct {
	client *mongo.Client
	dbname string
}

func NewQuery(client *mongo.Client, dbname string) Query {
	return &QueryImpl{client: client, dbname: dbname}
}

func (q *QueryImpl) FindManagement(ctx context.Context, guildID string) (*Management, error) {
	collection := q.client.Database(q.dbname).Collection("management")

	filter := bson.M{"guild_id": guildID}

	m := NewManagement()

	err := collection.FindOne(ctx, filter).Decode(&m)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &m, nil
}

func (q *QueryImpl) FindStudy(ctx context.Context, id string) (*Study, error) {
	collection := q.client.Database(q.dbname).Collection("study")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}

	s := New()

	err = collection.FindOne(ctx, filter).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (q *QueryImpl) FindStudies(ctx context.Context, guildID string) ([]*Study, error) {
	collection := q.client.Database(q.dbname).Collection("study")

	filter := bson.M{"guild_id": guildID}
	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	var studies []*Study

	for cursor.Next(ctx) {
		s := New()

		err := cursor.Decode(&s)
		if err != nil {
			return nil, err
		}

		studies = append(studies, &s)
	}

	return studies, nil
}
