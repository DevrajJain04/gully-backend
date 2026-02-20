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

func TestCreatePlayer_Success(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	playerRepo.On("Create", ctx, mock.AnythingOfType("*models.Player")).Return(nil)

	player, err := svc.CreatePlayer(ctx, "Alice", groupID)

	assert.NoError(t, err)
	assert.NotNil(t, player)
	assert.Equal(t, "Alice", player.Name)
	assert.Equal(t, groupID, player.GroupID)
	playerRepo.AssertExpectations(t)
}

func TestCreatePlayer_RepoError(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	playerRepo.On("Create", ctx, mock.AnythingOfType("*models.Player")).Return(errors.New("db error"))

	player, err := svc.CreatePlayer(ctx, "Alice", primitive.NewObjectID())

	assert.Error(t, err)
	assert.Nil(t, player)
}

func TestCreatePlayerIfNotExists_AlreadyExists(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	existing := &models.Player{
		ID:      primitive.NewObjectID(),
		Name:    "Alice",
		GroupID: groupID,
	}
	playerRepo.On("FindByNameAndGroupID", ctx, "Alice", groupID).Return(existing, nil)

	player, err := svc.CreatePlayerIfNotExists(ctx, "Alice", groupID)

	assert.NoError(t, err)
	assert.Equal(t, existing.ID, player.ID)
	// Should NOT call Create
	playerRepo.AssertNotCalled(t, "Create")
}

func TestCreatePlayerIfNotExists_NewPlayer(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	playerRepo.On("FindByNameAndGroupID", ctx, "Bob", groupID).Return(nil, errors.New("not found"))
	playerRepo.On("Create", ctx, mock.AnythingOfType("*models.Player")).Return(nil)

	player, err := svc.CreatePlayerIfNotExists(ctx, "Bob", groupID)

	assert.NoError(t, err)
	assert.NotNil(t, player)
	assert.Equal(t, "Bob", player.Name)
	playerRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*models.Player"))
}

func TestGetPlayers_Success(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	players := []models.Player{
		{ID: primitive.NewObjectID(), Name: "Alice", GroupID: groupID},
		{ID: primitive.NewObjectID(), Name: "Bob", GroupID: groupID},
	}
	playerRepo.On("FindByGroupID", ctx, groupID).Return(players, nil)

	result, err := svc.GetPlayers(ctx, groupID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestDeletePlayer_Success(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	playerID := primitive.NewObjectID()
	playerRepo.On("Delete", ctx, playerID).Return(nil)

	err := svc.DeletePlayer(ctx, playerID)

	assert.NoError(t, err)
	playerRepo.AssertExpectations(t)
}

// ── MergePlayer tests ──

func TestMergePlayer_Success(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	targetID := primitive.NewObjectID()
	sourceID := primitive.NewObjectID()

	target := &models.Player{ID: targetID, Name: "Alice", GroupID: groupID}
	source := &models.Player{ID: sourceID, Name: "Alice2", GroupID: groupID}

	playerRepo.On("FindByID", ctx, targetID).Return(target, nil)
	playerRepo.On("FindByID", ctx, sourceID).Return(source, nil)
	matchRepo.On("ReplacePlayerInMatches", ctx, groupID, sourceID, targetID, "Alice2", "Alice").Return(nil)
	playerRepo.On("Delete", ctx, sourceID).Return(nil)

	err := svc.MergePlayer(ctx, groupID, targetID, sourceID)

	assert.NoError(t, err)
	playerRepo.AssertExpectations(t)
	matchRepo.AssertExpectations(t)
}

func TestMergePlayer_SamePlayer_Fails(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	playerID := primitive.NewObjectID()

	err := svc.MergePlayer(ctx, primitive.NewObjectID(), playerID, playerID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot merge player with itself")
}

func TestMergePlayer_TargetNotFound(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	targetID := primitive.NewObjectID()
	sourceID := primitive.NewObjectID()

	playerRepo.On("FindByID", ctx, targetID).Return(nil, errors.New("not found"))

	err := svc.MergePlayer(ctx, primitive.NewObjectID(), targetID, sourceID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target player not found")
}

func TestMergePlayer_SourceNotFound(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	targetID := primitive.NewObjectID()
	sourceID := primitive.NewObjectID()

	target := &models.Player{ID: targetID, Name: "Alice"}
	playerRepo.On("FindByID", ctx, targetID).Return(target, nil)
	playerRepo.On("FindByID", ctx, sourceID).Return(nil, errors.New("not found"))

	err := svc.MergePlayer(ctx, primitive.NewObjectID(), targetID, sourceID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source player not found")
}

func TestMergePlayer_ReplaceError(t *testing.T) {
	playerRepo := new(MockPlayerRepo)
	matchRepo := new(MockMatchRepo)
	svc := NewPlayerService(playerRepo, matchRepo)
	ctx := context.Background()

	groupID := primitive.NewObjectID()
	targetID := primitive.NewObjectID()
	sourceID := primitive.NewObjectID()

	target := &models.Player{ID: targetID, Name: "Alice"}
	source := &models.Player{ID: sourceID, Name: "Alice2"}

	playerRepo.On("FindByID", ctx, targetID).Return(target, nil)
	playerRepo.On("FindByID", ctx, sourceID).Return(source, nil)
	matchRepo.On("ReplacePlayerInMatches", ctx, groupID, sourceID, targetID, "Alice2", "Alice").Return(errors.New("db error"))

	err := svc.MergePlayer(ctx, groupID, targetID, sourceID)

	assert.Error(t, err)
	// Should NOT delete source if replace failed
	playerRepo.AssertNotCalled(t, "Delete")
}
