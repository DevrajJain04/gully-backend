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

// ReplacePlayerInMatches replaces all occurrences of sourceID with targetID
// in team arrays and score history for matches in a group.
func (r *MatchRepo) ReplacePlayerInMatches(ctx context.Context, groupID, sourceID, targetID primitive.ObjectID, sourceName, targetName string) error {
	// Find all matches in this group that reference sourceID
	filter := bson.M{
		"group_id": groupID,
		"$or": bson.A{
			bson.M{"team1_ids": sourceID},
			bson.M{"team2_ids": sourceID},
		},
	}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var matches []models.Match
	if err := cursor.All(ctx, &matches); err != nil {
		return err
	}

	sourceHex := sourceID.Hex()
	targetHex := targetID.Hex()

	for _, m := range matches {
		// Replace in team IDs and names
		for i, id := range m.Team1IDs {
			if id == sourceID {
				m.Team1IDs[i] = targetID
				if i < len(m.Team1Names) {
					m.Team1Names[i] = targetName
				}
			}
		}
		for i, id := range m.Team2IDs {
			if id == sourceID {
				m.Team2IDs[i] = targetID
				if i < len(m.Team2Names) {
					m.Team2Names[i] = targetName
				}
			}
		}
		// Replace in score history
		for i, ev := range m.ScoreHistory {
			if ev.PlayerID == sourceHex {
				m.ScoreHistory[i].PlayerID = targetHex
			}
		}
		// Replace in positions
		for i, p := range m.Team1Positions {
			if p == sourceHex {
				m.Team1Positions[i] = targetHex
			}
		}
		for i, p := range m.Team2Positions {
			if p == sourceHex {
				m.Team2Positions[i] = targetHex
			}
		}
		// Replace serving player
		if m.ServingPlayerID == sourceHex {
			m.ServingPlayerID = targetHex
		}

		m.UpdatedAt = time.Now()
		if _, err := r.col.ReplaceOne(ctx, bson.M{"_id": m.ID}, m); err != nil {
			return err
		}
	}

	return nil
}
