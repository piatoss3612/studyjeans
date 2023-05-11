package study

import (
	"context"
	"time"

	models "github.com/piatoss3612/presentation-helper-bot/internal/models/study"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Store interface {
	CreateStudy(ctx context.Context, s models.Study) (*models.Study, error)
	CreateRound(ctx context.Context, r models.Round) (*models.Round, error)
	UpdateStudy(ctx context.Context, s models.Study) (*models.Study, error)
	UpdateRound(ctx context.Context, r models.Round) (*models.Round, error)
}

type StoreImpl struct {
	client *mongo.Client
	dbname string
}

func NewStore(client *mongo.Client, dbname string) Store {
	return &StoreImpl{client: client, dbname: dbname}
}

func (si *StoreImpl) CreateStudy(ctx context.Context, s models.Study) (*models.Study, error) {
	collection := si.client.Database(si.dbname).Collection("study")

	res, err := collection.InsertOne(ctx, s)
	if err != nil {
		return nil, err
	}

	s.SetID(res.InsertedID.(primitive.ObjectID).Hex())
	return &s, nil
}

func (si *StoreImpl) CreateRound(ctx context.Context, r models.Round) (*models.Round, error) {
	collection := si.client.Database(si.dbname).Collection("round")

	res, err := collection.InsertOne(ctx, r)
	if err != nil {
		return nil, err
	}

	r.SetID(res.InsertedID.(primitive.ObjectID).Hex())

	return &r, nil
}

func (si *StoreImpl) UpdateStudy(ctx context.Context, s models.Study) (*models.Study, error) {
	collection := si.client.Database(si.dbname).Collection("study")

	objID, err := primitive.ObjectIDFromHex(s.ID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}

	s.SetUpdatedAt(time.Now())

	update := bson.D{
		{
			Key: "$set", Value: bson.D{
				{Key: "guild_id", Value: s.GuildID},
				{Key: "notice_channel_id", Value: s.NoticeChannelID},
				{Key: "manager_id", Value: s.ManagerID},
				{Key: "ongoing_round_id", Value: s.OngoingRoundID},
				{Key: "current_stage", Value: s.CurrentStage},
				{Key: "total_round", Value: s.TotalRound},
				{Key: "updated_at", Value: s.UpdatedAt},
			},
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return &s, err
}

func (si StoreImpl) UpdateRound(ctx context.Context, r models.Round) (*models.Round, error) {
	collection := si.client.Database(si.dbname).Collection("round")

	objID, err := primitive.ObjectIDFromHex(r.ID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}

	r.SetUpdatedAt(time.Now())

	update := bson.D{
		{
			Key: "$set", Value: bson.D{
				{Key: "number", Value: r.Number},
				{Key: "title", Value: r.Title},
				{Key: "content_url", Value: r.ContentURL},
				{Key: "stage", Value: r.Stage},
				{Key: "members", Value: r.Members},
				{Key: "updated_at", Value: r.UpdatedAt},
			},
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return &r, err
}
