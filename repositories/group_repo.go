package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"gully-backend/models"
)

type GroupRepo struct {
	col *mongo.Collection
}

func NewGroupRepo(db *mongo.Database) *GroupRepo {
	return &GroupRepo{col: db.Collection("groups")}
}

func (r *GroupRepo) Create(ctx context.Context, group *models.Group) error {
	group.ID = primitive.NewObjectID()
	group.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, group)
	return err
}

func (r *GroupRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Group, error) {
	var group models.Group
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepo) FindByJoinCode(ctx context.Context, code string) (*models.Group, error) {
	var group models.Group
	err := r.col.FindOne(ctx, bson.M{"join_code": code}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}
