package services

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
	"gully-backend/repositories"
)

type PlayerService struct {
	playerRepo *repositories.PlayerRepo
}

func NewPlayerService(playerRepo *repositories.PlayerRepo) *PlayerService {
	return &PlayerService{playerRepo: playerRepo}
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

func (s *PlayerService) GetPlayers(ctx context.Context, groupID primitive.ObjectID) ([]models.Player, error) {
	return s.playerRepo.FindByGroupID(ctx, groupID)
}
