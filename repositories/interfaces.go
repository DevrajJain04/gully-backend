package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
}

// GroupRepository defines the interface for group persistence.
type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Group, error)
	FindByJoinCode(ctx context.Context, code string) (*models.Group, error)
	FindByMember(ctx context.Context, userID primitive.ObjectID) ([]models.Group, error)
	AddMember(ctx context.Context, groupID, userID primitive.ObjectID) error
}

// PlayerRepository defines the interface for player persistence.
type PlayerRepository interface {
	Create(ctx context.Context, player *models.Player) error
	FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Player, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Player, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByNameAndGroupID(ctx context.Context, name string, groupID primitive.ObjectID) (*models.Player, error)
	Update(ctx context.Context, player *models.Player) error
}

// MatchRepository defines the interface for match persistence.
type MatchRepository interface {
	Create(ctx context.Context, match *models.Match) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Match, error)
	FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Match, error)
	Update(ctx context.Context, match *models.Match) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	ReplacePlayerInMatches(ctx context.Context, groupID, sourceID, targetID primitive.ObjectID, sourceName, targetName string) error
}
