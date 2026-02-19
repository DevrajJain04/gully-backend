package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name      string               `bson:"name"          json:"name"`
	JoinCode  string               `bson:"join_code"     json:"join_code"`
	CreatedBy primitive.ObjectID   `bson:"created_by"    json:"created_by"`
	Members   []primitive.ObjectID `bson:"members"       json:"members"`
	CreatedAt time.Time            `bson:"created_at"    json:"created_at"`
}
