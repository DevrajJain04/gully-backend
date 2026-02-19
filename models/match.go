package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MatchStatusLive     = "live"
	MatchStatusFinished = "finished"
)

type Match struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID      primitive.ObjectID `bson:"group_id"      json:"group_id"`
	Player1ID    primitive.ObjectID `bson:"player1_id"    json:"player1_id"`
	Player2ID    primitive.ObjectID `bson:"player2_id"    json:"player2_id"`
	Player1Name  string             `bson:"player1_name"  json:"player1_name"`
	Player2Name  string             `bson:"player2_name"  json:"player2_name"`
	Score1       int                `bson:"score1"        json:"score1"`
	Score2       int                `bson:"score2"        json:"score2"`
	ScoreHistory []string           `bson:"score_history" json:"score_history"`
	Status       string             `bson:"status"        json:"status"`
	StartedAt    time.Time          `bson:"started_at"    json:"started_at"`
	FinishedAt   *time.Time         `bson:"finished_at"   json:"finished_at"`
	DurationSecs int                `bson:"duration_secs" json:"duration_secs"`
	CreatedAt    time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"    json:"updated_at"`
}
