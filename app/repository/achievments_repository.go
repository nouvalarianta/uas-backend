package repository

import (
	"context"
	"time"
	model "uas-backend/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AchievementRepository interface {
	Create(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*model.Achievement, error)
	GetByStudentID(ctx context.Context, studentID string, filter map[string]interface{}, skip, limit int64) ([]*model.Achievement, int64, error)
	GetAll(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]*model.Achievement, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, update *model.UpdateAchievementRequest) (*model.Achievement, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	AddAttachment(ctx context.Context, id primitive.ObjectID, attachment model.Attachment) error
}

type achievementRepository struct {
	collection *mongo.Collection
}

func NewAchievementRepository(db *mongo.Database) AchievementRepository {
	return &achievementRepository{
		collection: db.Collection("achievements"),
	}
}

func (r *achievementRepository) Create(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error) {
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, achievement)
	if err != nil {
		return nil, err
	}

	achievement.ID = result.InsertedID.(primitive.ObjectID)
	return achievement, nil
}

func (r *achievementRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Achievement, error) {
	var achievement model.Achievement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&achievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &achievement, nil
}

func (r *achievementRepository) GetByStudentID(ctx context.Context, studentID string, filter map[string]interface{}, skip, limit int64) ([]*model.Achievement, int64, error) {
	filter["studentId"] = studentID

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var achievements []*model.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, 0, err
	}

	return achievements, total, nil
}

func (r *achievementRepository) GetAll(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]*model.Achievement, int64, error) {
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var achievements []*model.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, 0, err
	}

	return achievements, total, nil
}

func (r *achievementRepository) Update(ctx context.Context, id primitive.ObjectID, update *model.UpdateAchievementRequest) (*model.Achievement, error) {
	updateDoc := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if update.Title != "" {
		updateDoc["$set"].(bson.M)["title"] = update.Title
	}
	if update.Description != "" {
		updateDoc["$set"].(bson.M)["description"] = update.Description
	}
	if update.Details != nil {
		updateDoc["$set"].(bson.M)["details"] = update.Details
	}
	if update.Points > 0 {
		updateDoc["$set"].(bson.M)["points"] = update.Points
	}
	if update.Tags != nil {
		updateDoc["$set"].(bson.M)["tags"] = update.Tags
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var achievement model.Achievement
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, updateDoc, opts).Decode(&achievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &achievement, nil
}

func (r *achievementRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *achievementRepository) AddAttachment(ctx context.Context, id primitive.ObjectID, attachment model.Attachment) error {
	attachment.UploadedAt = time.Now()

	update := bson.M{
		"$push": bson.M{"attachments": attachment},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
