package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
)

func TestCreateGroup_Success(t *testing.T) {
	groupRepo := new(MockGroupRepo)
	svc := NewGroupService(groupRepo)
	ctx := context.Background()
	userID := primitive.NewObjectID()

	groupRepo.On("Create", ctx, mock.AnythingOfType("*models.Group")).Return(nil)

	group, err := svc.CreateGroup(ctx, "Gully Boys", userID)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, "Gully Boys", group.Name)
	assert.Equal(t, userID, group.CreatedBy)
	assert.Contains(t, group.Members, userID)
	assert.Len(t, group.JoinCode, 6)
	groupRepo.AssertExpectations(t)
}

func TestCreateGroup_RepoError(t *testing.T) {
	groupRepo := new(MockGroupRepo)
	svc := NewGroupService(groupRepo)
	ctx := context.Background()
	userID := primitive.NewObjectID()

	groupRepo.On("Create", ctx, mock.AnythingOfType("*models.Group")).Return(errors.New("db error"))

	group, err := svc.CreateGroup(ctx, "Test", userID)

	assert.Error(t, err)
	assert.Nil(t, group)
	groupRepo.AssertExpectations(t)
}

func TestJoinGroup_Success(t *testing.T) {
	groupRepo := new(MockGroupRepo)
	svc := NewGroupService(groupRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	existing := &models.Group{
		ID:       groupID,
		Name:     "Test Group",
		JoinCode: "ABC123",
		Members:  []primitive.ObjectID{primitive.NewObjectID()},
	}
	updated := &models.Group{
		ID:       groupID,
		Name:     "Test Group",
		JoinCode: "ABC123",
		Members:  []primitive.ObjectID{existing.Members[0], userID},
	}

	groupRepo.On("FindByJoinCode", ctx, "ABC123").Return(existing, nil)
	groupRepo.On("AddMember", ctx, groupID, userID).Return(nil)
	groupRepo.On("FindByID", ctx, groupID).Return(updated, nil)

	result, err := svc.JoinGroup(ctx, "ABC123", userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Members, 2)
	groupRepo.AssertExpectations(t)
}

func TestJoinGroup_InvalidCode(t *testing.T) {
	groupRepo := new(MockGroupRepo)
	svc := NewGroupService(groupRepo)
	ctx := context.Background()

	groupRepo.On("FindByJoinCode", ctx, "BADCODE").Return(nil, errors.New("not found"))

	result, err := svc.JoinGroup(ctx, "BADCODE", primitive.NewObjectID())

	assert.Error(t, err)
	assert.Nil(t, result)
	groupRepo.AssertExpectations(t)
}

func TestGetGroup_Success(t *testing.T) {
	groupRepo := new(MockGroupRepo)
	svc := NewGroupService(groupRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	expected := &models.Group{ID: groupID, Name: "Test"}
	groupRepo.On("FindByID", ctx, groupID).Return(expected, nil)

	result, err := svc.GetGroup(ctx, groupID)

	assert.NoError(t, err)
	assert.Equal(t, "Test", result.Name)
	groupRepo.AssertExpectations(t)
}

func TestGetUserGroups_Success(t *testing.T) {
	groupRepo := new(MockGroupRepo)
	svc := NewGroupService(groupRepo)
	ctx := context.Background()

	userID := primitive.NewObjectID()
	groups := []models.Group{
		{ID: primitive.NewObjectID(), Name: "Group A"},
		{ID: primitive.NewObjectID(), Name: "Group B"},
	}
	groupRepo.On("FindByMember", ctx, userID).Return(groups, nil)

	result, err := svc.GetUserGroups(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	groupRepo.AssertExpectations(t)
}

func TestGenerateJoinCode_Length(t *testing.T) {
	code := generateJoinCode(6)
	assert.Len(t, code, 6)

	// All characters should be alphanumeric uppercase
	for _, ch := range code {
		assert.Contains(t, joinCodeChars, string(ch))
	}
}

func TestGenerateJoinCode_Uniqueness(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := generateJoinCode(6)
		codes[code] = true
	}
	// With 36^6 = ~2.2 billion possibilities, 100 codes should all be unique
	assert.Greater(t, len(codes), 90, "Most codes should be unique")
}
