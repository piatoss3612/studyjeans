package study

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Store interface {
	StoreManagement(ctx context.Context, m Management) (string, error)
	StoreStudy(ctx context.Context, s Study) (string, error)
	UpdateManagement(ctx context.Context, m Management) error
	UpdateStudy(ctx context.Context, s Study) error
}

type StoreImpl struct {
	client *mongo.Client
	dbname string
}

func NewStore(client *mongo.Client, dbname string) Store {
	return &StoreImpl{client: client, dbname: dbname}
}

func (si *StoreImpl) StoreManagement(ctx context.Context, m Management) (string, error) {
	collection := si.client.Database(si.dbname).Collection("management")

	res, err := collection.InsertOne(ctx, m)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (si *StoreImpl) StoreStudy(ctx context.Context, s Study) (string, error) {
	collection := si.client.Database(si.dbname).Collection("study")

	res, err := collection.InsertOne(ctx, s)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (si *StoreImpl) UpdateManagement(ctx context.Context, m Management) error {
	collection := si.client.Database(si.dbname).Collection("management")

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
	collection := si.client.Database(si.dbname).Collection("study")

	objID, err := primitive.ObjectIDFromHex(s.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}

	update := bson.D{
		{
			Key: "$set", Value: bson.D{
				{Key: "title", Value: s.Title},
				{Key: "content_url", Value: s.ContentURL},
				{Key: "members", Value: s.Members},
				{Key: "updated_at", Value: s.UpdatedAt},
			},
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}
