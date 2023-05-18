package study

import (
	"context"

	models "github.com/piatoss3612/presentation-helper-bot/internal/models/study"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Query interface {
	FindStudy(ctx context.Context, guildID string) (*models.Study, error)
	FindRound(ctx context.Context, roundID string) (*models.Round, error)
	FindRounds(ctx context.Context, guildID string) ([]*models.Round, error)
}

type QueryOptsFn func(*mongoQuery)

func WithQueryDBName(dbname string) QueryOptsFn {
	return func(q *mongoQuery) {
		q.dbname = dbname
	}
}

type mongoQuery struct {
	client *mongo.Client
	dbname string
}

func NewMongoQuery(client *mongo.Client, opts ...QueryOptsFn) Query {
	q := &mongoQuery{client: client, dbname: "default"}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

func (q *mongoQuery) FindStudy(ctx context.Context, guildID string) (*models.Study, error) {
	collection := q.client.Database(q.dbname).Collection("study")

	filter := bson.M{"guild_id": guildID}

	s := models.New()

	err := collection.FindOne(ctx, filter).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (q *mongoQuery) FindRound(ctx context.Context, roundID string) (*models.Round, error) {
	collection := q.client.Database(q.dbname).Collection("round")

	objID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}

	r := models.NewRound()

	err = collection.FindOne(ctx, filter).Decode(&r)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &r, nil
}

func (q *mongoQuery) FindRounds(ctx context.Context, guildID string) ([]*models.Round, error) {
	collection := q.client.Database(q.dbname).Collection("round")

	filter := bson.M{"guild_id": guildID}
	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	var rounds []*models.Round

	for cursor.Next(ctx) {
		r := models.NewRound()

		err := cursor.Decode(&r)
		if err != nil {
			return nil, err
		}

		rounds = append(rounds, &r)
	}

	return rounds, nil
}
