package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MatchStatusLive     = "live"
	MatchStatusFinished = "finished"
)

// ScoreEvent records who scored each point.
type ScoreEvent struct {
	Team     int    `bson:"team"      json:"team"`      // 1 or 2
	PlayerID string `bson:"player_id" json:"player_id"` // hex ObjectID of scorer
}

type Match struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID primitive.ObjectID `bson:"group_id"      json:"group_id"`

	// Teams — support 1v1, 1v2, 2v2
	Team1IDs   []primitive.ObjectID `bson:"team1_ids"   json:"team1_ids"`
	Team2IDs   []primitive.ObjectID `bson:"team2_ids"   json:"team2_ids"`
	Team1Names []string             `bson:"team1_names" json:"team1_names"`
	Team2Names []string             `bson:"team2_names" json:"team2_names"`

	// Scores
	Score1       int          `bson:"score1"        json:"score1"`
	Score2       int          `bson:"score2"        json:"score2"`
	ScoreHistory []ScoreEvent `bson:"score_history" json:"score_history"`

	// Serve tracking (visual only)
	ServingTeam     int    `bson:"serving_team"      json:"serving_team"`      // 1 or 2
	ServingPlayerID string `bson:"serving_player_id" json:"serving_player_id"` // hex ID

	// Court positions for doubles — player IDs in order [left, right]
	Team1Positions []string `bson:"team1_positions" json:"team1_positions"`
	Team2Positions []string `bson:"team2_positions" json:"team2_positions"`

	// Match lifecycle
	Status       string     `bson:"status"        json:"status"`
	StartedAt    time.Time  `bson:"started_at"    json:"started_at"`
	FinishedAt   *time.Time `bson:"finished_at"   json:"finished_at"`
	DurationSecs int        `bson:"duration_secs" json:"duration_secs"`
	CreatedAt    time.Time  `bson:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"    json:"updated_at"`
}
