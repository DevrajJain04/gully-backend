package services

import (
	"context"
	"math/rand"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
	"gully-backend/repositories"
)

const joinCodeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type GroupService struct {
	groupRepo repositories.GroupRepository
}

func NewGroupService(groupRepo repositories.GroupRepository) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

func (s *GroupService) CreateGroup(ctx context.Context, name string, createdBy primitive.ObjectID) (*models.Group, error) {
	group := &models.Group{
		Name:      name,
		JoinCode:  generateJoinCode(6),
		CreatedBy: createdBy,
		Members:   []primitive.ObjectID{createdBy},
	}
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) JoinGroup(ctx context.Context, code string, userID primitive.ObjectID) (*models.Group, error) {
	group, err := s.groupRepo.FindByJoinCode(ctx, code)
	if err != nil {
		return nil, err
	}
	// Add user to members (idempotent)
	if err := s.groupRepo.AddMember(ctx, group.ID, userID); err != nil {
		return nil, err
	}
	// Re-fetch to get updated members list
	return s.groupRepo.FindByID(ctx, group.ID)
}

func (s *GroupService) GetGroup(ctx context.Context, id primitive.ObjectID) (*models.Group, error) {
	return s.groupRepo.FindByID(ctx, id)
}

func (s *GroupService) GetUserGroups(ctx context.Context, userID primitive.ObjectID) ([]models.Group, error) {
	return s.groupRepo.FindByMember(ctx, userID)
}

func generateJoinCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = joinCodeChars[rand.Intn(len(joinCodeChars))]
	}
	return string(b)
}
