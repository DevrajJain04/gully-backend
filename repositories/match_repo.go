package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gully-backend/models"
)

type MatchRepo struct {
	col *mongo.Collection
}

func NewMatchRepo(db *mongo.Database) *MatchRepo {
	return &MatchRepo{col: db.Collection("matches")}
}

func (r *MatchRepo) Create(ctx context.Context, match *models.Match) error {
	match.ID = primitive.NewObjectID()
	match.CreatedAt = time.Now()
	match.UpdatedAt = match.CreatedAt
	_, err := r.col.InsertOne(ctx, match)
	return err
}

func (r *MatchRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Match, error) {
	var match models.Match
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&match)
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MatchRepo) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Match, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{"group_id": groupID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matches []models.Match
	if err := cursor.All(ctx, &matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *MatchRepo) Update(ctx context.Context, match *models.Match) error {
	match.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": match.ID}, match)
	return err
}

func (r *MatchRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
