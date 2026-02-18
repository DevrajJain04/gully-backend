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
	groupRepo *repositories.GroupRepo
}

func NewGroupService(groupRepo *repositories.GroupRepo) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

func (s *GroupService) CreateGroup(ctx context.Context, name string, createdBy primitive.ObjectID) (*models.Group, error) {
	group := &models.Group{
		Name:      name,
		JoinCode:  generateJoinCode(6),
		CreatedBy: createdBy,
	}
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) JoinGroup(ctx context.Context, code string) (*models.Group, error) {
	return s.groupRepo.FindByJoinCode(ctx, code)
}

func (s *GroupService) GetGroup(ctx context.Context, id primitive.ObjectID) (*models.Group, error) {
	return s.groupRepo.FindByID(ctx, id)
}

func generateJoinCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = joinCodeChars[rand.Intn(len(joinCodeChars))]
	}
	return string(b)
}
