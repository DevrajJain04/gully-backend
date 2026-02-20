package services

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
	"gully-backend/repositories"
)

type PlayerService struct {
	playerRepo repositories.PlayerRepository
	matchRepo  repositories.MatchRepository
}

func NewPlayerService(playerRepo repositories.PlayerRepository, matchRepo repositories.MatchRepository) *PlayerService {
	return &PlayerService{playerRepo: playerRepo, matchRepo: matchRepo}
}

func (s *PlayerService) CreatePlayer(ctx context.Context, name string, groupID primitive.ObjectID) (*models.Player, error) {
	player := &models.Player{
		Name:    name,
		GroupID: groupID,
	}
	if err := s.playerRepo.Create(ctx, player); err != nil {
		return nil, err
	}
	return player, nil
}

// CreatePlayerIfNotExists creates a player only if one with the same name doesn't exist in the group.
func (s *PlayerService) CreatePlayerIfNotExists(ctx context.Context, name string, groupID primitive.ObjectID) (*models.Player, error) {
	existing, err := s.playerRepo.FindByNameAndGroupID(ctx, name, groupID)
	if err == nil && existing != nil {
		return existing, nil // already exists
	}
	return s.CreatePlayer(ctx, name, groupID)
}

func (s *PlayerService) GetPlayers(ctx context.Context, groupID primitive.ObjectID) ([]models.Player, error) {
	return s.playerRepo.FindByGroupID(ctx, groupID)
}

func (s *PlayerService) DeletePlayer(ctx context.Context, playerID primitive.ObjectID) error {
	return s.playerRepo.Delete(ctx, playerID)
}

// MergePlayer merges sourcePlayer into targetPlayer:
// - Replaces all references to sourcePlayer in match data with targetPlayer
// - Deletes the sourcePlayer record
func (s *PlayerService) MergePlayer(ctx context.Context, groupID, targetPlayerID, sourcePlayerID primitive.ObjectID) error {
	if targetPlayerID == sourcePlayerID {
		return errors.New("cannot merge player with itself")
	}

	target, err := s.playerRepo.FindByID(ctx, targetPlayerID)
	if err != nil {
		return errors.New("target player not found")
	}

	source, err := s.playerRepo.FindByID(ctx, sourcePlayerID)
	if err != nil {
		return errors.New("source player not found")
	}

	// Replace all match references
	if err := s.matchRepo.ReplacePlayerInMatches(ctx, groupID, sourcePlayerID, targetPlayerID, source.Name, target.Name); err != nil {
		return err
	}

	// Delete source player
	return s.playerRepo.Delete(ctx, sourcePlayerID)
}
