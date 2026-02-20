package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"gully-backend/models"
)

type PlayerRepo struct {
	col *mongo.Collection
}

func NewPlayerRepo(db *mongo.Database) *PlayerRepo {
	return &PlayerRepo{col: db.Collection("players")}
}

func (r *PlayerRepo) Create(ctx context.Context, player *models.Player) error {
	player.ID = primitive.NewObjectID()
	player.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, player)
	return err
}

func (r *PlayerRepo) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Player, error) {
	cursor, err := r.col.Find(ctx, bson.M{"group_id": groupID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}
	return players, nil
}

func (r *PlayerRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Player, error) {
	var player models.Player
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&player)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *PlayerRepo) FindByNameAndGroupID(ctx context.Context, name string, groupID primitive.ObjectID) (*models.Player, error) {
	var player models.Player
	err := r.col.FindOne(ctx, bson.M{"name": name, "group_id": groupID}).Decode(&player)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepo) Update(ctx context.Context, player *models.Player) error {
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": player.ID}, player)
	return err
}
