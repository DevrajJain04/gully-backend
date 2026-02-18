package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Player struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name"          json:"name"`
	GroupID   primitive.ObjectID `bson:"group_id"      json:"group_id"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
}
